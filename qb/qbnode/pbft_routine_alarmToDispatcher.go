package qbnode

import "time"

// 线程2：alarmToDispatcher，提醒处理时间片（0.2s）到
func (node *NodeConsensus) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		node.Alarm <- true
	}
}
