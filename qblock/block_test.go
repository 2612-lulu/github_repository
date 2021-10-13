package qblock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"qbtx"
	"testing"
	"utils"
)

func TestBlock(t *testing.T) {
	fmt.Println("====================================[generate genesis block]==================================")
	addresses := make([]string, 0)
	addr_table := utils.InitConfig(utils.INIT_PATH + "wallet_addr.txt")
	for addr, _ := range addr_table { // 钱包地址
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
		Transactions:    []qbtx.Transaction{reserve_tx},
	}
	genesis.Hash = genesis.BlockToResolveHash()

	b := &bytes.Buffer{}
	encoder := json.NewEncoder(b)
	err := encoder.Encode(genesis)
	if err != nil {
		panic(err)
	}
	fmt.Println(b)

}
