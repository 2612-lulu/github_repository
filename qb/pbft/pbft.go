package pbft

type PBFT interface {
	GenReqMsg(operation string, node_name [2]byte) (*RequestMsg, bool)
	PrePrePare(request *RequestMsg) (*PrePrepareMsg, bool)
	PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, error)
	Commit(prepare *PrepareMsg) (*CommitMsg, error)
	GetReply(commit *CommitMsg) (*ReplyMsg, error)
}
