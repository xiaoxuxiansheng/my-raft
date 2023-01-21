package raft

type ProgressStateType int32

const (
	ProgressStateProbe     ProgressStateType = 0
	ProgressStateReplicate ProgressStateType = 1
)

type Progress struct {
	// leader 记录某个节点的状态信息，包括当前确认同步的日志索引和下一笔待同步的日志索引
	Match, Next uint64
	State       ProgressStateType
}

func (pr *Progress) maybeUpdate(n uint64) bool {
	var updated bool
	if pr.Match < n {
		pr.Match = n
		updated = true
	}
	if pr.Next < n+1 {
		pr.Next = n + 1
	}
	return updated
}

func (pr *Progress) mayDecrTo(logIndex, rejectHint uint64) bool {
	if pr.Next-1 != logIndex {
		return false
	}
	if pr.Next = min(logIndex, rejectHint+1); pr.Next < 1 {
		pr.Next = 1
	}
	return true
}
