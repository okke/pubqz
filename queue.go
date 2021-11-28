package pubqz

// Queue represents a in memory queue of messages
//
type Queue interface {
	EnQueue(msg Msg)
	Pause()
	IsPaused() bool
	Resume()
}

type queue struct {
	messages  chan Msg
	buffer    chan<- Msg
	failedMsg []Msg
	paused    bool
	pause     chan bool
	resume    chan bool
	handler   func(Msg) error
}

// NewQueue creates a new queue
//
func NewQueue(handler func(Msg) error) Queue {
	in := make(chan Msg)
	q := &queue{
		messages:  in,
		buffer:    NewBufferedMsgChannel(in, NewLLBuffer()),
		failedMsg: []Msg{},
		pause:     make(chan bool, 2),
		resume:    make(chan bool, 2),
		handler:   handler}

	go q.handle()

	return q
}

func (queue *queue) fail(msg Msg) {

	queue.failedMsg = append(queue.failedMsg, msg)
	queue.Pause()
}

func (queue *queue) IsPaused() bool {
	return queue.paused
}

func (queue *queue) Pause() {

	if queue.paused {
		return
	}

	queue.paused = true
	queue.pause <- true
}

func (queue *queue) Resume() {

	if queue.failedMsg != nil {

		for _, msg := range queue.failedMsg {
			queue.handleMsg(msg)
		}
		queue.failedMsg = []Msg{}
	}

	if !queue.paused {
		return
	}

	queue.resume <- true

	queue.paused = false
}

func (queue *queue) waitForResume() {
	for {
		if <-queue.resume {
			return
		}
	}
}

func (queue *queue) handleMsg(msg Msg) {
	if err := queue.handler(msg); err != nil {
		queue.fail(msg)
	}
}

func (queue *queue) handle() {
	for {
		select {
		case <-queue.pause:
			queue.waitForResume()
		case m := <-queue.messages:
			queue.handleMsg(m)
		}
	}
}

func (queue *queue) EnQueue(msg Msg) {
	queue.buffer <- msg
}
