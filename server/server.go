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
	TURN_TICKS = 12
	M_WIDTH    = 64
	M_DEPTH    = 64
)

var (
	port        int
	world       *Map
	running     bool
	pathCache   map[int][]int
	channel     = make(chan *Order)
	connections = make([]net.Conn, 0)
)

func main() {

	port := flag.Int("port", 11235, "Port to listen on.")
	flag.Parse()

	log.Printf("Starting server on port %d...\n", *port)

	pathCache = make(map[int][]int, 32)
	running = true
	world = new(Map)
	world.Init(M_WIDTH, M_DEPTH)
	world.Generate()
	GenerateObjects(world)

	// json := world.Encode()
	// fmt.Printf("%s\n", json)

	run()
}

func dispatchOrders() {
	for {
		order := <-channel
		for i := range connections {
			fmt.Fprintf(connections[i], "%s\n", order.Encode())
		}
	}
}

func handleConnection(c net.Conn) {

	a := c.RemoteAddr().String()

	for {

		data, _ := bufio.NewReader(c).ReadBytes('\n')

		if len(data) == 0 {
			log.Println("Client closed connection ", a)
			c.Close()
			return
		}

		order := new(Order).Decode(data)
		log.Printf("%s -> %s", a, data)
		switch order.Order {
		case OR_REQMAP:
			fmt.Fprintf(c, "%s\n", world.Encode())
		case OR_MOVE:
			channel <- order
		default:
			log.Println("Unhandled order: ", order.Order)
		}

		runtime.Gosched()
	}
}

func run() {

	go dispatchOrders()

	ln, err := net.Listen("tcp", ":11235")
	if err != nil {
		// handle error
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}

		connections = append(connections, c)
		go handleConnection(c)
	}

	// t := 0.0
	// const dt = 1.0 / 60
	// currentTime := float64(time.Now().UnixNano()) / 1000000000
	// accumulator := 0.0

	// for running {

	// 	newTime := float64(time.Now().UnixNano()) / 1000000000
	// 	frameTime := newTime - currentTime
	// 	currentTime = newTime
	// 	accumulator += frameTime

	// 	for accumulator >= dt {

	// 		// TODO read from sockets

	// 		update(dt) // update

	// 		accumulator -= dt
	// 		t += dt
	// 		// write to sockets
	// 	}
	// 	// renderer.Render(world)
	// }

}
