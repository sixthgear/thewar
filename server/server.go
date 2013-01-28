package main

import (
	"flag"
	"fmt"
	. "github.com/sixthgear/thewar/gamelib"
	"log"
	"net"
	// "time"
)

const (
	TURN_TICKS = 12
	M_WIDTH    = 64
	M_DEPTH    = 64
)

var (
	port      int
	world     *Map
	running   bool
	pathCache map[int][]int
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

func handleConnection(c net.Conn) {
	fmt.Fprintf(c, "%s\n", world.Encode())
	log.Println("Served map data to ", c.RemoteAddr().String())
	c.Close()
}

func run() {

	ln, err := net.Listen("tcp", ":11235")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}

		go handleConnection(conn)
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
