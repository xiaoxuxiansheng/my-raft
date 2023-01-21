package raft

func (r *raft) becomeLeader() {
	if r.state == StateFollower {
		panic("invalid transition [follower -> leader]")
	}
	r.step = stepLeader
	r.reset(r.Term)
	r.tick = r.tickHeartbeat
	r.lead = r.id
	r.state = StateLeader

	// 新上任的 leader 需要传一条当前的任期空消息
	r.appendEntry([]Entry{{Data: nil}}...)
}

func stepLeader(r *raft, m Message) {
	switch m.Type {
	case MsgBeat:
		// 广播心跳
		r.bcastHeartbeat()
		return
		// 写请求提议
	case MsgProp:
		if len(m.Entries) == 0 {
			panic("propose entries can not be empty")
		}
		// 检查自身是否在集群中
		if _, ok := r.prs[r.id]; !ok {
			return
		}

		// 判断是否有配置变更的信息

		// 先将日志添加到自身的日志中
		r.appendEntry(m.Entries...)

		// 广播发起提议
		r.bcastAppend()
		return
	case MsgReadIndex:
		// 处理读请求
		return
	}

	// 处理响应类型的消息
	// 首先检查消息发送者是否位于集群中
	pr, ok := r.prs[m.From]
	if !ok {
		return
	}

	switch m.Type {
	case MsgAppResp:
		// 如果同步日志请求被拒绝了
		if m.Reject {
			// 根据 rejectHint 发送新的日志
			if pr.mayDecrTo(m.LogIndex, m.RejectHint) {
				r.sendAppend(m.From)
			}
		}

	case MsgHearbeatResp:

	}

}

func (r *raft) bcastHeartbeat() {
	r.bcastHeartbeatWithCtx(nil)
}

func (r *raft) bcastHeartbeatWithCtx(ctx []byte) {
	for id := range r.prs {
		if id == r.id {
			continue
		}
		r.sendHeartbeat(id, ctx)
	}
}

func (r *raft) sendHeartbeat(id uint64, ctx []byte) {
	commit := min(r.raftLog.commitIndex, r.prs[id].Match)

	m := Message{
		To:          id,
		Type:        MsgHearbeat,
		CommitIndex: commit,
		Context:     ctx,
	}

	r.send(m)
}

func (r *raft) bcastAppend() {
	for id := range r.prs {
		if id == r.id {
			continue
		}
		r.sendAppend(id)
	}
}

func (r *raft) sendAppend(to uint64) {
	// 获取上一条日志任期和 index
	pr := r.prs[to]
	term, _ := r.raftLog.term(pr.Next - 1)
	ents, _ := r.raftLog.entries(pr.Next)
	m := Message{
		To:          to,
		Type:        MsgApp,
		LogTerm:     term,
		LogIndex:    pr.Next - 1,
		Entries:     ents,
		CommitIndex: r.raftLog.commitIndex,
	}
	r.send(m)
}
