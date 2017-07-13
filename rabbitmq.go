package main

import (
	"time"

	"github.com/apex/log"
	"github.com/streadway/amqp"
)

// RabbitMQ stores rabbitmq's connection information
// it also handles disconnection (purpose of URL and QueueName storage)
type RabbitMQ struct {
	URL        string
	Conn       *amqp.Connection
	Chann      *amqp.Channel
	Queue      amqp.Queue
	closeChann chan *amqp.Error
	quitChann  chan bool
}

func initRabbitMQ(config AMQP) (*RabbitMQ, error) {
	rmq := &RabbitMQ{
		URL: config.URL,
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

	rmq.closeChann = make(chan *amqp.Error)
	rmq.Conn.NotifyClose(rmq.closeChann)

	rmq.Queue, err = rmq.Chann.QueueDeclare("lci.content.create", true, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Info("connection to rabbitMQ established")

	return nil
}

// Shutdown closes rabbitmq's connection
func (rmq *RabbitMQ) Shutdown() {
	rmq.quitChann <- true

	log.Info("shutting down rabbitMQ's connection...")

	<-rmq.quitChann
}

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

// Publish sends the given body to the channel
func (rmq *RabbitMQ) Publish(body []byte) error {

	log.Debugf("publishing to %q %q", rmq.Queue.Name, body)

	return rmq.Chann.Publish("", rmq.Queue.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
	})
}
