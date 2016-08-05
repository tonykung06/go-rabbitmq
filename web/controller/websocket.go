package controller

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"sync"

	"github.com/go-rabbitmq/dataTransferObject"
	"github.com/go-rabbitmq/queueUtils"
	"github.com/go-rabbitmq/web/model"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

const url = "amqp://guest:guest@localhost:5672"

type websocketController struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	sockets  []*websocket.Conn
	mutex    sync.Mutex
	upgrader websocket.Upgrader
}

func newWebsocketController() *websocketController {
	wsc := new(websocketController)
	wsc.conn, wsc.ch = queueUtils.GetChannel(url)
	wsc.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	go wsc.listenForSources()
	go wsc.listenForMessages()
	return wsc
}

func (wsc *websocketController) handleMessage(w http.ResponseWriter, r *http.Request) {
	socket, _ := wsc.upgrader.Upgrade(w, r, nil)
	wsc.addSocket(socket)
	go wsc.listenForDiscoveryRequests(socket)
}

func (wsc *websocketController) addSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	//append() is not thread-safe
	wsc.sockets = append(wsc.sockets, socket)
	wsc.mutex.Unlock()
}

func (wsc *websocketController) removeSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	socket.Close()
	for i := range wsc.sockets {
		if wsc.sockets[i] == socket {
			wsc.sockets = append(wsc.sockets[:i], wsc.sockets[i+1:]...)
			break
		}
	}

	// for i := range wsc.sockets {
	// 	if wsc.sockets[i] == socket {
	// 		wsc.sockets[i], wsc.sockets[len(wsc.sockets)-1] = wsc.sockets[len(wsc.sockets)-1], wsc.sockets[i]
	// 		wsc.sockets = wsc.sockets[:len(wsc.sockets)-1]
	// 		break
	// 	}
	// }

	wsc.mutex.Unlock()
}

func (wsc *websocketController) listenForSources() {
	q := queueUtils.GetQueue("", wsc.ch, true)
	wsc.ch.QueueBind(q.Name, "", queueUtils.WebappSourceExchange, false, nil)
	msgs, _ := wsc.ch.Consume(q.Name, "", true, false, false, false, nil)
	for msg := range msgs {
		sensor, _ := model.GetSensorByName(string(msg.Body))
		wsc.sendMessage(message{
			Type: "source",
			Data: sensor,
		})
	}
}

func (wsc *websocketController) listenForMessages() {
	q := queueUtils.GetQueue("", wsc.ch, true)
	wsc.ch.QueueBind(q.Name, "", queueUtils.WebappReadingsExchange, false, nil)
	msgs, _ := wsc.ch.Consume(q.Name, "", true, false, false, false, nil)
	for msg := range msgs {
		buf := bytes.NewBuffer(msg.Body)
		dec := gob.NewDecoder(buf)
		sm := dataTransferObject.SensorMessage{}
		dec.Decode(&sm)
		wsc.sendMessage(message{
			Type: "reading",
			Data: sm,
		})
	}
}

func (wsc *websocketController) sendMessage(msg message) {
	socketsToRemove := []*websocket.Conn{}
	for _, socket := range wsc.sockets {
		err := socket.WriteJSON(msg)
		if err != nil {
			socketsToRemove = append(socketsToRemove, socket)
		}
		for _, socket := range socketsToRemove {
			wsc.removeSocket(socket)
		}
	}
}

func (wsc *websocketController) listenForDiscoveryRequests(socket *websocket.Conn) {
	for {
		msg := message{}
		err := socket.ReadJSON(&msg)

		if err != nil {
			wsc.removeSocket(socket)
			break
		}

		if msg.Type == "discover" {
			wsc.ch.Publish("", queueUtils.WebappDiscoveryQueue, false, false, amqp.Publishing{})
		}
	}
}

type message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
