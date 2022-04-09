package queue

import (
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Queues struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

func NewQueues(addr string) Queues {
	conn, err := amqp.Dial(addr)
	failOnError(err, "Failed to connect to RabbitMQ")

	// open channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	// declare queues
	_, err = ch.QueueDeclare(
		"task", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")

	_, err = ch.QueueDeclare(
		"result", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")

	return Queues{
		Conn: conn,
		Ch:   ch,
	}
}

func (q *Queues) Deconstruct() {
	q.Ch.Close()
	q.Conn.Close()
}
