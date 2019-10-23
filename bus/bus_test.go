package bus

import (
	"testing"
	"time"
)

func TestSimplePubSub(t *testing.T) {
	t.Log("test simple pub sub")

	b := New()

	c1 := make(map[string]bool, 0)
	c2 := make(map[string]bool, 0)

	ack := make(chan bool)

	b.Sub("c1", "test", func(msg Msg) {
		c1[string(msg.Data())] = true
		ack <- true
	})

	b.Sub("c2", "test", func(msg Msg) {
		c2[string(msg.Data())] = true
		ack <- true
	})

	b.Pub("test", NewTextMsg("chipotle"))
	b.Pub("test", NewTextMsg("jalapeno"))

	// should get four acks
	//
	for i := 0; i < 4; i++ {
		select {
		case <-ack:
			// ignore
		case <-time.After(1 * time.Second):
			t.Error("reading messages took too long")
		}
	}

	if _, f := c1["chipotle"]; !f {
		t.Error("c1 did not receive chipotle")
	}
	if _, f := c2["chipotle"]; !f {
		t.Error("c2 did not receive chipotle")
	}
	if _, f := c1["jalapeno"]; !f {
		t.Error("c1 did not receive jalapeno")
	}
	if _, f := c2["jalapeno"]; !f {
		t.Error("c2 did not receive jalapeno")
	}

}
