package quantumbc

import (
	"log"
	"qblock"

	"github.com/boltdb/bolt"
)

// BlockchainIterator，迭代器
type BlockchainIterator struct {
	currentHash []byte   // 当前区块hash
	db          *bolt.DB // 已经打开的数据库
}

// Next,获取当前区块
func (i *BlockchainIterator) Next() *qblock.Block {
	var b *qblock.Block
	// 根据hash获取块数据
	err := i.db.View(func(tx *bolt.Tx) error { // 查看数据库
		bucket := tx.Bucket([]byte(blocksBucket)) // 获取已有bucket
		encodedBlock := bucket.Get(i.currentHash) // 获取key-value
		// 解码当前块数据,获取区块
		b = qblock.DeserializeBlock(encodedBlock)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	// 当前块变更为前块hash
	i.currentHash = b.Prev_block_hash
	// 返回区块
	return b
}
