package qblock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"qbtx"
	"qkdserv"
	"testing"
	"utils"
)

func TestBlock(t *testing.T) {
	fmt.Println("====================================[generate genesis block]==================================")
	qkdserv.Node_name = "P1"
	qkdserv.QKD_sign_random_matrix_pool = make(map[qkdserv.QKDSignMatrixIndex]qkdserv.QKDSignRandomsMatrix)
	qbtx.N = 16

	addresses := make([]string, 0)
	addr_table := utils.InitConfig(utils.INIT_PATH + "wallet_addr.txt")
	for addr := range addr_table { // 钱包地址
		addresses = append(addresses, addr)
	}
	genesisReservebaseData := "The Times 27/Sept/2021 Reserve is made"
	reserve_tx := qbtx.NewReserveTX(addresses, genesisReservebaseData)
	genesis := &Block{
		Version:    1.0,
		Time_stamp: 0,
		Height:     0,

		Prev_block_hash: []byte{},
		Hash:            []byte{}, // 空
		Transactions:    []*qbtx.Transaction{reserve_tx},
	}
	genesis.Hash = genesis.BlockToResolveHash()

	g := &bytes.Buffer{}
	encoder := json.NewEncoder(g)
	err := encoder.Encode(genesis)
	if err != nil {
		panic(err)
	}
	//fmt.Println(g)

	block := NewBlock([]*qbtx.Transaction{reserve_tx}, []byte{}, 1)
	b := &bytes.Buffer{}
	encoder = json.NewEncoder(b)
	err = encoder.Encode(block)
	if err != nil {
		panic(err)
	}
	fmt.Println(b)

}
