package quantumbc

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"qb/qblock"
	"qb/qbtx"

	"github.com/boltdb/bolt"
)

// 数据库路径及名称
const dbFile = "/root/study/github_repository/qb/quantumbc/DB/blockchain_%s.db"

// bucket名称
const blocksBucket = "blocks"

// 创世区块留言
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte   // 存储区块链的tail的Block的Hash，提供了一种快速找到区块链中末位Block的方式，在区块链的遍历中非常有用
	DB  *bolt.DB // blot数据库，下一层bucket（类似于数据库中的表），bucket下是键值对
}

// CreateBlockchain,创建区块链结构，初始化时只有创世区块
func CreateBlockchain(address, nodeID string) *Blockchain {
	// 定义区块链数据库名称
	dbFile := fmt.Sprintf(dbFile, nodeID)

	// 只能第一次创建，所以需要查找是否存在相应的区块链数据库文件
	if dbExists(dbFile) {
		log.Println("Blockchain already exists.")
		return nil
	}

	var tip []byte

	reserve_tx := qbtx.NewReserveTX(address, genesisCoinbaseData) // 铸币交易
	genesis := qblock.NewGenesisBlock(reserve_tx)                 // 创建创世区块

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

		// 设置key-value
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

	bc := Blockchain{tip, db}
	log.Println("Create blockChain success")

	// 记录blockchain信息
	return &bc
}

// NewBlockchain,读取当前区块
func NewBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if !dbExists(dbFile) {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	// 打开数据库文件
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	// 更新数据库
	err = db.Update(func(tx *bolt.Tx) error {
		// 获取bucket
		b := tx.Bucket([]byte(blocksBucket))
		// 不是第一次使用，之前有块，所以此时不需要作判断
		tip = b.Get([]byte("last"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// 记录blockchain信息
	bc := Blockchain{tip, db}

	return &bc
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]qbtx.TXOutputs {
	UTXO := make(map[string]qbtx.TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsReserveTX() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.Prev_block_hash) == 0 {
			break
		}
	}

	return UTXO
}

// Iterator,通过blockchain构造迭代器
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{
		currentHash: bc.tip,
		db:          bc.DB,
	}
	return bci
}

/*
// AddBlock，向区块链中添加新区块
func (bc *Blockchain) AddBlock(blockb *block.Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(blockb.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := blockb.Serialize()
		err := b.Put(blockb.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := block.DeserializeBlock(lastBlockData)

		if blockb.Height > lastBlock.Height {
			err = b.Put([]byte("l"), blockb.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = blockb.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

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




// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock block.Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *block.DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (block.Block, error) {
	var blockb block.Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		blockb = *block.DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return blockb, err
	}

	return blockb, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
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

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *qbtx.Transaction) {
	prevTXs := make(map[string]qbtx.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
*/
// 判断数据库文件/区块链是否存在
func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
