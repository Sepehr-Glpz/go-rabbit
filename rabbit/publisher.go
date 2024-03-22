package rabbit

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func NewPublisher(connection IRabbitConnection, serializer IMessageSerializer) *Publisher {
	return &Publisher{
		serializer: serializer,
		conn:       connection,
	}
}

type Publisher struct {
	serializer IMessageSerializer
	conn       IRabbitConnection
}

func (pub *Publisher) Publish(exchange string, routingKey string, message any) error {
	var (
		channel *amqp.Channel
		err     error
		body    []byte = nil
	)

	if body, err = pub.serializer.Serialize(message); err != nil {
		return err
	}

	if channel, err = pub.conn.GetChannel(); err != nil {
		return err
	}

	defer func() {
		_ = channel.Close()
	}()

	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*3))

	defer cancel()

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         body,
	}

	return channel.PublishWithContext(ctx, exchange, routingKey, false, false, msg)
}
