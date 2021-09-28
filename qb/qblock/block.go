// block包，定义区块链数据结构，实现区块链打包
// 创建人：zhanglu
// 创建时间：2021/08/25
package qblock

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"qb/merkletree"
	"qb/qbtools"
	"qb/qbtx"
	"time"
)

// 区块包含的最小交易数量
const BLOCK_LENGTH = 1

// 区块结构
type Block struct {
	Version    int64 // 当前版本
	Time_stamp int64 // 系统当前时间
	Height     int64 // 区块高度

	Prev_block_hash []byte // 前一区块hash值
	Hash            []byte
	Transactions    []*qbtx.Transaction // 用于共识的交易信息
}

// NewBlock，生成新区块
func NewBlock(transactions []*qbtx.Transaction, prevBlockHash []byte, height int64) *Block {
	block := &Block{
		Version:    1.0,
		Time_stamp: time.Now().Unix(),
		Height:     height,

		Prev_block_hash: prevBlockHash,
		Hash:            []byte{}, // 空
		Transactions:    transactions,
	}
	block.Hash = block.BlockToResolveHash() // 生成当前区块hash值
	//fmt.Println("hash:=", block.Hash)
	return block
}

// NewGenesisBlock，创建创世区块
func NewGenesisBlock(reserve *qbtx.Transaction) *Block {
	//log.Println("create a new genesis block")
	return NewBlock([]*qbtx.Transaction{reserve}, []byte{}, 0)
}

// NewBlock，生成新区块hash值
func (b *Block) BlockToResolveHash() []byte {
	data := b.prepareData()
	hash := sha256.Sum256(data) // 取区块hash
	return hash[:]
}

// 数据准备函数，利用join完成字节切片的组合
func (b *Block) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			qbtools.IntToHex(b.Version),
			qbtools.IntToHex(b.Time_stamp),
			qbtools.IntToHex(b.Height),
			b.Prev_block_hash,
			b.HashTransactions(),
		},
		[]byte{},
	)

	return data
}

// HashTransactions，构建区块交易hash值，实现一种交易转[]byte的方法
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.SerializeTX())
	}
	mTree := merkletree.NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

// SerializeBlock，区块序列化
func (b *Block) SerializeBlock() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result) // 生成编码器encoder

	err := encoder.Encode(b) //编码
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock，区块反序列化
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d)) // 创建解码器
	err := decoder.Decode(&block)                 // 解析区块数据
	if err != nil {
		log.Panic(err)
	}

	return &block
}