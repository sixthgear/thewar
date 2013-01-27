package main

import "log"
import "flag"
import "github.com/sixthgear/thewar/client"

func main() {

	server := flag.Bool("server", false, "Run game as standalone server.")
	port := flag.Int("port", 11235, "Port to listen on.")
	flag.Parse()

	if *server {
		log.Printf("Starting server on port %d...\n", *port)

	} else {
		client := client.Client{}
		client.Init(false)
		client.Run()
	}

}
