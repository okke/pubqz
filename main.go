package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/okke/pubqz/bus"
)

type server struct {
	port string
	bus  bus.Bus
}

type Server interface {
	Listen()
}

func New(port string) Server {
	return &server{port: port, bus: bus.New()}
}

func (server *server) Listen() {

	l, err := net.Listen("tcp4", ":"+server.port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go server.handleConnection(c)
	}
}

func (server *server) handleConnection(c net.Conn) {
	defer c.Close()

	reader := bufio.NewReader(c)
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		netData, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		netSplitted := strings.Split(netData, " ")
		if len(netSplitted) == 0 {
			continue
		}

		cmd := strings.TrimSpace(netSplitted[0])

		fmt.Printf("CMD FROM %s:%s\n", c.RemoteAddr().String(), cmd)

		switch cmd {
		case "PUB":
			if len(netSplitted) < 3 {
				c.Write([]byte(string("ERROR not enough arguments\n")))
			} else {
				channel := strings.TrimSpace(netSplitted[1])
				msg := strings.TrimSpace(netSplitted[2])
				server.bus.Pub(channel, bus.NewTextMsg(msg))
			}
		case "SUB":
			if len(netSplitted) < 3 {
				c.Write([]byte(string("ERROR not enough arguments\n")))
			} else {
				client := strings.TrimSpace(netSplitted[1])
				channel := strings.TrimSpace(netSplitted[2])
				server.bus.Sub(client, channel, func(msg bus.Msg) error {
					_, err := c.Write([]byte(string(fmt.Sprintf("DATA %s\n", string(msg.Data())))))
					return err
				})
			}
		}

	}
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	New(arguments[1]).Listen()

}
