package raft

type stepFunc func(*raft, Message)

type raft struct {
	// 当前节点 id
	id uint64
	// 任期号
	Term uint64
	// 读一致性
	readStates []ReadState
	// 日模块
	raftLog *raftLog
	// 各节点的进度
	prs map[uint64]*Progress
	// 节点状态
	state StateType
	// 存放了哪些节点投票给了本节点
	votes map[uint64]bool
	// 存放了要投放的消息
	msgs []Message
	// 记录了 leader 的节点 id
	lead uint64
	// 标识当前还存在未被应用的配置数据
	pendingConf bool
	// 全局读请求信息
	readOnly *readOnly
	// 是否处于预竞选状态
	preVote bool
	// tick 函数，定时器到期时需要执行的操作，不同角色的处理逻辑不同
	tick func()
	// 步入某个角色后需要执行的逻辑，不同角色的处理逻辑不同
	step stepFunc
	// 投票给哪个节点 id
	Vote uint64
	// 是否确认不处于小分区
	checkQuorum bool
}

func newRaft(conf *Config) *raft {
	return &raft{}
}

func (r *raft) Step(m Message) error {
	// 第一个 switch 对任期进行 dispatch
	switch {
	// 来自本地消息，直接放行
	case m.Term == 0:
	case m.Term > r.Term:
		lead := m.From
		// 消息任期更大
		if m.Type == MsgVote || m.Type == MsgPreVote {
			lead = None
		}

		if m.Type != MsgPreVote && (m.Type != MsgPreVoteResp || m.Reject) {
			r.becomeFollower(m.Term, lead)
		}

	case m.Term < r.Term:
		// 消息任期更小
		// 心跳或者同步日志请求，有义务告知其更新的任期
		if r.checkQuorum && (m.Type == MsgHearbeat || m.Type == MsgApp) {
			r.send(Message{To: m.From, Type: MsgAppResp})
		}
		// 任期更小的消息一定无需理会
		return nil
	}

	// 第二个 switch 对消息类型 dispatch
	switch m.Type {
	// 推进本节点参与选举
	case MsgHup:
		// 已经是 leader 了，直接跳过
		if r.state == StateLeader {
			break
		}
		// 如果还有已提交未应用的配置变更消息，则不能发起投票
		ents, err := r.raftLog.slice(r.raftLog.applyIndex+1, r.raftLog.commitIndex+1, noLimit)
		if err != nil {
			panic(err)
		}

		if n := numOfPendingConf(ents); n > 0 {
			break
		}

		// 进行选举
		if r.preVote {
			// 进行预选举
			r.campaign(campaignPreElection)
			break
		}
		r.campaign(campaignElection)

	// 接收到号票消息
	case MsgVote, MsgPreVote:
		// 收到了任期大于等于自身的号票消息
		// 如果对方的数据比自己新并且后 3 个条件满足其一（1）当前节点没投过票、（2）竞选任期大于自身、（3）对方是自己之前的投票对象
		if r.raftLog.isUpToDate(m.LogIndex, m.LogTerm) && (r.Vote == None || m.Term > r.Term || m.From == r.Vote) {
			if m.Type == MsgVote {
				r.Vote = m.From
				r.send(Message{Type: MsgVoteResp, To: m.From})
				break
			}
			r.send(Message{Type: MsgPreVoteResp, To: m.From})
			break
		}
		if m.Type == MsgVote {
			r.send(Message{Type: MsgVoteResp, To: m.From, Reject: true})
			break
		}
		r.send(Message{Type: MsgPreVoteResp, To: m.From, Reject: true})

	// 其他情形走入定制化的状态机函数
	default:
		r.step(r, m)
	}

	return nil
}

func (r *raft) becomeFollower(term, lead uint64) {
	r.reset(term)
	r.step = stepFollower
	r.tick = tickElection
	r.lead = lead
	r.state = StateFollower
}

func (r *raft) reset(term uint64) {

}

func (r *raft) softState() *SoftState {
	return &SoftState{Lead: r.lead, RaftState: r.state}
}

func (r *raft) hardState() HardState {
	return HardState{
		Term:        r.Term,
		CommitIndex: r.raftLog.commitIndex,
		Vote:        r.Vote,
	}
}

func (r *raft) addNode(id uint64) {
	if _, ok := r.prs[id]; ok {
		return
	}
	r.prs[id] = &Progress{Match: 0, Next: r.raftLog.lastIndex() + 1}
}

func (r *raft) send(m Message) {
	if m.Type != MsgProp && m.Type != MsgReadIndex {
		m.Term = r.Term
	}

	r.msgs = append(r.msgs, m)
}

func (r *raft) campaign(typ CampaignType) {

}

func tickElection() {

}

func stepFollower(r *raft, m Message) {
}
