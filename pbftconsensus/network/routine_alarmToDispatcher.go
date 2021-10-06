package network

import "time"

// 线程2：alarmToDispatcher，提醒处理时间片（0.2s）到
func (consensus *NodeConsensus) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		consensus.Alarm <- true
	}
}
