package raft

func (r *raft) becomeFollower(term, lead uint64) {
	r.reset(term)
	r.step = stepFollower
	r.tick = r.tickElection
	r.lead = lead
	r.state = StateFollower
}

// follower 的状态机函数
func stepFollower(r *raft, m Message) {
	switch m.Type {
	case MsgProp:
		// 如果没有 lead，直接忽视
		if r.lead == None {
			return
		}

		// 如果有 lead，转给 leader
		m.To = r.lead
		r.send(m)

	case MsgApp:
		// 处理同步日志请求
		// 收到 leader 的同步日志请求，重置心跳
		r.electionElapsed = 0
		r.lead = m.From
		r.handleAppendEntries(m)
	case MsgHearbeat:
		// 处理心跳请求
		// 收到 leader 的同步日志请求，重置心跳
		r.electionElapsed = 0
		r.lead = m.From
		r.handleHeartbeat(m)
	case MsgReadIndex:
		// 处理读请求
		// 没有 lead 则忽略
		if r.lead == None {
			return
		}

		// 有 lead 则转发
		m.To = r.lead
		r.send(m)

	case MsgReadIndexResp:
		// 处理读请求响应
		r.readStates = append(r.readStates, ReadState{Index: m.LogIndex, RequestCtx: m.Entries[0].Data})
	}
}

func (r *raft) handleAppendEntries(m Message) {
	// 如果是已提交过的消息，直接无视
	if m.LogIndex < r.raftLog.commitIndex {
		r.send(Message{To: m.From, Type: MsgAppResp, LogIndex: r.raftLog.commitIndex})
		return
	}

	// 尝试添加，如果失败则 reject
	if mLastIndex, ok := r.raftLog.maybeAppend(m.LogIndex, m.LogIndex, m.CommitIndex, m.Entries...); ok {
		r.send(Message{To: m.From, Type: MsgAppResp, LogIndex: mLastIndex})
		return
	}

	// 添加失败
	r.send(Message{To: m.From, Type: MsgAppResp, LogIndex: m.LogIndex, Reject: true, RejectHint: r.raftLog.lastIndex()})
}

func (r *raft) handleHeartbeat(m Message) {
	r.raftLog.commitTo(m.CommitIndex)
	r.send(Message{To: m.From, Type: MsgHearbeatResp, Context: m.Context})
}
