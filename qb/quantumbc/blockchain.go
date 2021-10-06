package quantumbc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"qblock"
	"qbtx"

	"github.com/boltdb/bolt"
)

// 数据库路径及名称
const DBFile = "/root/study/github_repository/qb/quantumbc/DB/blockchain_%s.db"

// bucket名称
const blocksBucket = "blocks"

// 创世区块留言
const genesisReservebaseData = "The Times 27/Sept/2021 Reserve is made"

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte   // 存储区块链的tail的Block的Hash，提供了一种快速找到区块链中末位Block的方式，在区块链的遍历中非常有用
	DB  *bolt.DB // blot数据库，下一层bucket（类似于数据库中的表），bucket下是键值对
}

// CreateBlockchain,创建区块链结构，初始化时只有创世区块
func CreateBlockchain(addresses []string, nodeID string) *Blockchain {
	// 定义区块链数据库名称
	dbFile := fmt.Sprintf(DBFile, nodeID)

	// 只能第一次创建，所以需要查找是否存在相应的区块链数据库文件
	if DBExists(dbFile) {
		log.Println("Blockchain already exists.")
		return nil
	} else {
		var tip []byte
		reserve_tx := qbtx.NewReserveTX(addresses, genesisReservebaseData) // 发放准备金
		genesis := qblock.NewGenesisBlock(reserve_tx)                      // 创建创世区块

		// 1.不存在区块链则创建数据库文件，0600，仅限本用户可读可写
		db, err := bolt.Open(dbFile, 0600, nil)
		if err != nil {
			log.Panic(err)
		}

		// 2.更新数据库，插入数据库数据
		err = db.Update(func(tx *bolt.Tx) error {
			// 创建bucket
			b, err := tx.CreateBucket([]byte(blocksBucket)) // bolt包函数
			if err != nil {
				log.Panic(err)
			}
			// 设置key-value
			err = b.Put(genesis.Hash, genesis.SerializeBlock())
			if err != nil {
				log.Panic(err)
			}
			// 设置key-value，存储最新区块链哈希
			err = b.Put([]byte("last"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash

			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		bc := Blockchain{tip, db} // 记录blockchain信息
		// log.Println("Create blockChain success")
		return &bc
	}
}

// 打印区块链
func PrintBlockChain(nodeID string) {
	bc := NewBlockchain(nodeID) // 1.获取当前区块链信息
	bci := bc.Iterator()        // 2.设置迭代器
	for {
		b := bci.Next()                                                                            // 3.获取当前区块信息，并变更为前一区块以迭代
		fmt.Printf("==================== Block %d ==================================\n", b.Height) // 4，打印当前区块信息
		fmt.Printf("Version: %d\n", b.Version)
		fmt.Printf("Height: %d\n", b.Height)
		fmt.Printf("TimeStamp: %d\n", b.Time_stamp)
		fmt.Printf("PrevHash: %x\n", b.Prev_block_hash)
		fmt.Printf("CurrentHash: %x\n", b.Hash)
		fmt.Printf("Transactions:\n")
		for _, tx := range b.Transactions {
			tx.PrintTransaction()
		}
		fmt.Printf("\n\n")

		if len(b.Prev_block_hash) == 0 { // 遍历数据库至创世区块
			break
		}
	}
	bc.DB.Close()
}

// NewBlockchain,读取当前区块
func NewBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(DBFile, nodeID)
	// 判断账本/数据库是否存在
	if !DBExists(dbFile) {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil) // 1.打开数据库文件
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error { // 2.更新数据库
		b := tx.Bucket([]byte(blocksBucket)) // 获取bucket
		tip = b.Get([]byte("last"))          // 获取最新区块指针。不是第一次使用，之前有块，所以此时不需要作判断
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := Blockchain{tip, db} // 记录blockchain信息
	return &bc
}

// FindUTXO，查找所有未花费的交易输出
// 参数：
// 返回值：
func (bc *Blockchain) FindUTXO() map[string]qbtx.TXOutputs {
	UTXO := make(map[string]qbtx.TXOutputs)
	spentTXOs := make(map[string][]int) // 已花费的交易，key:txID,value:Vin.Vout
	bci := bc.Iterator()                // 迭代器

	for { // 迭代区块
		block := bci.Next() // 从最后一区块逐一向前迭代

		for _, tx := range block.Transactions { // 遍历当前区块存储的交易信息
			txID := hex.EncodeToString(tx.ID) // 转换为string格式

		Outputs: // label语法，适用于多级嵌套
			for outIdx, out := range tx.Vout { // 遍历该交易信息的交易输出
				if spentTXOs[txID] != nil { // 如果交易已经被花费，直接跳过此交易
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs // continue label跳出当前该次的循环圈，立马跳到label处继续上一层的下一次循环操作
						}
					}
				}
				// 如果交易未被花费，则放入UTXO
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsReserveTX() { // 如果该交易信息不是准备金发放交易
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.Prev_block_hash) == 0 { // 迭代至创世区块，结束遍历
			break
		}
	}
	return UTXO
}

// Iterator,通过blockchain构造迭代器
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{ // 初始为最新区块
		currentHash: bc.tip,
		db:          bc.DB,
	}
	return bci
}

// AddBlock，向区块链中添加新区块
func (bc *Blockchain) AddBlock(block *qblock.Block) {
	var tip []byte
	var lastblock *qblock.Block
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("last"))
		blockData := b.Get(tip)
		lastblock = qblock.DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	if block.Height > lastblock.Height {
		err = bc.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			err := b.Put(block.Hash, block.SerializeBlock())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("last"), block.Hash)
			if err != nil {
				log.Panic(err)
			}

			bc.tip = block.Hash
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
	}
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetlastHeight() int64 {
	var lastBlock qblock.Block

	err := bc.DB.View(func(tx *bolt.Tx) error { // 查询账本
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("last"))
		blockData := b.Get(lastHash)
		lastBlock = *qblock.DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastBlock.Height
}

//
func (bc *Blockchain) GetlastHash() []byte {
	var lastBlock qblock.Block

	err := bc.DB.View(func(tx *bolt.Tx) error { // 查询账本
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("last"))
		blockData := b.Get(lastHash)
		lastBlock = *qblock.DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastBlock.Hash
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (qblock.Block, error) {
	var block qblock.Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("block is not found")
		}

		block = *qblock.DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.Prev_block_hash) == 0 {
			break
		}
	}

	return blocks
}

// 判断数据库文件/区块链是否存在
func DBExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

/*
// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (transaction.Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.New("Transaction is not found")
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*transaction.Transaction) *block.Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		// TODO: ignore transaction if it's not valid
		//if bc.VerifyTransaction(tx) != true {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		blockb := block.DeserializeBlock(blockData)

		lastHeight = blockb.Height

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := block.NewBlock(transactions, lastHash, lastHeight+1)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}
*/
