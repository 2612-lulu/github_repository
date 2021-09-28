package qbcli

import (
	"fmt"
	"log"
	"qb/pbft"
	"qb/qbtools"
)

// 线程2：dispatchMsg，用于处理收到的消息，一般用于暂时存储消息
func (client *Client) dispatchMsg() {
	for {
		select {
		case msg := <-client.MsgEntrance:
			err := client.routeMsg(msg) // 进行routeMsg
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		case <-client.Alarm:
			err := client.routeMsgWhenAlarmed()
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		}
	}
}

func (client *Client) routeMsg(msg interface{}) []error {
	switch msg := msg.(type) {
	case *pbft.ReplyMsg:
		if len(client.ReplyMsgs) >= 2*pbft.F && client.CurrentState != pbft.Idle { // 收到符合要求的reply消息
			msgs := make([]*pbft.ReplyMsg, len(client.ReplyMsgs))
			copy(msgs, client.ReplyMsgs) // 复制缓冲数据
			msgs = append(msgs, msg)     // 附加新到达的消息
			client.MsgDelivery <- msgs
			client.ReplyMsgs = make([]*pbft.ReplyMsg, 0) // 清空
			qbtools.Init_log(CLIENT_LOG_PATH + "dispatch_" + client.Client_name + ".log")
			log.SetPrefix("【dispatch reply】")
			log.Println("get a reply")
		} else {
			client.ReplyMsgs = append(client.ReplyMsgs, msg)
		}

	}
	return nil
}

func (client *Client) routeMsgWhenAlarmed() []error {
	if len(client.ReplyMsgs) != 0 {
		msgs := make([]*pbft.ReplyMsg, len(client.ReplyMsgs))
		copy(msgs, client.ReplyMsgs) // 复制缓冲数据
		client.MsgDelivery <- msgs
		client.ReplyMsgs = make([]*pbft.ReplyMsg, 0) // 清空
	}
	return nil
}
