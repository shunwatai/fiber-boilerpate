package rabbitmq

import (
	"fmt"
	"golang-api-starter/internal/config"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

var cfg = config.Cfg

func GetUrl() string {
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/", *cfg.RabbitMqConf.User, *cfg.RabbitMqConf.Pass, *cfg.RabbitMqConf.Host, *cfg.RabbitMqConf.Port)

	return connectionString
}

func NewRabbitMQ(url, queueName string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{conn, channel, queue}, nil
}

func (r *RabbitMQ) Publish(body []byte) error {
	err := r.channel.Publish(
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	return err
}

func (r *RabbitMQ) Close() error {
	return r.conn.Close()
}
