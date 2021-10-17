package network

import (
	"errors"
	"fmt"
	"log"
	"pbft"
	"qblock"
	"utils"
)

// 线程：resolveMsg，用于对收到的信息作具体处理
func (consensus *NodeConsensus) resolveMsg() {
	for {
		msgs := <-consensus.MsgDelivery // 从调度器通道中获取缓存信息
		switch msgs := msgs.(type) {
		// 节点表决决策信息
		case []*qblock.Block:
			if consensus.PBFT.CurrentState == nil {
				err := consensus.resolveRequestMsg(msgs)
				if err != nil {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
		case []*pbft.PrePrepareMsg:
			if consensus.PBFT.CurrentState == nil {
				errs := consensus.resolvePrePrepareMsg(msgs)
				if len(errs) != 0 {
					for _, err := range errs {
						fmt.Println(err) // TODO: send err to ErrorChannel
					}
				}
			}
		case []*pbft.PrepareMsg:
			errs := consensus.resolvePrepareMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}

		case []*pbft.CommitMsg:
			errs := consensus.resolveCommitMsg(msgs)
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err) // TODO: send err to ErrorChannel
				}
			}
		}
	}
}

// createStateForNewConsensus，创建新的共识
// 参数：
// 返回值：创建错误error，默认为nil
func (consensus *NodeConsensus) createStateForNewConsensus() error {
	if consensus.PBFT.CurrentState != nil { // 判断当前节点是不是处于其他阶段（预准备阶段或者准备阶段等等）
		return errors.New("another pbft consensus is ongoing") // 如果有，则输出提示
	}
	var lastSequenceID int64 // 获取上一个序列号
	// 判断当前阶段是否已经发送过消息
	if len(consensus.Committed) == 0 { // 如果是首次进行共识，则上一个序列号lastSequenceID设置为-1
		lastSequenceID = -1
	} else { // 否则取出上一个序列号
		lastSequenceID = consensus.Committed[len(consensus.Committed)-1].Sequence_number
	}
	// 创建新的节点状态，即进行节点状态的初始化
	consensus.PBFT.CurrentState = pbft.CreateState(consensus.View.ID, lastSequenceID)
	return nil
}

// resolveRequestMsg,处理收到的区块数组
// 参数：区块数组[]*block.Block
// 返回值：处理错误[]error
func (consensus *NodeConsensus) resolveRequestMsg(msgs []*qblock.Block) []error {
	errs := make([]error, 0)

	err := consensus.createStateForNewConsensus() // 创建节点状态，因为从节点刚开始进入共识，所以需要初始化状态
	if err != nil {
		fmt.Println(err)
	}
	// 批量处理req信息
	for _, req := range msgs {
		err := consensus.resolveRequest(req)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 { // 如果有处理错误，则输出错误
		return errs
	}
	return nil
}

// resolveRequestMsg,处理单条区块，即请求消息，生成预准备消息
// 参数：区块*block.Block
// 返回值：处理错误error，默认为nil
func (consensus *NodeConsensus) resolveRequest(msgs *qblock.Block) error {
	if msgs != nil {
		prePrepareMsg := consensus.PBFT.CurrentState.PrePrePare(msgs) // 进入共识，获得preprepare消息
		if prePrepareMsg != nil {
			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			log.SetPrefix("[resolve requestMsg success]")
			log.Println("put pre-prepare message into broadcast channel")
			defer file.Close()

			consensus.MsgBroadcast <- prePrepareMsg // 将待广播消息放入通道
		}
	}

	return nil
}

// resolvePrePrepareMsg，由从节点处理PrePrepare消息数组
// 参数：预准备消息数组[]*pbft.PrePrepareMsg
// 返回值：处理错误[]error
func (consensus *NodeConsensus) resolvePrePrepareMsg(msgs []*pbft.PrePrepareMsg) []error {
	errs := make([]error, 0)
	err := consensus.createStateForNewConsensus() // 创建节点状态，因为从节点刚开始进入共识，所以需要初始化状态
	if err != nil {
		fmt.Println(err)
	}
	// 批量处理pre-prepare信息
	for _, prePrepareMsg := range msgs {
		err := consensus.resolvePrePrepare(prePrepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 { // 如果有处理错误，则输出错误
		return errs
	}
	return nil
}

// resolvePrePrepare,处理单条预准备消息，生成准备消息
// 参数：预准备消息*pbft.PrePrepareMsg
// 返回值：处理错误error，默认为nil
func (consensus *NodeConsensus) resolvePrePrepare(prePrepareMsg *pbft.PrePrepareMsg) error {
	if prePrepareMsg != nil {
		prePareMsg := consensus.PBFT.CurrentState.PrePare(prePrepareMsg) // 获得prepare信息
		if prePareMsg != nil {
			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			log.SetPrefix("[resolve pre-prepareMsg success]")
			log.Println("put prepare message into broadcast channel")
			defer file.Close()
			consensus.MsgBroadcastPrepare <- prePareMsg // 将待广播消息放入通道
		}

	}

	return nil
}

// resolvePrepareMsg,处理准备消息数组，得到提交消息
// 参数：准备消息数组[]*pbft.PrepareMsg
// 返回值：处理错误[]error
func (consensus *NodeConsensus) resolvePrepareMsg(msgs []*pbft.PrepareMsg) []error {
	errs := make([]error, 0)

	// 批量处理准备消息
	for _, prepareMsg := range msgs {
		err := consensus.resolvePrepare(prepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

// resolvePrepare,处理单条准备消息，生成提交消息
// 参数：准备消息*pbft.PrepareMsg
// 返回值：处理错误error，默认为nil
func (consensus *NodeConsensus) resolvePrepare(prepareMsg *pbft.PrepareMsg) error {
	if prepareMsg != nil {
		commitMsg := consensus.PBFT.CurrentState.Commit(prepareMsg)
		if commitMsg != nil {
			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			log.SetPrefix("[resolve prepareMsg success]")
			log.Println("put commit message into broadcast channel")
			defer file.Close()

			consensus.MsgBroadcastCommit <- commitMsg // 将待广播消息放入通道
		}
	} else {
		file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
		log.SetPrefix("[resolve prepareMsg error]")
		defer file.Close()
		log.Println("prepare is nil")
	}

	return nil
}

// resolveCommitMsg,所有节点处理接收的commit消息数组得到reply消息
// 参数：提交消息数组[]*pbft.CommitMsg
// 返回值：处理错误[]error
func (consensus *NodeConsensus) resolveCommitMsg(msgs []*pbft.CommitMsg) []error {
	errs := make([]error, 0)

	// 批量处理commit信息
	for _, commitMsg := range msgs {
		err := consensus.resolveCommit(commitMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

// consensus.resolveCommit,处理单条提交消息，生成应答消息
// 参数：准备消息*pbft.CommitMsg
// 返回值：处理错误error，默认为nil
func (consensus *NodeConsensus) resolveCommit(commitMsg *pbft.CommitMsg) error {
	if commitMsg != nil {
		replyMsgs := consensus.PBFT.CurrentState.Reply(commitMsg)
		if replyMsgs != nil {
			consensus.Committed = append(consensus.Committed, commitMsg)
			consensus.MsgBroadcast <- replyMsgs // 将待广播消息放入通道

			file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
			log.SetPrefix("[resolve commitMsg success]")
			log.Println("put reply message into broadcast channel")
			defer file.Close()
		}
	} else {
		file, _ := utils.Init_log(PBFT_LOG_PATH + consensus.Node_name + ".log")
		log.SetPrefix("[resolve commitMsg error]")
		log.Println("commit is nil")
		defer file.Close()
	}
	return nil
}
