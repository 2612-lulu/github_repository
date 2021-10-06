package qbwallet

import (
	"fmt"
	"log"
	"qkdserv"
	"testing"
)

func TestWallet(t *testing.T) {
	qkdserv.Node_name = "P2"
	w := NewWallet(qkdserv.Node_name)
	fmt.Println("Your new address:", string(w.Addr))
	if ValidateAddress(string(w.Addr)) {
		fmt.Println("validate success")
	}
	var wa Wallet
	err := wa.LoadFromFile(qkdserv.Node_name)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Your saved address:", string(wa.Addr))

}
