package qbnode

import (
	"fmt"
	"log"
	"qb/pbft"
	"qb/qblock"
	"qb/qbtools"
)

// 线程3：dispatchMsg，用于处理收到的消息，一般是对信息进行暂存
func (node *NodeConsensus) dispatchMsg() {
	for {
		select {
		case msg := <-node.MsgEntrance: // 信息接收通道：如果MsgEntrance通道有消息传送过来，拿到msg
			err := node.routeMsg(msg) // 进行routeMsg
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		case <-node.Alarm:
			err := node.routeMsgWhenAlarmed()
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		}
	}
}

// node.routeMsg，对收到的消息进行暂存处理，满足要求时发送到消息处理通道
// 参数：收到的消息
// 返回值：处理错误[]byte，默认为nil
func (node *NodeConsensus) routeMsg(msg interface{}) []error {
	switch msg := msg.(type) {
	case *qblock.Block:
		if node.PBFT.CurrentState == nil { // 如果此时不存在共识
			msgs := make([]*qblock.Block, len(node.PBFT.MsgBuffer.ReqMsgs))
			copy(msgs, node.PBFT.MsgBuffer.ReqMsgs)                // 复制缓冲数据
			msgs = append(msgs, msg)                               // 附加新到达的消息
			node.PBFT.MsgBuffer.ReqMsgs = make([]*qblock.Block, 0) // 清空重置
			node.MsgDelivery <- msgs                               // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		} else {
			node.PBFT.MsgBuffer.ReqMsgs = append(node.PBFT.MsgBuffer.ReqMsgs, msg)

			qbtools.Init_log(PBFT_LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[dispatch block error]")
			log.Println("exit another pbft")
		}
	// 处理PrePrepare信息
	case *pbft.PrePrepareMsg:
		if node.PBFT.CurrentState == nil { // 当CurrentState为nil时,此时不存在共识
			msgs := make([]*pbft.PrePrepareMsg, len(node.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrePrepareMsgs)                      // 复制缓冲数据
			msgs = append(msgs, msg)                                            // 附加新到达的消息
			node.PBFT.MsgBuffer.PrePrepareMsgs = make([]*pbft.PrePrepareMsg, 0) // 清空重置
			node.MsgDelivery <- msgs                                            // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		} else { // 当CurrentState不为nil时，直接往MsgBuffer缓冲通道中进行添加
			node.PBFT.MsgBuffer.PrePrepareMsgs = append(node.PBFT.MsgBuffer.PrePrepareMsgs, msg)
			qbtools.Init_log(PBFT_LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[dispatch PrePrepareMsg error]")
			log.Println("[get a pre-prepare message, but don't put it into channel]")
		}
	// 处理Prepare信息
	case *pbft.PrepareMsg:
		if node.PBFT.CurrentState == nil || node.PBFT.CurrentState.Current_stage != pbft.PrePrepared {
			node.PBFT.MsgBuffer.PrepareMsgs = append(node.PBFT.MsgBuffer.PrepareMsgs, msg)
			qbtools.Init_log(PBFT_LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[dispatch PrepareMsg error]")
			log.Println("[get a prepare message,but don't put it into channel]")
		} else {
			msgs := make([]*pbft.PrepareMsg, len(node.PBFT.MsgBuffer.PrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrepareMsgs)                   // 复制缓冲数据
			msgs = append(msgs, msg)                                      // 附加新到达的消息
			node.PBFT.MsgBuffer.PrepareMsgs = make([]*pbft.PrepareMsg, 0) // 清空重置
			node.MsgDelivery <- msgs                                      // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		}
	// 处理CommitMsg信息
	case *pbft.CommitMsg:
		if node.PBFT.CurrentState == nil || node.PBFT.CurrentState.Current_stage != pbft.Prepared {
			node.PBFT.MsgBuffer.CommitMsgs = append(node.PBFT.MsgBuffer.CommitMsgs, msg)
			qbtools.Init_log(PBFT_LOG_PATH + "dispatch_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[dispatch CommitMsg error]")
			log.Println("[get a commit message,but don't put it into channel]")
		} else {
			msgs := make([]*pbft.CommitMsg, len(node.PBFT.MsgBuffer.CommitMsgs))
			copy(msgs, node.PBFT.MsgBuffer.CommitMsgs)                  // 复制缓冲数据
			msgs = append(msgs, msg)                                    // 附加新到达的消息
			node.PBFT.MsgBuffer.CommitMsgs = make([]*pbft.CommitMsg, 0) // 清空重置
			node.MsgDelivery <- msgs                                    // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		}
	}
	return nil
}

// node.routeMsgWhenAlarmed,当时间片到时，对收到的消息进行暂存处理，满足要求时发送到消息处理通道
func (node *NodeConsensus) routeMsgWhenAlarmed() []error {
	if node.PBFT.CurrentState == nil {
		// 检查PrePrepareMsgs, 并发送到MsgDelivery.
		if len(node.PBFT.MsgBuffer.PrePrepareMsgs) != 0 {
			msgs := make([]*pbft.PrePrepareMsg, len(node.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.PBFT.MsgBuffer.PrePrepareMsgs)
			node.PBFT.MsgBuffer.PrePrepareMsgs = make([]*pbft.PrePrepareMsg, 0)
			node.MsgDelivery <- msgs
		}
	} else {
		switch node.PBFT.CurrentState.Current_stage {
		case pbft.PrePrepared:
			// 检查PrepareMsgs,并发送到MsgDelivery.
			if len(node.PBFT.MsgBuffer.PrepareMsgs) != 0 {
				msgs := make([]*pbft.PrepareMsg, len(node.PBFT.MsgBuffer.PrepareMsgs))
				copy(msgs, node.PBFT.MsgBuffer.PrepareMsgs)
				node.PBFT.MsgBuffer.PrepareMsgs = make([]*pbft.PrepareMsg, 0) // 清空重置
				node.MsgDelivery <- msgs
			}
		case pbft.Prepared:
			// 检查CommitMsgs,并发送到MsgDelivery.
			if len(node.PBFT.MsgBuffer.CommitMsgs) != 0 {
				msgs := make([]*pbft.CommitMsg, len(node.PBFT.MsgBuffer.CommitMsgs))
				copy(msgs, node.PBFT.MsgBuffer.CommitMsgs)
				node.PBFT.MsgBuffer.CommitMsgs = make([]*pbft.CommitMsg, 0) // 清空重置
				node.MsgDelivery <- msgs
			}
		}
	}
	return nil
}
