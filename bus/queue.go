package bus

import (
	"sync"
)

// Queue represents a in memory queue of messages
//
type Queue interface {
	EnQueue(msg Msg)
	Pause()
	Resume()
}

type queue struct {
	mutex    *sync.Mutex
	messages chan Msg
	buffer   chan<- Msg
	paused   bool
	pause    chan bool
	resume   chan bool
	handler  func(Msg)
}

// NewQueue creates a new queue
//
func NewQueue(handler func(Msg)) Queue {
	in := make(chan Msg)
	q := &queue{
		mutex:    &sync.Mutex{},
		messages: in,
		buffer:   NewBufferedMsgChannel(in, NewLLBuffer()),
		pause:    make(chan bool),
		resume:   make(chan bool),
		handler:  handler}

	go q.handle()

	return q
}

func (queue *queue) Pause() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if queue.paused {
		return
	}

	queue.paused = true
	queue.pause <- true
}

func (queue *queue) Resume() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if !queue.paused {
		return
	}

	queue.resume <- true
	queue.paused = false
}

func (queue *queue) waitForResume() {
	for {
		select {
		case <-queue.resume:
			return
		}
	}
}

func (queue *queue) handle() {
	for {
		select {
		case <-queue.pause:
			queue.waitForResume()
		case m := <-queue.messages:
			queue.handler(m)
		}
	}
}

func (queue *queue) EnQueue(msg Msg) {
	queue.buffer <- msg
}
