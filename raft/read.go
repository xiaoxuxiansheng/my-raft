package raft

type ReadState struct {
	// 保存这笔读请求时的 committed index
	Index uint64
	// 读请求的唯一 id
	RequestCtx []byte
}

type readIndexStatus struct {
	req Message
	// 接收到读请求时的 commit index
	index uint64
	// 保存有哪些节点对这笔读请求作了响应
	acks map[uint64]struct{}
}

type readOnly struct {
	// 当前仍未处理完成的读请求，key：读请求的标识 id，val：读请求的状态信息
	pendingReadIndex map[string]*readIndexStatus
	// 读请求队列，元素为读请求标识 id，通过队列记录了读请求的先后顺序
	readIndexQueue []string
}

func newReadOnly() *readOnly {
	return &readOnly{
		pendingReadIndex: make(map[string]*readIndexStatus),
	}
}
