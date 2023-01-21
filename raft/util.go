package raft

func IsResponseMsg(typ MessageType) bool {
	return typ == MsgAppResp || typ == MsgHearbeatResp || typ == MsgVoteResp || typ == MsgPreVoteResp
}

func numOfPendingConf(ents []Entry) int {
	var n int
	for _, ent := range ents {
		if ent.Type == EntryConfChange {
			n++
		}
	}
	return n
}

type uint64Slice []uint64

func (u uint64Slice) Len() int {
	return len(u)
}

func (u uint64Slice) Less(i, j int) bool {
	return u[i] < u[j]
}

func (u uint64Slice) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
