package infrastructure

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"rabbit-test/infrastructure/rabbit"
	"time"
)

func ProvideInfrastructure() fx.Option {
	return fx.Module("infrastructure",
		fx.Provide(
			fx.Annotate(
				rabbit.NewConnection,
				fx.As(new(rabbit.IRabbitConnection)),
			),
			fx.Annotate(
				rabbit.NewJsonSerializer,
				fx.As(new(rabbit.IMessageSerializer)),
			),
			fx.Annotate(
				rabbit.NewPublisher,
				fx.As(new(rabbit.IPublisher)),
			),
			fx.Annotate(
				rabbit.NewClient,
				fx.As(new(rabbit.IRabbitClient)),
			),
			fx.Annotate(
				rabbit.NewConsumerHandler,
				fx.As(new(rabbit.IConsumer)),
			),
			fx.Annotate(
				rabbit.NewTopology,
				fx.As(new(rabbit.ITopology)),
			),
			NewConfig,
		),
	)
}

func NewConfig(config *viper.Viper) (rabbit.Config, error) {
	rabbitConf := config.Sub("RabbitMQ")
	var (
		endpoints = make([]rabbit.Endpoint, 0)
		mapping   = new(rabbit.Mapping)
		err       error
	)

	if err = rabbitConf.UnmarshalKey("Endpoints", &endpoints); err != nil {
		return *new(rabbit.Config), err
	}
	if err = rabbitConf.UnmarshalKey("Mapping", mapping); err != nil {
		return *new(rabbit.Config), err
	}

	return rabbit.Config{
		Host:                  rabbitConf.GetString("Host"),
		Port:                  rabbitConf.GetUint16("Port"),
		Endpoints:             endpoints,
		VirtualHost:           rabbitConf.GetString("VirtualHost"),
		Username:              rabbitConf.GetString("Username"),
		Password:              rabbitConf.GetString("Password"),
		AutoReconnect:         true,
		AutoReconnectInterval: time.Second * 3,
		Mapping:               *mapping,
	}, nil
}
