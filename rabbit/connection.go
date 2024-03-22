package rabbit

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Connection struct {
	connection  *amqp.Connection
	config      Config
	doneChan    chan struct{}
	reconnected chan struct{}
}

func NewConnection(config Config) *Connection {
	var connection = &Connection{
		connection:  nil,
		config:      config,
		doneChan:    make(chan struct{}),
		reconnected: make(chan struct{}),
	}

	return connection
}

func (conn *Connection) Connect() error {
	var (
		err error
	)

	if connection := conn.connection; connection != nil && !connection.IsClosed() {
		return errors.New("already connected")
	} else if connection != nil {
		_ = connection.Close()
	}

	if conn.connection, err = amqp.DialConfig(conn.config.createUrl(), conn.getAmqpConfig()); err != nil {
		return err
	}

	if conn.config.AutoReconnect {
		go conn.listenReconnect()
	}

	return nil
}

func (conn *Connection) Close() error {
	if conn.connection != nil && !conn.connection.IsClosed() {
		close(conn.doneChan)
		close(conn.reconnected)
		return conn.connection.Close()
	}
	return nil
}

func (conn *Connection) getAmqpConfig() amqp.Config {
	return amqp.Config{
		SASL: []amqp.Authentication{
			&amqp.AMQPlainAuth{
				Username: conn.config.Username,
				Password: conn.config.Password,
			},
		},
		Vhost:     conn.config.VirtualHost,
		Heartbeat: time.Second * 30,
	}
}

func (conn *Connection) GetUnderlyingConnection() *amqp.Connection {
	return conn.connection
}

func (conn *Connection) GetChannel() (*amqp.Channel, error) {
	return conn.connection.Channel()
}

func (conn *Connection) NotifyReconnect() <-chan struct{} {
	return conn.reconnected
}

func (conn *Connection) reconnect() error {
	if conn.connection != nil && !conn.connection.IsClosed() {
		if err := conn.connection.Close(); err != nil {
			return err
		}
	}

	return conn.Connect()
}

func (conn *Connection) listenReconnect() {
	select {
	case <-conn.connection.NotifyClose(make(chan *amqp.Error)):
		{
			for conn.reconnect() != nil {
				time.Sleep(conn.config.AutoReconnectInterval)
			}
			conn.reconnected <- struct{}{}
			break
		}
	case <-conn.doneChan:
		{
			break
		}
	}
}
