package raft

type unstable struct {
	// 未持久化的日志
	entries []Entry
	// 第一笔未持久化的日志的索引
	offset uint64
}

func (u *unstable) maybeLastIndex() (uint64, bool) {
	if l := len(u.entries); l != 0 {
		return u.offset + uint64(l) - 1, true
	}
	return 0, false
}

func (u *unstable) maybeTerm(i uint64) (uint64, bool) {
	if i < u.offset {
		return 0, false
	}

	last, ok := u.maybeLastIndex()
	if !ok {
		return 0, false
	}

	if i > last {
		return 0, false
	}

	return u.entries[i-u.offset].Term, true
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

func newRaftLog(storage Storage) *raftLog {
	if storage == nil {
		panic("storage must not be nil")
	}

	r := raftLog{
		storage: storage,
	}

	firstIndex, err := storage.FirstIndex()
	if err != nil {
		panic(err)
	}

	lastIndex, err := storage.LastIndex()
	if err != nil {
		panic(err)
	}

	r.unstable.offset = lastIndex + 1
	r.commitIndex = firstIndex - 1
	r.applyIndex = firstIndex - 1
	return &r
}

func (r *raftLog) unstableEntries() []Entry {
	return r.unstable.entries
}

// 获取还未应用但是已提交的 entries
func (r *raftLog) nextEnts() []Entry {
	off := max(r.applyIndex+1, r.firstIndex())
	if r.commitIndex > off {
		ents, err := r.slice(off, r.commitIndex+1)
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

func (r *raftLog) slice(left, right uint64) ([]Entry, error) {
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

func (r *raftLog) appliedTo(i uint64) {
	if i == 0 {
		return
	}
	if r.commitIndex < i || i < r.applyIndex {
		panic("invalid apply index")
	}
	r.applyIndex = i
}

// 返回从 i 开始的日志
func (r *raftLog) entries(i uint64) ([]Entry, error) {
	if i > r.lastIndex() {
		return nil, nil
	}
	return r.slice(i, r.lastIndex()+1)
}

func (r *raftLog) maybeCommit(maxIndex, term uint64) bool {
	if maxIndex > r.commitIndex && r.zeroTermOnErrCompacted(r.term(maxIndex)) == term {
		r.commitTo(maxIndex)
		return true
	}
	return false
}

func (r *raftLog) term(i uint64) (uint64, error) {
	dummyIndex := r.firstIndex() - 1
	if i < dummyIndex || i > r.lastIndex() {
		return 0, nil
	}

	// 非持久化中取

	// 持久化中取
	t, err := r.storage.Term(i)
	if err == nil {
		return t, nil
	}

	if err == ErrCompacted || err == ErrUnavailable {
		return 0, err
	}

	panic(err)
}

// 修改 commit 索引
func (r *raftLog) commitTo(tocommit uint64) {
	if r.commitIndex >= tocommit {
		return
	}

	if r.lastIndex() < tocommit {
		panic("commit index over last index")
	}

	r.commitIndex = tocommit
}

func (r *raftLog) zeroTermOnErrCompacted(t uint64, err error) uint64 {
	if err == nil {
		return t
	}
	if err == ErrCompacted {
		return 0
	}

	panic(err)
}
