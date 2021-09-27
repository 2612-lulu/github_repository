package qbnode

import (
	"fmt"
	"log"
	"qb/qblock"
	"qb/qbtools"
	"qb/qbtx"
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

		qbtools.Init_log(LOG_PATH + "block_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[block.Transaction]")
		log.Println("get a transcation message")
	}
	return nil
}

// node.blockWhenClock,当时间片到时，将收到的交易信息打包
// 参数：无
// 返回值：处理错误error，默认为nil
func (node *Node) blockWhenClock() error {
	if len(node.TranscationMsgs) >= qblock.BLOCK_LENGTH {
		qbtools.Init_log(LOG_PATH + "block_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[block.Block]")
		log.Println("create a new block, and put it into MsgEntrance channel")
		//fmt.Println("打包了一个区块，长度=", len(node.PBFT.MsgBuffer.TranscationMsgs))

		msgs := make([]*qbtx.Transaction, len(node.TranscationMsgs))
		copy(msgs, node.TranscationMsgs) // 复制缓冲数据
		request := qblock.NewBlock(msgs, []byte{}, 0)
		node.TranscationMsgs = make([]*qbtx.Transaction, 0) // 清空重置
		node.PBFT_consensus.MsgEntrance <- request
	}
	return nil
}
