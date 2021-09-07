package mylog

import (
	"fmt"
	"qb/pbft"
)

func LogMsg(msg interface{}) {
	switch msg := msg.(type) {
	//case *pbft.RequestMsg:
	//fmt.Printf("[REQUEST] ClientID: %s, Timestamp: %d\n", msg.Client_name[:], msg.Time_stamp)
	//case *pbft.PrePrepareMsg:
	//fmt.Printf("[PREPREPARE] ClientID: %s,  SequenceID: %d\n", msg.Request.Client_name[:], msg.Sequence_number)
	case *pbft.PrepareMsg:
		fmt.Printf("[PREPARE] NodeID: %d\n", msg.Node_i)
	case *pbft.CommitMsg:
		fmt.Printf("[PREPARE] NodeID: %d\n", msg.Node_i)
	}
}

func LogStage(stage string, isDone bool) {
	if isDone {
		fmt.Printf("[STAGE-DONE] %s\n", stage)
	} else {
		fmt.Printf("[STAGE-BEGIN] %s\n", stage)
	}
}
