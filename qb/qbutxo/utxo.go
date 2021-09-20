package qbutxo

import (
	"encoding/hex"
	"log"
	"qb/qblock"
	"qb/qbtx"
	"qb/qbwallet"
	"qb/qkdserv"
	"qb/quantumbc"
	"qb/uss"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *quantumbc.Blockchain
}

// NewUTXOTransaction，创建普通交易
func NewUTXOTransaction(wa *qbwallet.Wallet, to string, amount int, UTXOSet *UTXOSet) *qbtx.Transaction {
	// 需要组合输入项和输出项
	var inputs []qbtx.TXInput
	var outputs []qbtx.TXOutput

	// 查询最小utxo
	acc, validOutputs := UTXOSet.FindSpendableOutputs(wa.Node_id[:], amount)

	if acc < amount { // 如果余额不足
		log.Panic("ERROR: Not enough funds")
	}

	// 构建输入项
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := qbtx.TXInput{
				Txid:      txID,
				Vout:      out,
				Signature: uss.USSToeplitzHashSignMsg{},
				From:      qkdserv.Node_name,
			}
			inputs = append(inputs, input)
		}
	}

	// 构建输出项
	//from := string(wa.GetAddress())
	outputs = append(outputs, *qbtx.NewTXOutput(amount, to))
	if acc > amount { // 需要找零
		outputs = append(outputs, *qbtx.NewTXOutput(acc-amount, qkdserv.Node_name)) // a change
	}

	// 交易生成
	tx := qbtx.Transaction{
		ID:   nil,
		Vin:  inputs,
		Vout: outputs,
	}
	tx.ID = tx.SetID()
	tx.SignTX(qkdserv.Node_name)
	//UTXOSet.Blockchain.SignTransaction(&tx, wa.PrivateKey)
	log.Println("create a new utxo tx")
	return &tx
}

// FindSpendableOutputs，获取部分满足交易的utxo
//
// 返回值：余额int，可使用/未花费的交易map[string][]int
func (u UTXOSet) FindSpendableOutputs(node_id []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int) // 可使用交易
	accumulated := 0                         // 记录余额
	db := u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := qbtx.DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if accumulated < amount {
					accumulated += out.Value
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

// Reindex,
func (u UTXOSet) Reindex() {
	db := u.Blockchain.DB
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.SerializeOutputs())
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
				for _, vin := range tx.Vin {
					updatedOuts := qbtx.TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs := qbtx.DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.Txid, updatedOuts.SerializeOutputs())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := qbtx.TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tx.ID, newOutputs.SerializeOutputs())
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
