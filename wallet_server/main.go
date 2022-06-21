package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Wallet Server: ")
}

func main() {
	port := flag.Uint("port", 9657, "TCP Port Number For Wallet Server")
	gateway := flag.String("gateway", "http://127.0.0.1:5655", "Blockhain Gateway")
	flag.Parse()

	app := NewWalletServer(uint16(*port), *gateway)
	app.Run()
}
