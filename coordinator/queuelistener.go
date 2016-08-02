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

//amqp.Delivery is receive-only channel
type QueueListener struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery //a registry of sources this coordinator is listening on, this is used to close down the listener when the associated sensors go offline
	ea      *EventAggregator
}

func NewQueueListener() *QueueListener {
	ql := QueueListener{
		sources: make(map[string]<-chan amqp.Delivery),
		ea:      NewEventAggregator(),
	}
	ql.conn, ql.ch = queueUtils.GetChannel(url)
	return &ql
}

func (ql *QueueListener) DiscoverSensors() {
	ql.ch.ExchangeDeclare(
		queueUtils.SensorDiscoveryExchange, // name,
		"fanout", // kind, direct, topic, header or fanout
		false,    // durable,
		false,    // autoDelete, delete the exchange if there are no bindings present
		false,    // internal, true to reject external publishing requests, used in advanced scenarios where exchanges are bound together to form complicated topologies in the broker
		false,    // noWait,
		nil,      // args,
	)

	ql.ch.Publish(
		queueUtils.SensorDiscoveryExchange, //exchange
		"",                // key,
		false,             // mandatory,
		false,             // immediate,
		amqp.Publishing{}, // msg,
	)
}

func (ql *QueueListener) ListenForNewSouce() {
	//by default, newly created queue is bound to default exchange
	q := queueUtils.GetQueue(
		"", //empty string queue name, rabbitmq server will generate a unique one
		ql.ch,
	)

	//rebind the queue to fanout exchange
	ql.ch.QueueBind(
		q.Name,       //name, this is the generated unique queue name
		"",           //key, if this were the default exchange, the default routing key would be the same as the queue name. For rebinding to the default exchange, the key can be changed and is used by the exchange to identify the queue, also you can bind the same queue to the same exchange several times by using different keys. For fanout exchange, this key is useless.
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

	ql.DiscoverSensors()

	fmt.Println("listening for new sources")

	//whenever a new fanout msg comes in, it means there is a new sensor ready to publish data via a new queue
	//then we need to listen to that new queue, whose name is discovered here
	//NOTE: when we have more than one consumers/coordinators listening on this queue, the default exchange will equally distribute messages to the consumers (one msg will only receive by one consumer)
	for msg := range msgs {
		if ql.sources[string(msg.Body)] == nil {
			fmt.Println("new source discovered")
			sourceChan, _ := ql.ch.Consume(
				string(msg.Body), // queue,
				"",               // consumer,
				true,             // autoAck,
				false,            // exclusive,
				false,            // noLocal,
				false,            // noWait,
				nil,              // args,
			)
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

		ed := EventData{
			Name:      sd.Name,
			Timestamp: sd.Timestamp,
			Value:     sd.Value,
		}
		ql.ea.PublishEvent("MessageReceived_"+msg.RoutingKey, ed)
	}
}
