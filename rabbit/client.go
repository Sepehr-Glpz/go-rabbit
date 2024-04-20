package rabbit

import "errors"

func NewClient(conn IRabbitConnection, publisher IPublisher, consumer IConsumer, topology ITopology) *Client {
	return &Client{
		conn:      conn,
		publisher: publisher,
		consumer:  consumer,
		topology:  topology,
	}
}

type Client struct {
	conn      IRabbitConnection
	publisher IPublisher
	consumer  IConsumer
	topology  ITopology
}

func (client *Client) Topology() ITopology {
	return client.topology
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
