package qbcli

import (
	"qb/pbft"
)

// 线程2：dispatchMsg，用于处理收到的消息，一般用于暂时存储消息
func (client *Client) dispatchMsg() {
	for {
		msg := <-client.MsgEntrance
		switch msg := msg.(type) {
		case *pbft.ReplyMsg:
			if len(client.ReplyMsgs) >= 2*pbft.F { // 收到符合要求的reply消息
				msgs := make([]*pbft.ReplyMsg, len(client.ReplyMsgs))
				copy(msgs, client.ReplyMsgs) // 复制缓冲数据
				msgs = append(msgs, msg)     // 附加新到达的消息
				client.MsgDelivery <- msgs
				client.ReplyMsgs = make([]*pbft.ReplyMsg, 0) // 清空
			} else {
				client.ReplyMsgs = append(client.ReplyMsgs, msg)
			}
		}
	}
}
