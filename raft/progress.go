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
