package main

import (
	"crypto-blockchain/wallet"
	"fmt"
	"log"
)

func init() {
	log.SetPrefix("[BC] ")
}

func main() {
	w := wallet.NewWallet()
	fmt.Println(w.Address())

	t := wallet.NewTransaction(w.PrivateKey(), w.PublicKey(), w.Address(), "B", 1.0)
	fmt.Printf("signature %s\n", t.GenerateSignature())
}
