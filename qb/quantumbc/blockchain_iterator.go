package quantumbc

import (
	"log"
	"qb/qblock"

	"github.com/boltdb/bolt"
)

// BlockchainIterator，迭代器
type BlockchainIterator struct {
	currentHash []byte   // 当前区块hash
	db          *bolt.DB // 已经打开的数据库
}

// Next,获取前块数据，返回当前区块的数据
func (i *BlockchainIterator) Next() *qblock.Block {
	var b *qblock.Block
	// 根据hash获取块数据
	err := i.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		encodedBlock := bucket.Get(i.currentHash)
		// 解码当前块数据
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
