package main

import (
	"time"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// RabbitMQ stores rabbitmq's connection information
// it also handles disconnection (purpose of URL and QueueName storage)
type RabbitMQ struct {
	URL        string
	Exchange   string
	Conn       *amqp.Connection
	Chann      *amqp.Channel
	Queue      amqp.Queue
	closeChann chan *amqp.Error
	quitChann  chan bool
}

func initRabbitMQ(config AMQP) (*RabbitMQ, error) {
	rmq := &RabbitMQ{
		URL:      config.URL,
		Exchange: config.Exchange,
	}

	err := rmq.load()
	if err != nil {
		return nil, err
	}

	rmq.quitChann = make(chan bool)

	go rmq.handleDisconnect()

	return rmq, err
}

func (rmq *RabbitMQ) load() error {
	var err error

	rmq.Conn, err = amqp.Dial(rmq.URL)
	if err != nil {
		return err
	}

	rmq.Chann, err = rmq.Conn.Channel()
	if err != nil {
		return err
	}

	log.Info("connection to rabbitMQ established")

	rmq.closeChann = make(chan *amqp.Error)
	rmq.Conn.NotifyClose(rmq.closeChann)

	// declare exchange if not exist
	err = rmq.Chann.ExchangeDeclare(rmq.Exchange, "direct", true, false, false, false, nil)
	if err != nil {
		return errors.Wrapf(err, "declaring exchange %q", rmq.Exchange)
	}

	args := make(amqp.Table)
	args["x-delayed-type"] = "direct"
	err = rmq.Chann.ExchangeDeclare("delayed", "x-delayed-message", true, false, false, false, args)
	if err != nil {
		return errors.Wrapf(err, "declaring exchange %q", "delayed")
	}

	err = declareConsumer(rmq)
	if err != nil {
		return err
	}

	return nil
}

// declareConsumer declares all queues and bindings for the consumer
func declareConsumer(rmq *RabbitMQ) error {
	var err error

	// rmq.Queue, err = rmq.Chann.QueueDeclare("user-created-queue", true, false, false, false, nil)
	// if err != nil {
	// 	return err
	// }
	// err = rmq.Chann.QueueBind(rmq.Queue.Name, "user.event.create", rmq.Exchange, false, nil)
	// if err != nil {
	// 	return err
	// }

	delayedQueue, err := rmq.Chann.QueueDeclare("user-published-queue", true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = rmq.Chann.QueueBind(delayedQueue.Name, "user.event.publish", "delayed", false, nil)
	if err != nil {
		return err
	}

	// Set our quality of service.  Since we're sharing 3 consumers on the same
	// channel, we want at least 2 messages in flight.
	err = rmq.Chann.Qos(2, 0, false)
	if err != nil {
		return err
	}

	published, err := rmq.Chann.Consume(
		"user-published-queue",
		"user-published-consumer",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		return err
	}

	go consume(published)

	return nil
}

func consume(ds <-chan amqp.Delivery) {
	for {
		log.Debug("start listenning to consume")

		select {
		case d, ok := <-ds:
			if !ok {
				return
			}
			log.Infof("consume %s", string(d.Body))
			d.Ack(false)
		}
	}
}

// Shutdown closes rabbitmq's connection
func (rmq *RabbitMQ) Shutdown() {
	rmq.quitChann <- true

	log.Info("shutting down rabbitMQ's connection...")

	<-rmq.quitChann
}

// handleDisconnect handle a disconnection trying to reconnect every 5 seconds
func (rmq *RabbitMQ) handleDisconnect() {
	for {
		select {
		case errChann := <-rmq.closeChann:
			if errChann != nil {
				log.Errorf("rabbitMQ disconnection: %v", errChann)
			}
		case <-rmq.quitChann:
			rmq.Conn.Close()
			log.Info("...rabbitMQ has been shut down")
			rmq.quitChann <- true
			return
		}

		log.Info("...trying to reconnect to rabbitMQ...")

		time.Sleep(5 * time.Second)

		if err := rmq.load(); err != nil {
			log.Errorf("rabbitMQ error: %v", err)
		}
	}
}

// Publish sends the given body on the routingKey to the channel
func (rmq *RabbitMQ) Publish(routingKey string, body []byte) error {
	return rmq.publish(rmq.Exchange, routingKey, body, 0)
}

// PublishWithDelay sends the given body on the routingKey to the channel with a delay
func (rmq *RabbitMQ) PublishWithDelay(routingKey string, body []byte, delay int64) error {
	return rmq.publish("delayed", routingKey, body, delay)
}

func (rmq *RabbitMQ) publish(exchange string, routingKey string, body []byte, delay int64) error {
	headers := make(amqp.Table)

	log.Debugf("publishing to %q %q", routingKey, body)

	if delay != 0 {
		headers["x-delay"] = delay
	}

	return rmq.Chann.Publish(exchange, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
		Headers:      headers,
	})
}
