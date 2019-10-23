package bus

import "testing"

func TestNewTextMsg(t *testing.T) {

	msg := NewTextMsg("chipotle")
	if string(msg.Data()) != "chipotle" {
		t.Error("expected chipotle")
	}
}
