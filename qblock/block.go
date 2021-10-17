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
	"qkdserv"
	"time"
	"uss"
	"utils"
)

// 区块包含的最小交易数量
const BLOCK_LENGTH = 1

// 区块结构
type Block struct {
	Version    int64 `json:"Version"`   // 当前版本
	Time_stamp int64 `json:"Timestamp"` // 系统当前时间
	Height     int64 `json:"Height"`    // 区块高度

	Prev_block_hash []byte              `json:"Prevblockhash"` // 前一区块hash值
	Hash            []byte              `json:"Currentblockhash"`
	Transactions    []*qbtx.Transaction `json:"Transactions"` // 用于共识的交易信息
	Block_uss       uss.USSToeplitzHashSignMsg
}

// NewBlock，生成新区块
// 参数：交易[]*qbtx.Transaction，前一区块hashprevBlockHash，高度值int64
// 返回值：新区块*Block
func NewBlock(transactions []*qbtx.Transaction, prevBlockHash []byte, height int64) *Block {
	block := Block{
		Version:    1.0,
		Time_stamp: time.Now().Unix(),
		Height:     height,

		Prev_block_hash: prevBlockHash,
		Hash:            []byte{}, // 空
		Transactions:    transactions,
		Block_uss: uss.USSToeplitzHashSignMsg{ // 签名信息
			Sign_index: qkdserv.QKDSignMatrixIndex{ // 签名索引
				Sign_dev_id:  utils.GetNodeID(qkdserv.Node_name), // 签名者ID
				Sign_task_sn: uss.GenSignTaskSN(16),              // 签名序列号
			},
			Main_row_num: qkdserv.QKDSignRandomMainRowNum{
				Sign_node_name:    qkdserv.Node_name, // 签名者节点号
				Main_row_num:      0,                 // 签名主行号，签名时默认为0
				Random_row_counts: qbtx.N - 1,
				Random_unit_len:   16,
			},
			USS_counts:   qbtx.N - 1, // 验签者的数量
			USS_unit_len: 16,         // 签名的单位长度，一般默认为16
		},
	}
	block.Hash = block.BlockToResolveHash() // 生成当前区块hash值
	block.Block_uss.USS_message = block.Hash
	block.Block_uss = uss.UnconditionallySecureSign(block.Block_uss.Sign_index,
		block.Block_uss.USS_counts, block.Block_uss.USS_unit_len,
		block.Block_uss.USS_message)

	return &block
}

// NewGenesisBlock，创建创世区块。实际操作是读取固定区块以保证各节点有相同的账本
// 参数：准备金交易
// 返回值：创世区块
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
// 参数：区块
// 返回值：该区块hash值
func (b *Block) BlockToResolveHash() []byte {
	data := b.prepareData()
	hash := sha256.Sum256(data) // 取区块hash
	return hash[:]
}

// prepareData，数据准备函数，利用join完成字节切片的组合
// 参数：区块
// 返回值：待hash的区块消息
func (b *Block) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			utils.IntToHex(b.Version),
			utils.IntToHex(b.Time_stamp),
			utils.IntToHex(b.Height),
			b.Prev_block_hash,
			b.HashTransactions(), // 默克尔树根节点
		},
		[]byte{},
	)

	return data
}

// HashTransactions，构建区块交易hash值，实现一种交易转[]byte的方法
// 参数：区块
// 返回值：区块中交易信息的hash值
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.SerializeTX())
	}
	mTree := merkletree.NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

// SerializeBlock，区块序列化
// 参数：待序列化的区块结构
// 返回值：序列化结果
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
// 参数：序列化结果
// 返回值：区块
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d)) // 创建解码器
	err := decoder.Decode(&block)                 // 解析区块数据
	if err != nil {
		log.Println(err)
	}
	return &block
}
