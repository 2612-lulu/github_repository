package pbft

import "qb/qblock"

type PBFT interface {
	PrePrePare(request *qblock.Block) (*PrePrepareMsg, error)
	PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, error)
	Commit(prepare *PrepareMsg) (*CommitMsg, error)
	Reply(commit *CommitMsg) ([]*ReplyMsg, error)
}
