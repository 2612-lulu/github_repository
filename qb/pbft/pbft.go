package pbft

type PBFT interface {
	GenReqMsg(operation string, node_name [2]byte) (*RequestMsg, bool)
	PrePrePare(request *RequestMsg) (*PrePrepareMsg, bool)
	PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, bool)
	GetCommitMsg(prepare *PrepareMsg) *CommitMsg
	GetReplyMsg(commit *CommitMsg) *ReplyMsg
}
