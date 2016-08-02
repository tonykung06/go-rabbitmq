package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	go client()
	go server()
	var a string
	fmt.Scanln(&a)
}
func client() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	//msgChannel is a golang channel
	msgChannel, err := ch.Consume(
		q.Name, //queue name
		"",     //consumer identifier, rabbitmq server will roughly load-balances amongst multiple consumers to the queue in the direct exchange. This identifier can also be used to cancel the subscription/connection so that rabbitmq server will no longer distribute msg to it. Empty string tells rabbitmq server generate the identifier.
		true,   //autoAck, automatically ack the receival of the msg, so that the server could remove the msg as quickly as possible. Set this to false, when we want to manually acknowledge after saving the msg to database successfully.
		false,  //exclusive, ensure this client is the only consumer of this queue
		false,  //noLocal, prevent rabbitmq server from sending msg to the consumers that are on the same connection as the publisher/sender
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")
	for msg := range msgChannel {
		log.Printf("Received message: %s", msg.Body)
	}
}

func server() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("Hello RabbitMQ"),
	}

	for {
		ch.Publish(
			"",     //exchange, empty string to denote default exchange "direct"
			q.Name, //key, the routing key for the queue
			false,  //mandatory, if true, throwing error if the queue is not there to receive the publishing
			false,  //immediate, if true, error out if there are no active consumers on the queue now
			msg,
		)
	}
}

//technically, we need to publish messages to an exchange,
//there is a shortcut to deal with default exchange (TYPE: direct) and directly publish to a queue (behind the scenes, it goes through the default exchange)
func getQueue() (*amqp.Connection, *amqp.Channel, *amqp.Queue) { //separating connection and channel to allow multiple channels over a single connection, multiple queues over a single channel, to reduce resouce usage on client and server
	conn, err := amqp.Dial("amqp://guest@localhost:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	//create a queue or use the existing one if there is one, when the configs of the existing one doesn't match this request, rabbitmq will reject the request and error out
	q, err := ch.QueueDeclare(
		"hello", //queue name
		false,   //durable, if true, save the msg to disk when it is added to the queue, so that will survive server restarts
		false,   //autoDelete, if true, rabbitmq server will delete the msg if there are no active consumers on that queue. If false, rabbitmq server will keep that msg around until a consumer comes along to receive it
		false,   //exclusive, if true, only allowing access from the connection that creates this queue. Otherwise, multiple connections to the same queue.
		false,   //noWait, if true, dont create new queue, only try to get an existing queue, error out if there is no existing one
		nil,
	)
	failOnError(err, "Failed to declare a queue")
	return conn, ch, &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
