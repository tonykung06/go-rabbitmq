package main

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/go-rabbitmq/dataTransferObject"
	"github.com/go-rabbitmq/datamanager"
	"github.com/go-rabbitmq/queueUtils"
)

const url = "amqp://guest:guest@localhost:5672"

func main() {
	conn, ch := queueUtils.GetChannel(url)
	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(queueUtils.PersistReadingsQueue, "", false, true, false, false, nil)
	if err != nil {
		log.Fatalln("Failed to get access to message queue")
	}
	for msg := range msgs {
		buf := bytes.NewReader(msg.Body)
		dec := gob.NewDecoder(buf)
		sd := &dataTransferObject.SensorMessage{}
		dec.Decode(sd)

		err := datamanager.SaveReading(sd)
		if err != nil {
			log.Printf("Failed to save reading from sensor %v. Error: %s", sd.Name, err.Error())
		} else {
			msg.Ack(false)
		}
	}
}
