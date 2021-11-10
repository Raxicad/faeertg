package core

import (
	"context"
	"fmt"
	"github.com/bandar-monitors/monitors/core/domain"
	"github.com/bandar-monitors/monitors/core/domain/stats"
	"github.com/bandar-monitors/monitors/core/domain/webhook"
	"github.com/bandar-monitors/monitors/core/rmq"
	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	jsoniter "github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"go.elastic.co/apm"
	"time"
)

type publisher struct {
	conn        *rabbitmq.Connection
	channel     *rabbitmq.Channel
	slug        string
	amqpConnStr string
	//collector   MonitorStatsCollector
	//ctx         context.Context
}

type MonitorWatcherStatsConsumer = func(*stats.MonitorWatcherStats)

func (p *publisher) Publish(ctx context.Context, url string, status ProductStatus, webhookPayload WebhookPayload) error {
	if status == Unavailable {
		return nil
	}

	rootSpan, _ := apm.StartSpan(ctx, "publish", p.slug)
	rootSpan.Context.SetTag("status", fmt.Sprintf("%d", status))
	rootSpan.Context.SetTag("url", url)
	defer rootSpan.End()
	if p.conn.IsClosed() {
		err := connect(p.amqpConnStr, p, ctx)
		if err != nil {
			e := apm.CaptureError(ctx, err)
			e.SetSpan(rootSpan)
			e.Send()
			return err
		}
	}
	jsonSpan, _ := apm.StartSpan(ctx, "serialize_payload", p.slug)
	json := webhookPayload.ToJSON()
	jsonStr := string(json)
	jsonSpan.Context.SetLabel("json", jsonStr)
	payload := domain.NotificationPayload{
		Slug:      p.slug,
		Payload:   jsonStr,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}

	body, err := jsoniter.Marshal(payload)
	if err != nil {
		e := apm.CaptureError(ctx, err)
		e.SetSpan(jsonSpan)
		e.Send()
		return err
	}
	jsonSpan.End()
	sendSpan, _ := apm.StartSpan(ctx, "send", p.slug)
	err = p.channel.Publish(
		webhook.PublishExchangeName,
		webhook.BalancerRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	sendSpan.End()
	if err != nil {
		e := apm.CaptureError(ctx, err)
		e.SetSpan(jsonSpan)
		e.Send()
		return err
	}

	return err
}

func (p *publisher) Dispose() {
	_ = p.channel.Close()
	_ = p.conn.Close()
}

type ProductStatusChangePublisher interface {
	PublishStats(ctx context.Context, stats *stats.ComponentStats) error
	Publish(context context.Context, url string, status ProductStatus, webhookPayload WebhookPayload) error
	Dispose()
}
type StatsPublisher = func(context context.Context, stats *stats.ComponentStats) error

func (p *publisher) PublishStats(ctx context.Context, stats *stats.ComponentStats) error {
	json, err := jsoniter.Marshal(stats)
	if err != nil {
		return err
	}

	sendSpan, _ := apm.StartSpan(ctx, "send", p.slug)
	defer sendSpan.End()
	err = p.channel.Publish(
		webhook.ComponentExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        json,
		},
	)

	return err
}

func CreatePublisher(monitorSlug string, amqpConnStr string, ctx context.Context) (ProductStatusChangePublisher, error) {
	p := &publisher{
		slug:        monitorSlug,
		amqpConnStr: amqpConnStr,
	}

	err := connect(amqpConnStr, p, ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println("Publisher connected to rmq and started")

	return p, nil
}

func connect(amqpConnStr string, p *publisher, ctx context.Context) error {
	span, _ := apm.StartSpan(ctx, "rmq_connect", p.slug)
	defer span.End()
	conn, err := rabbitmq.Dial(amqpConnStr)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	p.conn = conn
	p.channel = ch

	_, err = rmq.ConfigureExchanges(ch)
	return err
}
