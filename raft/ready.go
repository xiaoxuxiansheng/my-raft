package raft

type Ready struct {
	// 软状态
	SoftState *SoftState

	// 硬状态
	HardState HardState

	// 读一致性
	ReadStates []ReadState

	// 发送消息前需要持久化的日志
	Entries []Entry

	// 已经提交，需要应用到状态机的数据
	CommittedEntries []Entry

	// 消息
	Message []Message
}

func newReady(r *raft, preSoft *SoftState, preHard HardState) Ready {
	rd := Ready{
		// 尚未持久化，需要持久化的日志
		Entries: r.raftLog.unstableEntries(),
		// 需要提交的日志
		CommittedEntries: r.raftLog.nextEnts(),
		// 待发送的消息
		Message: r.msgs,
	}
	if soft := r.softState(); !soft.equal(preSoft) {
		rd.SoftState = soft
	}
	if hard := r.hardState(); !isHardStateEqual(hard, preHard) {
		rd.HardState = hard
	}
	if len(r.readStates) != 0 {
		rd.ReadStates = r.readStates
	}
	return rd
}

func (rd Ready) containsUpdates() bool {
	return rd.SoftState != nil || !IsEmptyHardState(rd.HardState) ||
		len(rd.Entries) > 0 || len(rd.CommittedEntries) > 0 ||
		len(rd.Message) > 0 || len(rd.ReadStates) > 0
}

func IsEmptyHardState(h HardState) bool {
	return isHardStateEqual(h, emptyHardState)
}
