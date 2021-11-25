package bus

import (
	"fmt"
	"testing"
	"time"
)

func expectsNrOfAcks(t *testing.T, nr int, ack chan bool) {
	got := 0
	for i := 0; i < nr; i++ {
		select {
		case <-ack:
			got = got + 1
			// ignore
		case <-time.After(1 * time.Second):
			t.Error("expected", nr, "acks but only got", got)
		}
	}

}

func TestSimplePubSub(t *testing.T) {
	t.Log("test simple pub sub")

	b := New()

	c1 := make(map[string]bool)
	c2 := make(map[string]bool)

	ack := make(chan bool, 4)

	b.Sub("c1", "test", func(msg Msg) error {
		c1[string(msg.Data())] = true
		ack <- true
		return nil
	})

	b.Sub("c2", "test", func(msg Msg) error {
		c2[string(msg.Data())] = true
		ack <- true
		return nil
	})

	b.Pub("test", NewTextMsg("chipotle"))
	b.Pub("test", NewTextMsg("jalapeno"))

	// should get four acks
	//
	expectsNrOfAcks(t, 4, ack)

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

func TestPubSubWithMultiClients(t *testing.T) {

	b := New()

	h1 := make(map[string]bool)
	h2 := make(map[string]bool)

	ack := make(chan bool, 2)

	b.Sub("c1", "test", func(msg Msg) error {
		t.Log("c1 receives in h1", string(msg.Data()))
		h1[string(msg.Data())] = true
		ack <- true
		return nil
	})

	// subscribe same client
	//
	b.Sub("c1", "test", func(msg Msg) error {
		t.Log("c1 receives in h2", string(msg.Data()))
		h2[string(msg.Data())] = true
		ack <- true
		return nil
	})

	// first bup should be handled by first client
	//
	b.Pub("test", NewTextMsg("chipotle"))

	// second by second
	//
	b.Pub("test", NewTextMsg("jalapeno"))

	// third by first
	//
	b.Pub("test", NewTextMsg("habanero"))

	// should get 3 acks
	//
	expectsNrOfAcks(t, 3, ack)

	if _, f := h1["chipotle"]; !f {
		t.Error("h1 did not receive chipotle")
	}
	if _, f := h2["jalapeno"]; !f {
		t.Error("h2 did not receive jalapeno")
	}
	if _, f := h1["habanero"]; !f {
		t.Error("h1 did not receive habanero")
	}

}

func TestPubSubWithFailingClient(t *testing.T) {

	b := New()

	h1 := make(map[string]bool)
	h3 := make(map[string]bool)

	ack := make(chan bool, 2)

	b.Sub("c1", "test", func(msg Msg) error {
		t.Log("c1 receives in h1", string(msg.Data()))
		h1[string(msg.Data())] = true
		ack <- true
		return nil
	})

	b.Sub("c1", "test", func(msg Msg) error {
		return fmt.Errorf("I'm out")
	})

	b.Sub("c1", "test", func(msg Msg) error {
		t.Log("c1 receives in h3", string(msg.Data()))
		h3[string(msg.Data())] = true
		ack <- true
		return nil
	})

	b.Pub("test", NewTextMsg("chipotle"))
	b.Pub("test", NewTextMsg("jalapeno"))
	b.Pub("test", NewTextMsg("habanero"))
	b.Pub("test", NewTextMsg("tabasco"))

	// should get 3 acks
	//
	expectsNrOfAcks(t, 3, ack)

	if _, f := h1["chipotle"]; !f {
		t.Error("h1 did not receive chipotle")
	}
	if _, f := h3["jalapeno"]; !f {
		t.Error("h3 did not receive jalapeno")
	}
	if _, f := h1["habanero"]; !f {
		t.Error("h1 did not receive habanero")
	}
	if _, f := h3["tabasco"]; !f {
		t.Error("h3 did not receive tabasco")
	}

}

func TestPubSubWithOnlyFailingClients(t *testing.T) {

	b := New()

	b.Sub("c1", "test", func(msg Msg) error {
		return fmt.Errorf("I'm out")
	})

	b.Sub("c1", "test", func(msg Msg) error {
		return fmt.Errorf("me too")
	})

	b.Pub("test", NewTextMsg("chipotle"))
	b.Pub("test", NewTextMsg("jalapeno"))
	b.Pub("test", NewTextMsg("habanero"))
	b.Pub("test", NewTextMsg("tabasco"))

	// TODO mechanism that waits until bus has paused channel for this client
	//
	<-time.After(1 * time.Second)

	if !b.IsPaused("c1", "test") {
		t.Error("expected bus to be paused")
	}

	ack := make(chan bool, 4)

	b.Sub("c1", "test", func(msg Msg) error {
		ack <- true
		return nil
	})

	// and expect 4 acks for new handler
	//
	expectsNrOfAcks(t, 4, ack)

}
