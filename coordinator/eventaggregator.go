package coordinator

import "time"

type EventRaiser interface {
	AddListener(string, func(interface{}))
}

func NewEventAggregator() *EventAggregator {
	ea := EventAggregator{
		listeners: make(map[string][]func(interface{})),
	}
	return &ea
}

func (ea *EventAggregator) AddListener(eventName string, cb func(interface{})) {
	ea.listeners[eventName] = append(ea.listeners[eventName], cb)
}

func (ea *EventAggregator) PublishEvent(eventName string, eventData interface{}) {
	if ea.listeners[eventName] != nil {
		for _, r := range ea.listeners[eventName] {
			//passing a copy of eventData
			r(eventData)
		}
	}
}

type EventAggregator struct {
	listeners map[string][]func(interface{})
}

//although this struct is quite similar to dataTransferObject, they mean different things and should be kept separate to allow evolving
type EventData struct {
	Name      string
	Value     float64
	Timestamp time.Time
}
