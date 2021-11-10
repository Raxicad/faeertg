package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bandar-monitors/monitors/core/domain"
	"github.com/bandar-monitors/monitors/core/domain/webhook"
	"github.com/bandar-monitors/monitors/core/util"
	appCtx "github.com/bandar-monitors/monitors/manager/context"
	jsoniter "github.com/json-iterator/go"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
	"go.elastic.co/apm"
	"log"
	"os"
	"time"
)

func main() {
	amqpConn := os.Getenv("CONNECTION_STRINGS_RABBITMQ")
	psqlConn := "postgres://rixti:@localhost:5432/rixti?sslmode=disable"
	if amqpConn == "" {
		//panic("no amqp connection string provided")
		amqpConn = "amqp://guest:guest@localhost:5672/"
	}
	if psqlConn == "" {
		//panic("no psql connection string provided")
		psqlConn = "host=localhost port=15432 user=webhook-manager password=webhook-manager dbname=webhook-manager sslmode=disable"
	}

	runtimeCtx := context.Background()
	tx := apm.DefaultTracer.StartTransaction("startup", "balancer")

	tracingCtx := apm.ContextWithTransaction(runtimeCtx, tx)
	prepareSpan := tx.StartSpan("prepare_context", "balancer", nil)
	ctx := appCtx.NewContext(amqpConn, psqlConn, tracingCtx)
	prepareSpan.End()
	defer ctx.Dispose()

	//go schedulePublishersRefresh(ctx)
	go scheduleSettingsRefresh(ctx, runtimeCtx)
	go startConsumer(ctx, runtimeCtx)

	tx.End()
	util.WaitForShutdown()
}

//
//func schedulePublishersRefresh(ctx *ManagerContext) {
//	for true {
//		_ = RefreshPublishers(ctx)
//		time.Sleep(30 * time.Second)
//	}
//}

func scheduleSettingsRefresh(ctx *appCtx.ManagerContext, runtimeContext context.Context) {
	for true {
		_ = appCtx.RefreshSettings(ctx, runtimeContext)
		time.Sleep(30 * time.Second)
	}
}

func startConsumer(ctx *appCtx.ManagerContext, runtimeContext context.Context) {
	channel := ctx.Bus.Channel
	payloads, err := channel.Consume(
		ctx.Bus.Queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	util.FailOnError(err, "Failed to register consumer")

	forever := make(chan bool)

	go func() {
		for payload := range payloads {
			tx := apm.DefaultTracer.StartTransaction("notification_routing", "balancer")
			traceCtx := apm.ContextWithTransaction(runtimeContext, tx)

			prepareSpan, _ := apm.StartSpan(traceCtx, "prepare", "balancer")
			var notification domain.NotificationPayload
			err := jsoniter.Unmarshal(payload.Body, &notification)
			util.FailOnError(err, "Failed to parse notification payload")
			list, ok := (*ctx.SettingsCache)[notification.Slug]
			if !ok {
				apm.CaptureError(traceCtx, errors.New(fmt.Sprintf("Can't find config for %s", notification.Slug)))
				log.Printf("Can't find config for %s", notification.Slug)
				prepareSpan.End()
				tx.End()
				continue
			}

			publishPayload := webhook.PublishPayload{
				Payload:   notification.Payload,
				Timestamp: notification.Timestamp,
			}
			prepareSpan.End()
			tx.Context.SetLabel("listeners_count", len(list))
			tx.Context.SetLabel("monitor", notification.Slug)
			tx.Context.SetLabel("timestamp", notification.Timestamp)
			for _, webHookUrl := range list {
				publishSpan, _ := apm.StartSpan(traceCtx, "publish", "balancer")
				publishSpan.Context.SetLabel("webhook_url", webHookUrl)
				publishPayload.Subscriber = webHookUrl

				body, _ := jsoniter.Marshal(publishPayload)
				err = channel.Publish(
					webhook.PublishExchangeName,
					webhook.SenderRoutingKey,
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        body,
					},
				)

				if err != nil {
					apm.CaptureError(traceCtx, err)
					log.Printf("[ERR] Can't publish payload to the queue")
				} else {
					log.Printf("Published webhook from monitor '%s' url '%s'", notification.Slug, webHookUrl)
				}

				publishSpan.End()
			}

			_ = payload.Ack(false)
			tx.End()
		}
	}()

	log.Println("Waiting for notifications. To exit press CTRL+C")
	<-forever
}
