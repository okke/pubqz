package pubqz

import "container/list"

// MsgBuffer is a buffer for messages
//
type MsgBuffer interface {
	Len() int
	Front() Msg
	Add(msg Msg)
	DeleteFirst()
}

type llBuffer struct {
	list *list.List
}

func (llBuffer *llBuffer) Len() int {
	return llBuffer.list.Len()
}

func (llBuffer *llBuffer) Front() Msg {
	return llBuffer.list.Front().Value.(Msg)
}

func (llBuffer *llBuffer) Add(msg Msg) {
	llBuffer.list.PushBack(msg)
}

func (llBuffer *llBuffer) DeleteFirst() {
	llBuffer.list.Remove(llBuffer.list.Front())
}

// NewLLBuffer creates a message buffer based on a linked list implementation
//
func NewLLBuffer() MsgBuffer {
	return &llBuffer{list: list.New()}
}

// NewBufferedMsgChannel will create a channel that does not block on writes
// Thanks to https://rogpeppe.wordpress.com/2010/02/10/unlimited-buffering-with-low-overhead/
//
func NewBufferedMsgChannel(out chan<- Msg, buf MsgBuffer) chan<- Msg {
	in := make(chan Msg)
	go func() {
		for {
			outc := out
			var v Msg
			n := buf.Len()
			if n == 0 {
				// buffer empty: don't try to send on output
				if in == nil {
					close(out)
					return
				}
				outc = nil
			} else {
				v = buf.Front()
			}
			select {
			case e, ok := <-in:
				if !ok {
					in = nil
				} else {
					buf.Add(e)
				}
			case outc <- v:
				buf.DeleteFirst()
			}
		}
	}()
	return in
}
