package main

import (
	"bufio"
	"flag"
	"fmt"
	. "github.com/sixthgear/thewar/gamelib"
	"log"
	"net"
	"runtime"
	"time"
)

const (
	M_WIDTH = 64
	M_DEPTH = 64
)

var (
	port        int
	world       *Map
	running     bool
	connections map[net.Addr]net.Conn
)

func main() {

	port := flag.Int("port", 11235, "Port to listen on.")
	flag.Parse()
	log.Printf("Starting server on port %d...\n", *port)

	world = new(Map).Generate(M_WIDTH, M_DEPTH)
	GenerateObjects(world)

	run()
}

func communicate(channel chan *Order, c net.Conn) {

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

func listen(channel chan *Order) {
	// block until a new connection arrives
	ln, err := net.Listen("tcp", ":11235")
	if err != nil {
		log.Fatal("Unable to open socket. ", err)
	}
	for {
		if c, err := ln.Accept(); err == nil {
			go communicate(channel, c)
		} else {
			log.Fatal("Error accepting connection. ", err)
			continue
		}

		runtime.Gosched()
	}
}

func run() {

	// set up network IO Loop
	connections = make(map[net.Addr]net.Conn)
	channel := make(chan *Order)
	go listen(channel)

	running = true

	for {

		var o *Order
		select {
		case o = <-channel:
			// dispatch order to all connections
			for i := range connections {
				fmt.Fprintf(connections[i], "%s\n", o.Encode())
			}
		default:
			time.Sleep(100 * time.Millisecond)
			// runtime.Gosched()
		}

		// order := <-channel
		// TODO actual simulation!
	}

}
