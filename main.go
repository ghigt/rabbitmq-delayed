package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/jinzhu/configor"
)

func run() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.DebugLevel)

	var config Config

	err := configor.Load(&config, "config/config.json")
	if err != nil {
		log.Fatalf("run: failed to init config: %v", err)
	}

	rmq, err := initRabbitMQ(config.AMQP)
	if err != nil {
		log.Fatalf("run: failed to init rabbitmq: %v", err)
	}
	defer rmq.Shutdown()

	err = rmq.Publish([]byte("hello"))
	if err != nil {
		log.Fatalf("run: failed to publish into rabbitmq: %v", err)
	}
}

func main() {
	run()
}
