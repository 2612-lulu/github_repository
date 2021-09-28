package qbnode

import "time"

// 线程1：clockToBlock，提醒打包时间片（1s）到
func (node *Node) clockToBlock() {
	for {
		time.Sleep(BlockTimeDuration)
		node.Block_clock <- true
	}
}
