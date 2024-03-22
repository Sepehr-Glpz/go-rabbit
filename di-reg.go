package infrastructure

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"test-5/infrastructure/rabbit"
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
			NewConfig,
		),
	)
}

func NewConfig(config *viper.Viper) (rabbit.Config, error) {
	rabbitConf := config.Sub("RabbitMQ")
	var (
		endpoints = make([]rabbit.Endpoint, 0)
		err       error
	)

	if err = rabbitConf.UnmarshalKey("Endpoints", &endpoints); err != nil {
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
	}, nil
}
