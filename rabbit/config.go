package rabbit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host                  string
	Port                  uint16
	Endpoints             []Endpoint
	VirtualHost           string
	Username              string
	Password              string
	AutoReconnect         bool
	AutoReconnectInterval time.Duration
	Mapping               Mapping
}

type Mapping struct {
	Exchanges []ExchangeDef
	Queues    []QueueDef
	Bindings  []BindingDef
}

type QueueDef struct {
	Id         int
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	Arguments  map[string]any
}

type ExchangeDef struct {
	Id         int
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Arguments  map[string]any
}

type BindingDef struct {
	FromExchangeId int
	ToQueueId      int
	Keys           []string
	Arguments      map[string]any
}

type Endpoint struct {
	Host string
	Port uint16
}

func (end *Endpoint) String() string {
	return fmt.Sprintf("%s:%v", end.Host, end.Port)
}

func (cnf Config) createUrl() string {
	builder := &strings.Builder{}
	builder.WriteString("amqp://")
	if cnf.Endpoints == nil || len(cnf.Endpoints) == 0 {
		builder.WriteString(cnf.Host)
		builder.WriteRune(':')
		builder.WriteString(strconv.Itoa(int(cnf.Port)))

		return builder.String()
	}

	endpoints := make([]string, len(cnf.Endpoints))
	for _, endpoint := range cnf.Endpoints {
		var (
			host string
			port uint16
		)

		if host = endpoint.Host; host == "" {
			host = cnf.Host
		}
		if port = endpoint.Port; port == 0 {
			port = cnf.Port
		}

		address := fmt.Sprintf("%s:%s", host, strconv.Itoa(int(port)))
		endpoints = append(endpoints, address)
	}

	builder.WriteString(strings.Join(endpoints, ","))

	return builder.String()
}
