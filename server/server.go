package main

import (
	"bufio"
	"flag"
	"fmt"
	. "github.com/sixthgear/thewar/gamelib"
	"log"
	"net"
	"runtime"
)

const (
	M_WIDTH = 64
	M_DEPTH = 64
)

var (
	port        int
	world       *Map
	running     bool
	channel     = make(chan *Order)
	connections = make(map[net.Addr]net.Conn)
)

func main() {

	port := flag.Int("port", 11235, "Port to listen on.")
	flag.Parse()
	log.Printf("Starting server on port %d...\n", *port)

	running = true
	world = new(Map)
	world.Init(M_WIDTH, M_DEPTH)
	world.Generate()
	GenerateObjects(world)

	run()
}

func listen() {
	// block until a new connection arrives
	ln, err := net.Listen("tcp", ":11235")
	if err != nil {
		log.Fatal("Unable to open socket. ", err)
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal("Error accepting connection. ", err)
			continue
		}

		go handlePlayer(c)
		runtime.Gosched()
	}
}

func handlePlayer(c net.Conn) {

	a := c.RemoteAddr().String()

	for {

		// read a line from the socket
		data, err := bufio.NewReader(c).ReadBytes('\n')

		if err != nil {
			switch err.Error() {
			case "EOF":
				log.Println(a, "-> Client closed connection.")
				delete(connections, c.RemoteAddr())
				c.Close()
				return
			}
		}

		// decode an order
		order, err := new(Order).Decode(data)
		switch order.Order {
		case OR_REQMAP:
			// transmit complete world to client
			fmt.Fprintf(c, "%s\n", world.Encode())
			// HACK: do not add player to connections until we've sent the map
			connections[c.RemoteAddr()] = c
		case OR_MOVE:
			// standard order
			channel <- order
		default:
			log.Println(a, " -> Unhandled order: ", order.Order)
		}

		log.Printf("%s -> %s", a, data)
		runtime.Gosched()
	}
}

func run() {

	go listen()

	for {
		order := <-channel
		// dispatch order to all connections
		for i := range connections {
			fmt.Fprintf(connections[i], "%s\n", order.Encode())
		}

		// TODO actual simulation!
	}

}
