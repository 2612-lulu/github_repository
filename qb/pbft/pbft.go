package pbft

import "qb/block"

type PBFT interface {
	PrePrePare(request *block.Block) (*PrePrepareMsg, error)
	PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, error)
	Commit(prepare *PrepareMsg) (*CommitMsg, error)
	Reply(commit *CommitMsg) (*ReplyMsg, error)
}
