package raft

func (r *raft) becomePreCandidate() {
	if r.state == StateLeader {
		panic("invalid transition leader -> pre-candidate")
	}
	r.step = stepCandidate
	r.tick = r.tickElection
	r.state = StatePreCandidate
}

func (r *raft) becomeCandidate() {
	if r.state == StateLeader {
		panic("invalid transition leader -> candidate")
	}
	r.step = stepCandidate
	r.reset(r.Term + 1)
	r.tick = r.tickElection
	// 候选人会先投给自己
	r.Vote = r.id
	r.state = StateCandidate
}

// candidate 的状态机函数
func stepCandidate(r *raft, m Message) {
	var voteRespType MessageType
	if r.state == StatePreCandidate {
		voteRespType = MsgPreVoteResp
	} else {
		voteRespType = MsgVoteResp
	}

	switch m.Type {
	case MsgProp:
		// candidate 没有负责提交写请求的义务
		return
	case MsgApp:
		// 收到任期更大的同步日志请求消息，需要退回 follower
		r.becomeFollower(r.Term, m.From)
		// 处理同步日志请求
		r.handleAppendEntries(m)
	case MsgHearbeat:
		r.becomeFollower(m.Term, m.From)
		// 通过心跳请求更新提交索引
		r.handleHeartbeat(m)
	case voteRespType:
		// 计算当前各个节点的投票
		granted := r.poll(m.From, !m.Reject)
		switch r.quorum() {
		// 赞同票达到了多数派
		case granted:
			if r.state == StatePreCandidate {
				r.campaign(campaignElection)
			} else {
				r.becomeLeader()
				// 成为 leader 后要广播一波消息，同步一下存量的日志，在 become leader 方法中，已经添加了一条当前任期内的空日志
				r.bcastAppend()
			}
		// 拒绝票达到了多数派
		case len(r.votes) - granted:
			// 退回 follower
			r.becomeFollower(r.Term, None)
		}
	}
}
