package main

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"go.step.sm/crypto/randutil"
)

const (
	host                   = "localhost"
	port                   = "6969"
	buffer_size            = 64
	max_message_per_client = 10
	random_id_len          = 8
)

type Client struct {
	Name     string
	Messages []string
	Socket   net.Conn
}

var Clients map[string]*Client

func client_handler(socket net.Conn, ch chan<- [buffer_size]byte) {
	defer socket.Close()

	id, err := randutil.Alphanumeric(random_id_len)
	if err != nil {
		log.Println("Could not generate identifier!")
		return
	}

	client := &Client{
		Name:     fmt.Sprintf("user_%s", id),
		Messages: []string{},
		Socket:   socket,
	}

	Clients[id] = client

	reader := bufio.NewReader(socket)
	for {
		socket.Write([]byte(client.Name + "# ")) // echo
		var chunk [buffer_size]byte
		message, err := reader.ReadString('\n')

		if err != nil {
			log.Println("Failed to read bytes")
			return
		}

		if len(client.Messages) < max_message_per_client {
			client.Messages = append(client.Messages[:], message)

			for i := 0; i < len(message) && i < 64; i++ {
				chunk[i] = message[i]
			}
			ch <- chunk

			go broadcast_message(client, chunk)
		} else {
			client.Messages = client.Messages[:0]
		}
	}
}

func broadcast_message(client *Client, buff [buffer_size]byte) {
	for _, c := range Clients {
		if c != client {
			c.Socket.Write([]byte("\n"))
			c.Socket.Write([]byte(client.Name + "# " + string(buff[:])))
			c.Socket.Write([]byte(c.Name + "# "))
		}
	}
}

func server_handler(ch <-chan [buffer_size]byte) {
	for chunk := range ch {
		log.Println(chunk)
	}
}

func main() {

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))

	if err != nil {
		log.Fatalln(err)
	}

	Clients = map[string]*Client{}
	ch := make(chan [buffer_size]byte)

	go server_handler(ch)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatalln(err)
		}

		go client_handler(conn, ch)
	}

}
