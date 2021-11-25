package bus

import (
	"fmt"
	"sync"
)

// Bus is an interface decribing basic bus operations
//
type Bus interface {
	Pub(channel string, msg Msg)
	Sub(client, channel string, handler func(Msg) error)
	IsPaused(client, channel string) bool
}

type clientHandler struct {
	name     string
	index    int
	handlers []func(Msg) error
}

func newClientHandler(name string) *clientHandler {
	return &clientHandler{name: name, handlers: make([]func(Msg) error,0)}
}

func (clientHandler *clientHandler) handle(msg Msg) error {
	handled := false

	for !handled {
		if len(clientHandler.handlers) == 0 {
			return fmt.Errorf("no handlers to handle message for %s", clientHandler.name)
		}
		if clientHandler.index >= len(clientHandler.handlers) {
			clientHandler.index = 0
		}
		err := clientHandler.handlers[clientHandler.index](msg)

		if err != nil {
			// handler failure, do not trust this handler anymore so remove it
			//
			clientHandler.handlers = append(clientHandler.handlers[:clientHandler.index], clientHandler.handlers[clientHandler.index+1:]...)

		} else {
			handled = true
			clientHandler.index = clientHandler.index + 1
		}
	}

	return nil
}

func (clientHandler *clientHandler) add(handler func(Msg) error) {
	clientHandler.handlers = append(clientHandler.handlers, handler)
}

type msgChannel struct {

	// publish queue
	//
	in Queue

	// subscribers (queues for clients)
	//
	out map[string]Queue

	// handlers
	handlers map[string]*clientHandler
}

type bus struct {
	mutex    *sync.Mutex
	channels map[string]*msgChannel
}

// New constructs a new message bus
//
func New() Bus {
	return &bus{
		mutex:    &sync.Mutex{},
		channels: make(map[string]*msgChannel)}
}

func (bus *bus) getChannel(channel string) *msgChannel {
	ch, found := bus.channels[channel]
	if !found {
		ch = &msgChannel{
			in:       NewQueue(bus.handlerFor(channel)),
			out:      make(map[string]Queue),
			handlers: make(map[string]*clientHandler)}
		bus.channels[channel] = ch
	}
	return ch
}

func (bus *bus) Pub(channel string, msg Msg) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.getChannel(channel).in.EnQueue(msg)
}

func (bus *bus) IsPaused(client, channel string) bool {
	ch := bus.getChannel(channel)
	_, found := ch.out[client]
	if found {
		return ch.out[client].IsPaused()
	}
	return false
}

func (bus *bus) Sub(client, channel string, handler func(Msg) error) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	ch := bus.getChannel(channel)
	_, found := ch.out[client]
	if found {
		clientHandler := ch.handlers[client]
		clientHandler.add(handler)

		// new handler so in case the queue was paused, we should try to resume
		//
		ch.out[client].Resume()

		return
	}

	clientHandler := newClientHandler(client + ":" + channel)
	clientHandler.add(handler)

	ch.handlers[client] = clientHandler
	ch.out[client] = NewQueue(clientHandler.handle)

}

func (bus *bus) handlerFor(channel string) func(Msg) error {
	return func(msg Msg) error {

		// get all clients interested in channel
		// and forward message
		//
		for _, q := range bus.getChannel(channel).out {
			q.EnQueue(msg)
		}

		return nil
	}
}
