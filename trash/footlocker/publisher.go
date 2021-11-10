package footlocker

import (
	"github.com/bandar-monitors/monitors/sites/domain"
	"github.com/json-iterator/go"
	"github.com/streadway/amqp"
)

type Publisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	ctx        *ProcessingContext
	queue      *amqp.Queue
}

func CreateQueuePublisher(ctx *ProcessingContext) (*Publisher, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		ctx.cfg.GetPublishQueueName(),
		true,
		false,
		false,
		false,
		nil)

	if err != nil {
		return nil, err
	}

	publisherContext := Publisher{
		connection: conn,
		channel:    channel,
		ctx:        ctx,
		queue:      &queue,
	}

	return &publisherContext, nil
}

func (publisher *Publisher) PublishProduct(p *domain.Product) error {
	json, err := jsoniter.Marshal(p)
	if err != nil {
		return err
	}

	err = publisher.channel.Publish(
		publisher.ctx.cfg.GetPublishExchange(),
		publisher.queue.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         json,
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (publisher *Publisher) Release() {
	publisher.connection.Close()
	publisher.channel.Close()
}
