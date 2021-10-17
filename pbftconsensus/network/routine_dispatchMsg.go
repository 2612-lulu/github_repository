package network

import (
	"fmt"
	"log"
	"pbft"
	"qblock"
	"utils"
)

// 线程：dispatchMsg，用于处理收到的消息，一般是对信息进行暂存
func (consensus *NodeConsensus) dispatchMsg() {
	for {
		select {
		case msg := <-consensus.MsgEntrance: // 信息接收通道：如果MsgEntrance通道有消息传送过来，拿到msg
			err := consensus.routeMsg(msg) // 进行routeMsg
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		case <-consensus.Alarm:
			err := consensus.routeMsgWhenAlarmed()
			if err != nil {
				fmt.Println(err) // TODO: send err to ErrorChannel
			}
		}
	}
}

// routeMsg，对收到的消息进行暂存处理，满足要求时发送到消息处理通道
// 参数：收到的消息
// 返回值：处理错误[]error，默认为nil
func (consensus *NodeConsensus) routeMsg(msg interface{}) []error {
	switch msg := msg.(type) {
	case *qblock.Block:
		if consensus.PBFT.CurrentState == nil { // 如果此时不存在共识
			msgs := make([]*qblock.Block, len(consensus.PBFT.MsgBuffer.ReqMsgs))
			copy(msgs, consensus.PBFT.MsgBuffer.ReqMsgs)                // 复制缓冲数据
			msgs = append(msgs, msg)                                    // 附加新到达的消息
			consensus.PBFT.MsgBuffer.ReqMsgs = make([]*qblock.Block, 0) // 清空重置
			consensus.MsgDelivery <- msgs                               // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			defer file.Close()
			log.SetPrefix("[dispatch block]")
			log.Println("start a new pbft")
		} else {
			consensus.PBFT.MsgBuffer.ReqMsgs = append(consensus.PBFT.MsgBuffer.ReqMsgs, msg)
			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			defer file.Close()
			log.SetPrefix("[dispatch block error]")
			log.Println("exit another pbft")
		}
	// 处理PrePrepare信息
	case *pbft.PrePrepareMsg:
		if consensus.PBFT.CurrentState == nil { // 当CurrentState为nil时,此时不存在共识
			msgs := make([]*pbft.PrePrepareMsg, len(consensus.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, consensus.PBFT.MsgBuffer.PrePrepareMsgs)                      // 复制缓冲数据
			msgs = append(msgs, msg)                                                 // 附加新到达的消息
			consensus.PBFT.MsgBuffer.PrePrepareMsgs = make([]*pbft.PrePrepareMsg, 0) // 清空重置
			consensus.MsgDelivery <- msgs                                            // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			defer file.Close()
			log.SetPrefix("[dispatch PrePrepareMsg]")
			log.Println("[receive preprepare message, and prepare to resolve it]")
		} else { // 当CurrentState不为nil时，直接往MsgBuffer缓冲通道中进行添加
			consensus.PBFT.MsgBuffer.PrePrepareMsgs = append(consensus.PBFT.MsgBuffer.PrePrepareMsgs, msg)
		}
	// 处理Prepare信息
	case *pbft.PrepareMsg:
		if consensus.PBFT.CurrentState == nil || consensus.PBFT.CurrentState.Current_stage != pbft.PrePrepared {
			consensus.PBFT.MsgBuffer.PrepareMsgs = append(consensus.PBFT.MsgBuffer.PrepareMsgs, msg)
		} else {
			msgs := make([]*pbft.PrepareMsg, len(consensus.PBFT.MsgBuffer.PrepareMsgs))
			copy(msgs, consensus.PBFT.MsgBuffer.PrepareMsgs)                   // 复制缓冲数据
			msgs = append(msgs, msg)                                           // 附加新到达的消息
			consensus.PBFT.MsgBuffer.PrepareMsgs = make([]*pbft.PrepareMsg, 0) // 清空重置
			consensus.MsgDelivery <- msgs
			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			defer file.Close()
			log.SetPrefix("[dispatch PrepareMsg]")
			log.Println("[receive prepare message, and prepare to resolve it]") // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
		}
	// 处理CommitMsg信息
	case *pbft.CommitMsg:
		if consensus.PBFT.CurrentState == nil || consensus.PBFT.CurrentState.Current_stage != pbft.Prepared {
			consensus.PBFT.MsgBuffer.CommitMsgs = append(consensus.PBFT.MsgBuffer.CommitMsgs, msg)
		} else {
			msgs := make([]*pbft.CommitMsg, len(consensus.PBFT.MsgBuffer.CommitMsgs))
			copy(msgs, consensus.PBFT.MsgBuffer.CommitMsgs)                  // 复制缓冲数据
			msgs = append(msgs, msg)                                         // 附加新到达的消息
			consensus.PBFT.MsgBuffer.CommitMsgs = make([]*pbft.CommitMsg, 0) // 清空重置
			consensus.MsgDelivery <- msgs                                    // 信息发送通道：将msgs中的信息发送给MsgDelivery通道
			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			defer file.Close()
			log.SetPrefix("[dispatch CommitMsg]")
			log.Println("[receive commit message, and prepare to resolve it]")
		}
	}
	return nil
}

// routeMsgWhenAlarmed,当时间片到时，对收到的消息进行暂存处理，满足要求时发送到消息处理通道
// 参数：收到的消息
// 返回值：处理错误[]error，默认为nil
func (consensus *NodeConsensus) routeMsgWhenAlarmed() []error {
	if consensus.PBFT.CurrentState == nil {
		// 检查PrePrepareMsgs, 并发送到MsgDelivery.
		if len(consensus.PBFT.MsgBuffer.PrePrepareMsgs) != 0 {
			msgs := make([]*pbft.PrePrepareMsg, len(consensus.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, consensus.PBFT.MsgBuffer.PrePrepareMsgs)
			consensus.PBFT.MsgBuffer.PrePrepareMsgs = make([]*pbft.PrePrepareMsg, 0)
			consensus.MsgDelivery <- msgs
			file, _ := utils.Init_log(PBFT_LOG_PATH + "dispatch_" + consensus.Node_name + ".log")
			defer file.Close()
			log.SetPrefix("[dispatch request when alarmed]")
			log.Println("[start a pbft]")
		}
		if len(consensus.PBFT.MsgBuffer.PrePrepareMsgs) != 0 {
			msgs := make([]*pbft.PrePrepareMsg, len(consensus.PBFT.MsgBuffer.PrePrepareMsgs))
			copy(msgs, consensus.PBFT.MsgBuffer.PrePrepareMsgs)
			consensus.PBFT.MsgBuffer.PrePrepareMsgs = make([]*pbft.PrePrepareMsg, 0)
			consensus.MsgDelivery <- msgs
			file, _ := utils.Init_log(PBFT_LOG_PATH + "dispatch_" + consensus.Node_name + ".log")
			defer file.Close()
			log.SetPrefix("[dispatch preprepare when alarmed]")
			log.Println("[receive preprepare message, and prepare to resolve it]")
		}
	} else {
		switch consensus.PBFT.CurrentState.Current_stage {
		case pbft.PrePrepared:
			// 检查PrepareMsgs,并发送到MsgDelivery.
			if len(consensus.PBFT.MsgBuffer.PrepareMsgs) != 0 {
				msgs := make([]*pbft.PrepareMsg, len(consensus.PBFT.MsgBuffer.PrepareMsgs))
				copy(msgs, consensus.PBFT.MsgBuffer.PrepareMsgs)
				consensus.PBFT.MsgBuffer.PrepareMsgs = make([]*pbft.PrepareMsg, 0) // 清空重置
				consensus.MsgDelivery <- msgs
			}
		case pbft.Prepared:
			// 检查CommitMsgs,并发送到MsgDelivery.
			if len(consensus.PBFT.MsgBuffer.CommitMsgs) != 0 {
				msgs := make([]*pbft.CommitMsg, len(consensus.PBFT.MsgBuffer.CommitMsgs))
				copy(msgs, consensus.PBFT.MsgBuffer.CommitMsgs)
				consensus.PBFT.MsgBuffer.CommitMsgs = make([]*pbft.CommitMsg, 0) // 清空重置
				consensus.MsgDelivery <- msgs
			}
		}
	}
	return nil
}
