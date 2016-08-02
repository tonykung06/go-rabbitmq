package coordinator

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/go-rabbitmq/dataTransferObject"
	"github.com/go-rabbitmq/queueUtils"
	"github.com/streadway/amqp"
)

const url = "amqp://guest:guest@localhost:5672"

type QueueListener struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery
}

func NewQueueListener() *QueueListener {
	ql := QueueListener{
		sources: make(map[string]<-chan amqp.Delivery),
	}
	ql.conn, ql.ch = queueUtils.GetChannel(url)
	return &ql
}
func (ql *QueueListener) ListenForNewSouce() {
	q := queueUtils.GetQueue("", ql.ch)

	//rebind
	ql.ch.QueueBind(
		q.Name,       //name,
		"",           //key,
		"amq.fanout", // exchange,
		false,        // noWait,
		nil,          // args
	)

	msgs, _ := ql.ch.Consume(
		q.Name, // queue,
		"",     // consumer,
		true,   // autoAck,
		false,  // exclusive,
		false,  // noLocal,
		false,  // noWait,
		nil,    // args,
	)

	for msg := range msgs {
		sourceChan, _ := ql.ch.Consume(
			string(msg.Body), // queue,
			"",               // consumer,
			true,             // autoAck,
			false,            // exclusive,
			false,            // noLocal,
			false,            // noWait,
			nil,              // args,
		)

		if ql.sources[string(msg.Body)] == nil {
			ql.sources[string(msg.Body)] = sourceChan
			go ql.AddListener(sourceChan)
		}
	}
}

func (ql *QueueListener) AddListener(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		r := bytes.NewReader(msg.Body)
		d := gob.NewDecoder(r)
		sd := new(dataTransferObject.SensorMessage)
		d.Decode(sd)

		fmt.Printf("Received message: %v\n", sd)
	}
}
