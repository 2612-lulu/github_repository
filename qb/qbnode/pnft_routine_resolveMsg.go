package qbnode

import (
	"errors"
	"fmt"
	"log"
	"qb/pbft"
	"qb/qblock"
	"qb/qbtools"
)

// 线程4：resolveMsg，用于对收到的信息作具体处理
func (node *NodeConsensus) resolveMsg() {
	for {
		msgs := <-node.MsgDelivery // 从调度器通道中获取缓存信息
		switch msgs := msgs.(type) {
		// 节点表决决策信息
		case []*qblock.Block:
			if node.PBFT.CurrentState != nil {
				qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
				log.SetPrefix(node.Node_name + "-[resolve block error]")
				log.Println("get a block,but another pbft consensus is ongoing")
			} else {
				qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
				log.SetPrefix(node.Node_name + "-[resolve block.Block]")
				log.Println("Msgdelivery channel get a new block,and create a new consensus")

				err := node.resolveRequestMsg(msgs)
				if err != nil {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
		case []*pbft.PrePrepareMsg:
			if node.PBFT.CurrentState != nil {
				qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
				log.SetPrefix(node.Node_name + "-[resolve PrePrepareMsg error]")
				log.Println("get a preprepare,but another pbft consensus is ongoing")
			} else {
				qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
				log.SetPrefix(node.Node_name + "-[resolve PrePrepareMsg]")
				log.Println("Msgdelivery channel get a preprepare, and handle it")

				errs := node.resolvePrePrepareMsg(msgs)
				if len(errs) != 0 {
					for _, err := range errs {
						fmt.Println(err) // TODO: send err to ErrorChannel
					}
				}
			}
		case []*pbft.PrepareMsg:
			qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[resolve PrepareMsg]")
			log.Println("Msgdelivery channel get a prepare message, and handle it")

			errs := node.resolvePrepareMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}

		case []*pbft.CommitMsg:
			qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
			log.SetPrefix(node.Node_name + "-[resolve CommitMsg]")
			log.Println("Msgdelivery channel get a commit message, and handle it")

			errs := node.resolveCommitMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
		}
	}
}

// node.createStateForNewConsensus，创建新的共识
// 参数：
// 返回值：创建错误error，默认为nil
func (node *NodeConsensus) createStateForNewConsensus() error {
	if node.PBFT.CurrentState != nil { // 判断当前节点是不是处于其他阶段（预准备阶段或者准备阶段等等）
		return errors.New("another pbft consensus is ongoing") // 如果有，则输出提示
	}
	var lastSequenceID int64 // 获取上一个序列号
	// 判断当前阶段是否已经发送过消息
	if len(node.Committed) == 0 { // 如果是首次进行共识，则上一个序列号lastSequenceID设置为-1
		lastSequenceID = -1
	} else { // 否则取出上一个序列号
		lastSequenceID = node.Committed[len(node.Committed)-1].Sequence_number
	}
	// 创建新的节点状态，即进行节点状态的初始化
	node.PBFT.CurrentState = pbft.CreateState(node.View.ID, lastSequenceID)
	return nil
}

// node.resolveRequestMsg,处理收到的区块数组
// 参数：区块数组[]*block.Block
// 返回值：处理错误[]error
func (node *NodeConsensus) resolveRequestMsg(msgs []*qblock.Block) []error {
	errs := make([]error, 0)

	err := node.createStateForNewConsensus() // 创建节点状态，因为从节点刚开始进入共识，所以需要初始化状态
	if err != nil {
		fmt.Println(err)
	}
	// 批量处理req信息
	for _, req := range msgs {
		err := node.resolveRequest(req)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 { // 如果有处理错误，则输出错误
		return errs
	}
	return nil
}

// node.resolveRequestMsg,处理单条区块，即请求消息，生成预准备消息
// 参数：区块*block.Block
// 返回值：处理错误error，默认为nil
func (node *NodeConsensus) resolveRequest(msgs *qblock.Block) error {
	node.Req = msgs
	prePrepareMsg, err := node.PBFT.CurrentState.PrePrePare(msgs) // 进入共识，获得preprepare消息
	if err != nil {
		return err
	} else {
		qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[resolve requestMsg success]")
		log.Println("put pre-prepare message into broadcast channel")

		node.MsgBroadcast <- prePrepareMsg // 将待广播消息放入通道
		return nil
	}

}

// node.resolvePrePrepareMsg，由从节点处理PrePrepare消息数组
// 参数：预准备消息数组[]*pbft.PrePrepareMsg
// 返回值：处理错误[]error
func (node *NodeConsensus) resolvePrePrepareMsg(msgs []*pbft.PrePrepareMsg) []error {
	errs := make([]error, 0)
	err := node.createStateForNewConsensus() // 创建节点状态，因为从节点刚开始进入共识，所以需要初始化状态
	if err != nil {
		fmt.Println(err)
	}
	// 批量处理pre-prepare信息
	for _, prePrepareMsg := range msgs {
		err := node.resolvePrePrepare(prePrepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 { // 如果有处理错误，则输出错误
		return errs
	}
	return nil
}

// node.resolvePrePrepare,处理单条预准备消息，生成准备消息
// 参数：预准备消息*pbft.PrePrepareMsg
// 返回值：处理错误error，默认为nil
func (node *NodeConsensus) resolvePrePrepare(prePrepareMsg *pbft.PrePrepareMsg) error {
	node.Req = prePrepareMsg.Request
	prePareMsg, err := node.PBFT.CurrentState.PrePare(prePrepareMsg) // 获得prepare信息
	if err != nil {
		return err
	}
	if prePareMsg != nil {
		qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[resolve pre-prepareMsg success]")
		log.Println("put prepare message into broadcast channel")

		node.MsgBroadcast <- prePareMsg // 将待广播消息放入通道
	}
	return nil
}

// node.resolvePrepareMsg,处理准备消息数组，得到提交消息
// 参数：准备消息数组[]*pbft.PrepareMsg
// 返回值：处理错误[]error
func (node *NodeConsensus) resolvePrepareMsg(msgs []*pbft.PrepareMsg) []error {
	errs := make([]error, 0)

	// 批量处理准备消息
	for _, prepareMsg := range msgs {
		err := node.resolvePrepare(prepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

// node.resolvePrepare,处理单条准备消息，生成提交消息
// 参数：准备消息*pbft.PrepareMsg
// 返回值：处理错误error，默认为nil
func (node *NodeConsensus) resolvePrepare(prepareMsg *pbft.PrepareMsg) error {
	commitMsg, err := node.PBFT.CurrentState.Commit(prepareMsg)
	if err != nil {
		return err
	}
	if commitMsg != nil {
		qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[resolve prepareMsg success]")
		log.Println("put commit message into broadcast channel")

		node.MsgBroadcast <- commitMsg // 将待广播消息放入通道
	}
	return nil
}

// node.resolveCommitMsg,所有节点处理接收的commit消息数组得到reply消息
// 参数：提交消息数组[]*pbft.CommitMsg
// 返回值：处理错误[]error
func (node *NodeConsensus) resolveCommitMsg(msgs []*pbft.CommitMsg) []error {
	errs := make([]error, 0)

	// 批量处理commit信息
	for _, commitMsg := range msgs {
		err := node.resolveCommit(commitMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

// node.resolveCommit,处理单条提交消息，生成应答消息
// 参数：准备消息*pbft.CommitMsg
// 返回值：处理错误error，默认为nil
func (node *NodeConsensus) resolveCommit(commitMsg *pbft.CommitMsg) error {
	replyMsg, err := node.PBFT.CurrentState.Reply(commitMsg)
	if err != nil {
		return err
	}
	if replyMsg != nil {
		qbtools.Init_log(PBFT_LOG_PATH + "resolve_" + node.Node_name + ".log")
		log.SetPrefix(node.Node_name + "-[resolve commitMsg success]")
		log.Println("put reply message into broadcast channel")

		node.Committed = append(node.Committed, commitMsg)
		node.MsgBroadcast <- replyMsg // 将待广播消息放入通道

	}
	return nil
}
