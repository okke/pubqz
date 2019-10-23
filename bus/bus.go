package bus

import "sync"

// Bus is an interface decribing basic bus operations
//
type Bus interface {
	Pub(channel string, msg Msg)
	Sub(client, channel string, handler func(Msg))
}

type msgChannel struct {

	// publish queue
	//
	in Queue

	// subscribers
	//
	out map[string]Queue
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
		channels: make(map[string]*msgChannel, 0)}
}

func (bus *bus) getChannel(channel string) *msgChannel {
	ch, found := bus.channels[channel]
	if !found {
		ch = &msgChannel{in: NewQueue(bus.handlerFor(channel)), out: make(map[string]Queue, 0)}
		bus.channels[channel] = ch
	}
	return ch
}

func (bus *bus) Pub(channel string, msg Msg) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.getChannel(channel).in.EnQueue(msg)
}

func (bus *bus) Sub(client, channel string, handler func(Msg)) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	ch := bus.getChannel(channel)
	_, found := ch.out[client]
	if found {
		// already subscribed
		// ignore
		return
	}

	ch.out[client] = NewQueue(handler)
}

func (bus *bus) handlerFor(channel string) func(Msg) {
	return func(msg Msg) {

		// get all clients interested in channel
		// and forward message
		//
		for _, q := range bus.getChannel(channel).out {
			q.EnQueue(msg)
		}
	}
}
