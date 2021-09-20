package qbwallet

import (
	"log"
	"qb/qkdserv"
	"testing"
)

func TestWallet(t *testing.T) {
	qkdserv.Node_name = "P2"
	w := NewWallet()
	log.Println("Your new address:", string(w.Addr))
	var wa Wallet
	err := wa.LoadFromFile(qkdserv.Node_name)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Your saved address:", string(wa.Addr))
}
