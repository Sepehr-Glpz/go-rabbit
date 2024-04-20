package rabbit

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"slices"
)

type (
	Topology struct {
		conn   IRabbitConnection
		config *Mapping
	}
)

func NewTopology(conn IRabbitConnection, config Config) *Topology {
	return &Topology{
		conn:   conn,
		config: &config.Mapping,
	}
}

func (top *Topology) CommitAll() error {
	return errors.Join(
		top.CommitExchanges(),
		top.CommitQueues(),
		top.CommitBindings(),
	)
}

func (top *Topology) CommitBindings() error {
	var (
		channel *amqp.Channel
		err     error
		targets []BindingDef
	)

	if targets = top.config.Bindings; len(targets) == 0 {
		return nil
	}

	if channel, err = top.conn.GetChannel(); err != nil {
		return err
	}

	fails := make([]error, len(top.config.Bindings))
	for _, bind := range top.config.Bindings {
		ex := findExById(bind.FromExchangeId, top.config.Exchanges)
		queue := findQueueById(bind.ToQueueId, top.config.Queues)
		for _, key := range bind.Keys {
			if err = channel.QueueBind(queue.Name, key, ex.Name, false, bind.Arguments); err != nil {
				fails = append(fails, err)
			}
		}
	}

	return errors.Join(fails...)
}

func findExById(id int, exs []ExchangeDef) ExchangeDef {
	index := slices.IndexFunc(exs, func(def ExchangeDef) bool {
		return def.Id == id
	})

	return exs[index]
}

func findQueueById(id int, queues []QueueDef) QueueDef {
	index := slices.IndexFunc(queues, func(def QueueDef) bool {
		return def.Id == id
	})

	return queues[index]
}

func (top *Topology) CommitQueues() error {
	var (
		channel *amqp.Channel
		err     error
		targets []QueueDef
	)

	if targets = top.config.Queues; len(targets) == 0 {
		return nil
	}

	if channel, err = top.conn.GetChannel(); err != nil {
		return err
	}

	fails := make([]error, len(top.config.Queues))
	for _, queue := range top.config.Queues {
		if _, err = channel.QueueDeclare(queue.Name, queue.Durable, queue.AutoDelete, queue.Exclusive, false, queue.Arguments); err != nil {
			fails = append(fails, err)
		}
	}

	return errors.Join(fails...)
}

func (top *Topology) CommitExchanges() error {
	var (
		channel *amqp.Channel
		err     error
		targets []ExchangeDef
	)

	if targets = top.config.Exchanges; len(targets) == 0 {
		return nil
	}

	if channel, err = top.conn.GetChannel(); err != nil {
		return err
	}

	fails := make([]error, len(top.config.Exchanges))
	for _, ex := range top.config.Exchanges {
		if err = channel.ExchangeDeclare(ex.Name, ex.Type, ex.Durable, ex.AutoDelete, false, false, ex.Arguments); err != nil {
			fails = append(fails, err)
		}
	}

	return errors.Join(fails...)
}
