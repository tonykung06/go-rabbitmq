package coordinator

import (
	"bytes"
	"encoding/gob"

	"github.com/go-rabbitmq/dataTransferObject"
	"github.com/go-rabbitmq/queueUtils"
	"github.com/streadway/amqp"
)

type WebappConsumer struct {
	er      EventRaiser
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources []string
}

func NewWebappConsumer(er EventRaiser) *WebappConsumer {
	wc := WebappConsumer{
		er: er,
	}
	wc.conn, wc.ch = queueUtils.GetChannel(url)
	queueUtils.GetQueue(queueUtils.PersistReadingsQueue, wc.ch, false)

	go wc.ListenForDiscoveryRequests()
	wc.er.AddListener("DataSourceDiscovered", func(eventData interface{}) {
		wc.SubscribeToDataEvent(eventData.(string))
	})
	wc.ch.ExchangeDeclare(queueUtils.WebappSourceExchange, "fanout", false, false, false, false, nil)
	wc.ch.ExchangeDeclare(queueUtils.WebappReadingsExchange, "fanout", false, false, false, false, nil)
	return &wc
}

func (wc *WebappConsumer) ListenForDiscoveryRequests() {
	q := queueUtils.GetQueue(queueUtils.WebappDiscoveryQueue, wc.ch, false)
	msgs, _ := wc.ch.Consume(q.Name, "", true, false, false, false, nil)
	for range msgs {
		for _, src := range wc.sources {
			wc.SendMessageSource(src)
		}
	}
}

func (wc *WebappConsumer) SendMessageSource(src string) {
	wc.ch.Publish(queueUtils.WebappSourceExchange, "", false, false, amqp.Publishing{Body: []byte(src)})
}

func (wc *WebappConsumer) SubscribeToDataEvent(eventName string) {
	for _, v := range wc.sources {
		if v == eventName {
			return
		}
	}
	wc.sources = append(wc.sources, eventName)
	wc.SendMessageSource(eventName)
	wc.er.AddListener("MessageReceived_"+eventName, func() func(interface{}) {
		return func(eventData interface{}) {
			ed := eventData.(EventData)
			sm := dataTransferObject.SensorMessage{
				Name:      ed.Name,
				Value:     ed.Value,
				Timestamp: ed.Timestamp,
			}
			buf := new(bytes.Buffer)
			enc := gob.NewEncoder(buf)
			enc.Encode(sm)
			msg := amqp.Publishing{
				Body: buf.Bytes(),
			}
			wc.ch.Publish(queueUtils.WebappReadingsExchange, "", false, false, msg)
		}
	}())
}
