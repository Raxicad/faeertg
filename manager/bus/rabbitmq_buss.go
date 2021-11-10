package bus

import (
	"github.com/bandar-monitors/monitors/core/rmq"
	"github.com/bandar-monitors/monitors/core/util"
	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type DiscordHooksConsumerContext struct {
	connection *rabbitmq.Connection
	Channel    *rabbitmq.Channel
	Queue      *amqp.Queue
}

func (ctx *DiscordHooksConsumerContext) Close() {
	_ = ctx.connection.Close()
	_ = ctx.Channel.Close()
}

func EstablishAmqpConnection(amqpConnStr string) *DiscordHooksConsumerContext {
	//var conn *amqp.Connection
	//var err error
	conn, err := rabbitmq.Dial(amqpConnStr)
	if err != nil {
		delay := time.Second * 2
		log.Fatalf("Failed to connect to rabbit mq: %v. Retry in %v", err, delay)
		//time.Sleep(delay)
	}

	ch, err := conn.Channel()
	util.FailOnError(err, "Failed to open a channel")
	//
	//queue, err := ch.QueueDeclare(
	//	webhook.PublishQueueName,
	//	true,
	//	false,
	//	false,
	//	false,
	//	nil,
	//)
	//util.FailOnError(err, "Can't declare a queue")
	//
	//err = ch.ExchangeDeclare(
	//	webhook.PublishExchangeName,
	//	"direct",
	//	true,
	//	false,
	//	false,
	//	false,
	//	nil,
	//)
	//
	//util.FailOnError(err, "Failed to declare an exchange")
	//
	//err = ch.QueueBind(
	//	queue.Name,
	//	webhook.BalancerRoutingKey,
	//	webhook.PublishExchangeName,
	//	false,
	//	nil,
	//)

	cfg, err := rmq.ConfigureExchanges(ch)
	util.FailOnError(err, "failed to configure rmq")

	ctx := DiscordHooksConsumerContext{
		connection: conn,
		Channel:    ch,
		Queue:      cfg.BalancerQueue,
	}

	return &ctx
}
