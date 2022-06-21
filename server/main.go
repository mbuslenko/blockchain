package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	port := flag.Uint("port", 5655, "TCP Port Number For Blockchain Server")
	flag.Parse()

	app := NewServer(uint16(*port))
	app.Run()
}
