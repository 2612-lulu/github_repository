package qbnode

import (
	"encoding/hex"
	"fmt"
	"log"
	"qb/quantumbc"
	"qblock"
	"qbtx"
	"utils"
)

// 线程2：blockMsg，用于打包收到的交易信息
func (node *Node) blockMsg() {
	for {
		select {
		case msg := <-node.MsgBlock: // 从调度器通道中获取缓存信息
			err := node.startTopbft(msg)
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		case <-node.Block_clock:
			err := node.blockWhenClock()
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		}
	}
}

// node.startTopbft,将收到的交易信息存储到数组
// 参数：收到的消息
// 返回值：处理错误error，默认为nil
func (node *Node) startTopbft(msg interface{}) error {
	switch msg := msg.(type) {
	case *qbtx.Transaction:
		node.TranscationMsgs = append(node.TranscationMsgs, msg)
	}
	return nil
}

// node.blockWhenClock,当时间片到时，将收到的交易信息打包
// 参数：无
// 返回值：处理错误error，默认为nil
func (node *Node) blockWhenClock() error {
	if len(node.TranscationMsgs) >= qblock.BLOCK_LENGTH {
		msgs := make([]*qbtx.Transaction, len(node.TranscationMsgs))
		copy(msgs, node.TranscationMsgs) // 复制缓冲数据
		request := node.block(msgs)
		node.TranscationMsgs = make([]*qbtx.Transaction, 0) // 清空重置

		file, _ := utils.Init_log(utils.SIGN_PATH + node.Node_name + ".log")
		log.SetPrefix("[BLOCK           SIGN]")
		log.Printf("block Height:%d\n", request.Height)
		defer file.Close()
		log.Println("Index of uss:", hex.EncodeToString(request.Block_uss.Sign_index.Sign_task_sn[:]))
		log.Println("plaintext:", hex.EncodeToString(request.Block_uss.USS_message))
		log.Println("signature:", hex.EncodeToString(request.Block_uss.USS_signature))
		log.Printf("Sign of block success\n\n\n")

		file, _ = utils.Init_log(utils.FLOW_PATH + node.Node_name + ".log")
		defer file.Close()
		log.SetPrefix("BLOCK-------------------")
		log.Println("collect enough transactions, create a block")
		node.MsgBroadcast <- request
	}
	return nil
}

func (node *Node) block(txs []*qbtx.Transaction) *qblock.Block {
	var block *qblock.Block
	bc := quantumbc.NewBlockchain(node.Node_name)
	preHash := bc.GetlastHash()
	lastHeight := bc.GetlastHeight()
	bc.DB.Close() // 关闭数据库
	block = qblock.NewBlock(txs, preHash, lastHeight+1)
	return block
}
