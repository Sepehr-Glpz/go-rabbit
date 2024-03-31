package rabbit

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

func NewConsumerHandler(connection IRabbitConnection, serializer IMessageSerializer) *ConsumerHandler {
	return &ConsumerHandler{
		conn:       connection,
		serializer: serializer,
		consumers:  make(map[string]*consumer),
		doneChan:   make(chan struct{}),
		lock:       new(sync.RWMutex),
	}
}

type (
	ConsumerHandler struct {
		conn          IRabbitConnection
		serializer    IMessageSerializer
		consumers     map[string]*consumer
		doneChan      chan struct{}
		lock          *sync.RWMutex
		isRecoverable bool
	}
	ConsumeResult byte
	ConsumerFunc  func(IConsumeActions) ConsumeResult
)

const (
	Reject        ConsumeResult = iota
	RejectRequeue ConsumeResult = iota
	Ack           ConsumeResult = iota
)

func (con *ConsumerHandler) Consume(queueName string, name string, exclusive bool, props amqp.Table, handler ConsumerFunc) error {
	var (
		channel    *amqp.Channel
		err        error
		deliveries <-chan amqp.Delivery
	)

	con.lock.Lock()
	defer con.lock.Unlock()

	if err = con.CancelConsumer(name); err != nil {
		return err
	}
	//if cons, ok := con.consumers[name]; ok {
	//	if err = cons.channel.Cancel(cons.name, false); err != nil {
	//		return err
	//	}
	//}

	if channel, err = con.conn.GetChannel(); err != nil {
		return err
	}

	if deliveries, err = channel.Consume(queueName, name, false, exclusive, false, false, props); err != nil {
		return err
	}

	newConsumer := &consumer{
		queueName:  queueName,
		name:       name,
		exclusive:  exclusive,
		props:      props,
		handler:    handler,
		cancel:     make(chan struct{}),
		deliveries: deliveries,
		channel:    channel,
	}
	con.consumers[name] = newConsumer

	go listenConsumer(newConsumer, con.serializer)

	return nil
}

func listenConsumer(consumer *consumer, serializer IMessageSerializer) {
	listen := true
	for listen {
		select {
		case msg, open := <-consumer.deliveries:
			{
				if !open {
					listen = false
					break
				}

				actions := &consumeAction{
					msg:        &msg,
					serializer: serializer,
				}

				switch consumer.handler(actions) {
				case RejectRequeue:
					{
						for msg.Nack(false, true) != nil {
							time.Sleep(time.Millisecond * 300)
						}
						break
					}
				case Reject:
					{
						for msg.Nack(false, false) != nil {
							time.Sleep(time.Millisecond * 300)
						}
						break
					}
				case Ack:
					{
						for msg.Ack(false) != nil {
							time.Sleep(time.Millisecond * 300)
						}
						break
					}
				}
			}
		case <-consumer.cancel:
			{
				listen = false
				break
			}
		}
	}
}

func (con *ConsumerHandler) CancelConsumer(name string) error {
	var (
		channel *amqp.Channel
		err     error
	)
	if cons, ok := con.consumers[name]; ok {
		if channel, err = con.conn.GetChannel(); err != nil {
			return err
		}
		defer func() {
			_ = channel.Close()
		}()

		if err = channel.Cancel(name, false); err != nil {
			return err
		}
		if err = cons.channel.Close(); err != nil {
			return err
		}
		close(cons.cancel)
		delete(con.consumers, name)
	}
	return nil
}

func (con *ConsumerHandler) Close() error {
	var (
		err error
	)

	con.lock.Lock()
	defer con.lock.Unlock()

	for _, cons := range con.consumers {
		if err = con.CancelConsumer(cons.name); err != nil {
			return err
		}
	}

	return nil
}

func (con *ConsumerHandler) EnableConsumerRecovery() {
	con.lock.RLock()
	defer con.lock.RUnlock()

	if con.isRecoverable {
		return
	}

	go func() {
		for listen := true; listen; {
			select {
			case _, open := <-con.conn.notifyReconnect():
				{
					if !open {
						listen = false
						break
					}

					for _, cons := range con.consumers {
						for con.Consume(cons.queueName, cons.name, cons.exclusive, cons.props, cons.handler) != nil {
							time.Sleep(time.Second)
						}
					}
				}
			case <-con.doneChan:
				{
					listen = false
					break
				}
			}
		}
	}()
	con.isRecoverable = true
}

type (
	consumeAction struct {
		msg        *amqp.Delivery
		serializer IMessageSerializer
	}
	consumer struct {
		queueName  string
		name       string
		exclusive  bool
		props      amqp.Table
		handler    ConsumerFunc
		cancel     chan struct{}
		deliveries <-chan amqp.Delivery
		channel    *amqp.Channel
	}
)

func (action *consumeAction) GetMessageAs(result any) error {
	return action.serializer.Deserialize(action.msg.Body, result)
}

func (action *consumeAction) GetRawBody() []byte {
	result := make([]byte, len(action.msg.Body))
	copy(result, action.msg.Body)
	return result
}

func (action *consumeAction) GetHeaders() amqp.Table {
	return action.msg.Headers
}
