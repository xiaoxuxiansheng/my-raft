package raft

type unstable struct {
	// 未持久化的日志
	entries []Entry
	// 第一笔未持久化的日志的索引
	offset uint64
}

type raftLog struct {
	// 存储接口，提供了持久化日志的查询能力
	storage Storage
	// 未持久化的日志
	unstable unstable
	// 已提交的日志索引
	commitIndex uint64
	applyIndex  uint64
}

func (r *raftLog) unstableEntries() []Entry {
	return r.unstable.entries
}

// 获取还未应用但是已提交的 entries
func (r *raftLog) nextEnts() []Entry {
	off := max(r.applyIndex+1, r.firstIndex())
	if r.commitIndex > off {
		ents, err := r.slice(off, r.commitIndex+1, noLimit)
		if err != nil {
			panic(err)
		}
		return ents
	}
	return nil
}

func (r *raftLog) firstIndex() uint64 {
	index, err := r.storage.FirstIndex()
	if err != nil {
		panic(err)
	}
	return index
}

func (r *raftLog) slice(left, right, maxSize uint64) ([]Entry, error) {
	return nil, nil
}

func (r *raftLog) apend(ents ...Entry) uint64 {
	return 0
}

func (r *raftLog) lastIndex() uint64 {
	return 0
}

func (r *raftLog) lastTerm() uint64 {
	return 0
}

// 数据是否新于自身
func (r *raftLog) isUpToDate(index, term uint64) bool {
	return term > r.lastTerm() || (term == r.lastIndex() && index >= r.lastIndex())
}
