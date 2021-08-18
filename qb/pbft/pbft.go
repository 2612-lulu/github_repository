package pbft

type PBFT interface {
	PrePrePare(request *RequestMsg) (*PrePrepareMsg, bool)
	PrePare(preprepare *PrePrepareMsg) (*PrepareMsg, bool)
	GetCommitMsg(prepare *PrepareMsg) *CommitMsg
	GetReplyMsg(commit *CommitMsg) *ReplyMsg
}
