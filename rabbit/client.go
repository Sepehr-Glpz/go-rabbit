package rabbit

import "errors"

func NewClient(conn IRabbitConnection, publisher IPublisher, consumer IConsumer) *Client {
	return &Client{
		conn:      conn,
		publisher: publisher,
		consumer:  consumer,
	}
}

type Client struct {
	conn      IRabbitConnection
	publisher IPublisher
	consumer  IConsumer
}

func (client *Client) Connect() error {
	return client.conn.Connect()
}

func (client *Client) Close() error {
	return errors.Join(client.conn.Close(), client.consumer.Close())
}

func (client *Client) Publisher() IPublisher {
	return client.publisher
}

func (client *Client) Consumer() IConsumer {
	return client.consumer
}
