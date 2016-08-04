package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-rabbitmq/dataTransferObject"
	"github.com/go-rabbitmq/queueUtils"
	"github.com/streadway/amqp"
)

var url = "amqp://guest:guest@localhost:5672"

var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "maximum value for generated readings")
var min = flag.Float64("min", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")

var r = rand.New(rand.NewSource(time.Now().UnixNano()))
var value float64
var nom float64

func main() {
	flag.Parse()
	value = r.Float64()*(*max-*min) + *min
	nom = (*max-*min)/2 + *min

	conn, ch := queueUtils.GetChannel(url)
	defer conn.Close()
	defer ch.Close()

	//technically, we only need to tell the broker the routing key(queue name) and dont need to create the queue here
	//but we need to ensure the queue is there
	dataQueue := queueUtils.GetQueue(*name, ch, false)

	publishQueueName(ch)
	discoveryQueue := queueUtils.GetQueue("", ch, true)
	ch.QueueBind(
		discoveryQueue.Name, // name,
		"",                  // key,
		queueUtils.SensorDiscoveryExchange, // exchange,
		false, // noWait,
		nil,   // args,
	)

	go listenForDiscoverRequest(discoveryQueue.Name, ch)

	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")
	signal := time.Tick(dur)
	buf := new(bytes.Buffer)
	for range signal {
		calcValue()
		reading := dataTransferObject.SensorMessage{
			Name:      *name, //this tells the message broker which queue to use
			Value:     value,
			Timestamp: time.Now(),
		}
		buf.Reset()
		enc := gob.NewEncoder(buf)
		enc.Encode(reading)
		msg := amqp.Publishing{
			Body: buf.Bytes(),
		}
		ch.Publish("", dataQueue.Name, false, false, msg)
		log.Printf("Reading sent. Value: %v\n", value)
	}
}

func listenForDiscoverRequest(name string, ch *amqp.Channel) {
	msgs, _ := ch.Consume(
		name,  // queue,
		"",    // consumer,
		true,  // autoAck,
		false, // exclusive,
		false, // noLocal,
		false, // noWait,
		nil,   // args,
	)

	for range msgs {
		publishQueueName(ch)
	}
}

func publishQueueName(ch *amqp.Channel) {
	//1.
	//this "direct" exchange doesn't allow multiple consumers, so we need to use fanout exchange to allow multiple consumers to listen on this
	// sensorQueue := queueUtils.GetQueue(queueUtils.SensorListQueue, ch)
	// msg := amqp.Publishing{Body: []byte(*name)}
	// ch.Publish("", sensorQueue.Name, false, false, msg)

	//2.
	//that is the responsibility of each individual consumer to create a queue to receive messages from fanout exchange
	//for fanout exchange, msg will be lost if no active queues bound to it
	msg := amqp.Publishing{Body: []byte(*name)}
	ch.Publish(
		"amq.fanout",
		"", //queue name is unknown here for the fanout exchange
		false,
		false,
		msg,
	)
}

func calcValue() {
	var maxStep, minStep float64
	if value < nom {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (nom - *min)
	} else {
		maxStep = *stepSize * (*max - value) / (*max - nom)
		minStep = -1 * *stepSize
	}
	value += r.Float64()*(maxStep-minStep) + minStep
}
