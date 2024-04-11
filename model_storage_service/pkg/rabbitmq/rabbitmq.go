package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

type RabbitMQConnection struct {
	Connection  *amqp.Connection
	ChannelPool chan *amqp.Channel
}

func NewRabbitMQConnection(uri string, poolSize int) (*RabbitMQConnection, error) {
	op := "rabbitmq.NewRabbitMQConnection"

	var conn *amqp.Connection
	err := DoWithTries(func() error {
		var err error
		conn, err = amqp.Dial(uri)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}, 5, 5*time.Second)
	if err != nil {
		return nil, err
	}

	pool := make(chan *amqp.Channel, poolSize)
	for i := 0; i < poolSize; i++ {
		ch, err := conn.Channel()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		pool <- ch
	}

	return &RabbitMQConnection{
		Connection:  conn,
		ChannelPool: pool,
	}, nil
}

func (rmq *RabbitMQConnection) Close() {
	rmq.Connection.Close()
	close(rmq.ChannelPool)
	for ch := range rmq.ChannelPool {
		ch.Close()
	}
}

func (rmq *RabbitMQConnection) GetChannel() (*amqp.Channel, error) {
	select {
	case ch := <-rmq.ChannelPool:
		return ch, nil
	default:
		return rmq.Connection.Channel()
	}
}

func (rmq *RabbitMQConnection) ReleaseChannel(ch *amqp.Channel) {
	select {
	case rmq.ChannelPool <- ch:
	default:
		ch.Close()
	}
}

func (rmq *RabbitMQConnection) InitQueues(queues []string) error {
	op := "rabbitmq.RabbitMQConnection.InitQueues"
	ch, err := rmq.GetChannel()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer rmq.ReleaseChannel(ch)
	for _, queue := range queues {
		_, err = ch.QueueDeclare(
			queue,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (rmq *RabbitMQConnection) Publish(queue string, message []byte) error {
	op := "rabbitmq.RabbitMQConnection.Publish"
	ch, err := rmq.GetChannel()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer rmq.ReleaseChannel(ch)

	err = ch.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (rmq *RabbitMQConnection) Consume(queue string) (<-chan amqp.Delivery, func(), error) {
	op := "rabbitmq.RabbitMQConnection.Consume"
	ch, err := rmq.GetChannel()
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	q, err := ch.QueueDeclare(
		queue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		rmq.ReleaseChannel(ch)
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		rmq.ReleaseChannel(ch)
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	closeFunc := func() { rmq.ReleaseChannel(ch) }

	return msgs, closeFunc, nil
}

func DoWithTries(fn func() error, attempts int, delay time.Duration) (err error) {
	var prevErr error
	for attempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attempts--
			prevErr = err
			continue
		}
		return nil
	}
	return prevErr
}
