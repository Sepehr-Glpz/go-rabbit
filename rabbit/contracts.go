package rabbit

import amqp "github.com/rabbitmq/amqp091-go"

type (
	IRabbitConnection interface {
		Connect() error
		Close() error
		GetUnderlyingConnection() *amqp.Connection
		GetChannel() (*amqp.Channel, error)
	}
	IMessageSerializer interface {
		Serialize(any) ([]byte, error)
		Deserialize([]byte, any) error
	}
	IPublisher interface {
		Publish(exchange string, key string, message any) error
	}
	IConsumer interface {
		Consume(queueName string, name string, exclusive bool, props amqp.Table, handler ConsumerFunc) error
		CancelConsumer(string) error
		Close() error
	}
	IRabbitClient interface {
		Connect() error
		Close() error
		Publisher() IPublisher
		Consumer() IConsumer
	}
	IConsumeActions interface {
		GetMessageAs(any) error
		GetRawBody() []byte
		GetHeaders() amqp.Table
	}
)
