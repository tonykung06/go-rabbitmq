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

var name = flag.String("name", "senor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "maximum value for generated readings")
var min = flag.Float64("min", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")

var r = rand.New(rand.NewSource(time.Now().UnixNano()))
var value = r.Float64()*(*max-*min) + *min
var nom = (*max-*min)/2 + *min

func main() {
	flag.Parse()

	conn, ch := queueUtils.GetChannel(url)
	defer conn.Close()
	defer ch.Close()

	//technically, we only need to tell the broker the routing key(queue name) and dont need to create the queue here
	//but we need to ensure the queue is there
	dataQueue := queueUtils.GetQueue(*name, ch)
	sensorQueue := queueUtils.GetQueue(queueUtils.SensorListQueue, ch)
	msg := amqp.Publishing{Body: []byte(*name)}
	ch.Publish("", sensorQueue.Name, false, false, msg)

	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")
	signal := time.Tick(dur)
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	for range signal {
		calcValue()
		reading := dataTransferObject.SensorMessage{
			Name:      *name, //this tells the message broker which queue to use
			Value:     value,
			Timestamp: time.Now(),
		}
		buf.Reset()
		enc.Encode(reading)
		msg := amqp.Publishing{
			Body: buf.Bytes(),
		}
		ch.Publish("", dataQueue.Name, false, false, msg)
		log.Printf("Reading sent. Value: %v\n", value)
	}
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
