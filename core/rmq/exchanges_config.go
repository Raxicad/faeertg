package rmq

import (
	"github.com/bandar-monitors/monitors/core/domain/webhook"
	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"
)

const senderName = "DefaultSender"

type ConfigResult struct {
	SenderQueue   *amqp.Queue
	BalancerQueue *amqp.Queue
}

func ConfigureExchanges(ch *rabbitmq.Channel) (*ConfigResult, error) {
	err := ch.ExchangeDeclare(
		webhook.PublishExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		webhook.ComponentExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	senderQueue, err := ch.QueueDeclare(
		senderName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		senderQueue.Name,
		webhook.SenderRoutingKey,
		webhook.PublishExchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	balancerQueue, err := ch.QueueDeclare(
		webhook.PublishQueueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		balancerQueue.Name,
		webhook.BalancerRoutingKey,
		webhook.PublishExchangeName,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	cfg := &ConfigResult{
		SenderQueue:   &senderQueue,
		BalancerQueue: &balancerQueue,
	}

	return cfg, nil
}
