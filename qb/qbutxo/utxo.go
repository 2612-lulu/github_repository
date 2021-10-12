package qbutxo

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"qb/quantumbc"
	"qblock"
	"qbtx"
	"uss"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *quantumbc.Blockchain
}

// NewUTXOTransaction，创建普通交易
func NewUTXOTransaction(from, to, nodeID string, amount int, UTXOSet *UTXOSet) *qbtx.Transaction {
	// 需要组合输入项和输出项
	var inputs []qbtx.TXInput
	var outputs []qbtx.TXOutput

	// 查询最小utxo
	acc, validOutputs := UTXOSet.FindSpendableOutputs(from, amount)

	if acc < amount { // 如果余额不足
		fmt.Println("ERROR: Not enough funds")
		os.Exit(1)
	}

	// 构建输入项
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := qbtx.TXInput{
				Refer_tx_id:       txID,
				Refer_tx_id_index: out,
				TX_uss_sign:       uss.USSToeplitzHashSignMsg{},
				TX_src:            from,
			}
			inputs = append(inputs, input)
		}
	}

	// 构建输出项
	outputs = append(outputs, *qbtx.NewTXOutput(amount, to))
	if acc > amount { // 需要找零
		outputs = append(outputs, *qbtx.NewTXOutput(acc-amount, from)) // 需要找零
	}

	// 交易生成
	tx := qbtx.Transaction{
		TX_id:   nil,
		TX_vin:  inputs,
		TX_vout: outputs,
	}
	tx.USSTransactionSign(nodeID) // 输入项签名
	tx.TX_id = tx.SetID()
	return &tx
}

// FindSpendableOutputs，获取部分满足交易的utxo
//
// 返回值：余额int，可使用/未花费的交易map[string][]int
func (u UTXOSet) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int) // 可使用交易
	accumulated := 0                         // 记录余额
	db := u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket)) // 获取bucket
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() { // 遍历key
			txID := hex.EncodeToString(k)
			outs := qbtx.DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if address == out.TX_dst && accumulated < amount {
					accumulated += out.TX_value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// Reindex,更新UTXO
func (u UTXOSet) Reindex() {
	db := u.Blockchain.DB
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error { // 更新数据库
		err := tx.DeleteBucket(bucketName) // 删除bucket
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName) // 创建bucket
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO() // 查找未花费交易

	_ = db.Update(func(tx *bolt.Tx) error { // 更新数据库
		b := tx.Bucket(bucketName) // 获取已有bucket

		for txID, outs := range UTXO { // 遍历未花费交易并存入数据库
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.SerializeOutputs()) // 存入数据库，key：txID，value:TXOutputs
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *qblock.Block) {
	db := u.Blockchain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if !tx.IsReserveTX() {
				for _, vin := range tx.TX_vin {
					updatedOuts := qbtx.TXOutputs{}
					outsBytes := b.Get(vin.Refer_tx_id)
					outs := qbtx.DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Refer_tx_id_index {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Refer_tx_id)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.Refer_tx_id, updatedOuts.SerializeOutputs())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := qbtx.TXOutputs{}
			for _, out := range tx.TX_vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tx.TX_id, newOutputs.SerializeOutputs())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindUTXO，返回所有用户未使用的交易输出
func (u UTXOSet) FindUTXO(address string) []qbtx.TXOutput {
	var UTXOs []qbtx.TXOutput
	db := u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error { // 查看账本，启动一个只读事务
		b := tx.Bucket([]byte(utxoBucket)) // 获取bucket
		c := b.Cursor()                    //要遍历 key，我们将使用一个 Cursor

		for k, v := c.First(); k != nil; k, v = c.Next() { // 遍历
			outs := qbtx.DeserializeOutputs(v) // 反序列化交易输出
			for _, out := range outs.Outputs {
				if address == out.TX_dst {
					UTXOs = append(UTXOs, out) // 获取所有交易输出项信息
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return UTXOs
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.DB
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter
}
