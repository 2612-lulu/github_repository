package pbft

import "qblock"

type PBFT interface {
	PrePrePare(request qblock.Block) PrePrepareMsg
	PrePare(preprepare PrePrepareMsg) PrepareMsg
	Commit(prepare PrepareMsg) CommitMsg
	Reply(commit CommitMsg) ReplyMsg
}
