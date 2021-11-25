package pubqz

// Msg represents a message containing data
//
type Msg interface {
	Data() []byte
}

type msg struct {
	data []byte
}

// NewTextMsg creates a message from a string of text
//
func NewTextMsg(data string) Msg {
	return &msg{data: []byte(data)}
}

func (msg *msg) Data() []byte {
	return msg.data
}
