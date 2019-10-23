package bus

import (
	"testing"
	"time"
)

func TestEnqueue(t *testing.T) {
	ack := make(chan bool, 4)

	q := NewQueue(func(msg Msg) error {
		ack <- true
		return nil
	})

	q.EnQueue(NewTextMsg("m1"))

	q.EnQueue(NewTextMsg("m2"))

	q.EnQueue(NewTextMsg("m3"))

	q.EnQueue(NewTextMsg("m4"))

	// should get four acks
	//
	for i := 0; i < 4; i++ {
		select {
		case <-ack:
			// ignore
		case <-time.After(3 * time.Second):
			t.Error("reading messages took too long")
			break
		}
	}

}

func TestPauseResume(t *testing.T) {
	ack := make(chan bool, 4)

	q := NewQueue(func(msg Msg) error {
		ack <- true
		return nil
	})

	q.EnQueue(NewTextMsg("m1"))
	q.EnQueue(NewTextMsg("m2"))

	// unpaused queue can be resumed
	//
	q.Resume()

	// really pause
	//
	q.Pause()

	// this pause should not have any effect since it is already paused
	//
	q.Pause()

	q.EnQueue(NewTextMsg("m3"))
	q.EnQueue(NewTextMsg("m4"))

	// really resume
	//
	q.Resume()

	// should get four acks
	//
	for i := 0; i < 4; i++ {
		select {
		case <-ack:
			// ignore
		case <-time.After(3 * time.Second):
			t.Error("reading messages took too long")
			break
		}
	}

}

func BenchmarkQueue(b *testing.B) {
	count := 0

	q := NewQueue(func(msg Msg) error {
		count = count + 1
		return nil
	})

	msg := NewTextMsg("bla")
	for i := 0; i < 100000; i++ {
		q.EnQueue(msg)
	}

}
