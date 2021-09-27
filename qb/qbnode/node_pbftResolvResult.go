package qbnode

import (
	"fmt"
	"qb/pbft"
)

// 线程2：blockMsg，用于打包收到的交易信息
func (node *Node) pbftResolveResult() {
	for {
		msg := <-node.PBFT_consensus.Result // 从调度器通道中获取缓存信息
		switch msg := msg.(type) {
		case *pbft.ResultMsg:
			if msg.Result {
				fmt.Println("prepare to add block")
			}

		}
	}
}
