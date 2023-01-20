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
