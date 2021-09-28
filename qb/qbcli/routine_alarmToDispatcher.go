package qbcli

import "time"

// 线程2：alarmToDispatcher，提醒处理时间片（0.2s）到
func (client *Client) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		client.Alarm <- true
	}
}
