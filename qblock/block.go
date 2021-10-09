// block包，定义区块链数据结构，实现区块链打包
// 创建人：zhanglu
// 创建时间：2021/08/25
package qblock

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"log"
	"merkletree"
	"os"
	"qbtx"
	"time"
	"utils"
)

// 区块包含的最小交易数量
const BLOCK_LENGTH = 1

// 区块结构
type Block struct {
	Version    int64 `json:"version"`   // 当前版本
	Time_stamp int64 `json:"timestamp"` // 系统当前时间
	Height     int64 `json:"height"`    // 区块高度

	Prev_block_hash []byte              `json:"prevblockhash"` // 前一区块hash值
	Hash            []byte              `json:"currentblockhash"`
	Transactions    []*qbtx.Transaction `json:"transactions"` // 用于共识的交易信息
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
	return block
}

// NewGenesisBlock，创建创世区块
func NewGenesisBlock(reserve *qbtx.Transaction) *Block {
	file, _ := os.Open("../config/genesisblock.json") // 打开文件
	defer file.Close()                                // 关闭文件
	// NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
	decoder := json.NewDecoder(file)
	var block Block
	err := decoder.Decode(&block) //Decode从输入流读取下一个json编码值并保存在v指向的值里
	if err != nil {
		panic(err)
	}
	return &block
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
			utils.IntToHex(b.Version),
			utils.IntToHex(b.Time_stamp),
			utils.IntToHex(b.Height),
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
		log.Println(err)
	}

	return &block
}
