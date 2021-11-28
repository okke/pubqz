package pubqz

import "testing"

func TestNewMsg(t *testing.T) {

	msg := NewMsg([]byte("chipotle"))
	if string(msg.Data()) != "chipotle" {
		t.Error("expected chipotle")
	}
}

func TestNewTextMsg(t *testing.T) {

	msg := NewTextMsg("chipotle")
	if string(msg.Data()) != "chipotle" {
		t.Error("expected chipotle")
	}
}
