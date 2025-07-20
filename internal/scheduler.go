package demokubenet

import (
	"log"
	"sync"
)

type EventBus struct {
	handlers   map[EventType][]EventHandler
	eventQueue chan EventWrapper
}

type EventWrapper struct {
	Event Event
	Wg    *sync.WaitGroup
}

func NewEventBus(queueSize int) *EventBus {
	eb := &EventBus{
		eventQueue: make(chan EventWrapper, queueSize),
		handlers:   make(map[EventType][]EventHandler),
	}
	return eb
}

func (eb *EventBus) Subscribe(eventType string, h EventHandler) {
	eb.handlers[eventType] = append(eb.handlers[eventType], h)
}

// func (eb *EventBus) Publish(evt Event) {
// 	eb.eventQueue <- evt
// }

func (eb *EventBus) PublishWithWait(evt Event, wg *sync.WaitGroup) {
	eb.eventQueue <- EventWrapper{
		Event: evt,
		Wg:    wg,
	}
}

func (eb *EventBus) Start() {
	// log.Println("scheduler.Start")
	go func() {
		// log.Println("scheduler goroutine running")
		for evt := range eb.eventQueue {
			eb.dispatch(evt)
		}
	}()
}

func (eb *EventBus) dispatch(evtWrapper EventWrapper) {
	// log.Printf("Scheduler: Dispatching event")
	// eb.mu.RLock()
	evt := evtWrapper.Event
	wg := evtWrapper.Wg

	handlers := eb.handlers[evt.Type]
	// eb.mu.RUnlock()

	for _, h := range handlers {

		go func(handler EventHandler) {
			defer wg.Done()
			if err := handler(eb, evt); err != nil {
				// 错误处理（可扩展错误回调）
				log.Printf("Error handling event: %v", err)
			}
		}(h)
	}
}
