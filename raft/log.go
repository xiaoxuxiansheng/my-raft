package raft

type unstable struct {
	// 未持久化的日志
	entries []Entry
	// 第一笔未持久化的日志的索引
	offset uint64
}

func (u *unstable) mustCheckOutOfBounds(l, r uint64) {
	if l > r {
		panic("invalid unstable slice")
	}

	if l < u.offset || r > u.offset+uint64(len(u.entries)) {
		panic("invalid unstable slice")
	}
}

func (u *unstable) slice(l, r uint64) []Entry {
	u.mustCheckOutOfBounds(l, r)
	return u.entries[l-u.offset : r-u.offset]
}

// 在非持久化区域中传入 entries，可能导致原先数据的截断或者重叠追加
func (u *unstable) truncateAndAppend(ents []Entry) {
	after := ents[0].Index
	switch {
	case after == u.offset+uint64(len(u.entries)):
		u.entries = append(u.entries, ents...)
	case after <= u.offset:
		u.offset = after
		u.entries = ents
	default:
		u.entries = append([]Entry{}, u.slice(u.offset, after)...)
		u.entries = append(u.entries, ents...)
	}
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

func (u *unstable) stableTo(i, t uint64) {
	gt, ok := u.maybeTerm(i)
	if !ok {
		return
	}

	if t == gt && i >= u.offset {
		u.entries = u.entries[i-u.offset+1:]
		u.offset = i + 1
	}
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

func (r *raftLog) stableTo(i, t uint64) {
	r.unstable.stableTo(i, t)
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

func (r *raftLog) mustCheckOutOfBounds(lo, hi uint64) error {
	if lo > hi {
		panic("invalid raft log index")
	}

	fi := r.firstIndex()
	if lo < fi {
		panic("invalid raft log index")
	}

	if hi > r.lastIndex()+1 {
		panic("invalid raft log index")
	}

	return nil
}

func (r *raftLog) slice(lo, hi uint64) ([]Entry, error) {
	r.mustCheckOutOfBounds(lo, hi)
	if lo == hi {
		return nil, nil
	}

	var ents []Entry
	if lo < r.unstable.offset {
		entries, err := r.storage.Entries(lo, min(r.unstable.offset, hi))
		if err != nil {
			panic(err)
		}

		ents = append(ents, entries...)
	}

	if hi > r.unstable.offset {
		unstable := r.unstable.slice(max(lo, r.unstable.offset), hi)
		ents = append(ents, unstable...)
	}

	return ents, nil
}

// 此处的 append 只能添加非持久化的日志
func (r *raftLog) append(ents ...Entry) uint64 {
	if len(ents) == 0 {
		return r.lastIndex()
	}

	if after := ents[0].Index - 1; after < r.commitIndex {
		panic("entry index less then commit index")
	}

	r.unstable.truncateAndAppend(ents)
	return r.lastIndex()
}

func (r *raftLog) lastIndex() uint64 {
	if i, ok := r.unstable.maybeLastIndex(); ok {
		return i
	}

	i, err := r.storage.LastIndex()
	if err != nil {
		panic(err)
	}

	return i
}

func (r *raftLog) lastTerm() uint64 {
	t, err := r.term(r.lastIndex())
	if err != nil {
		panic(err)
	}
	return t
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
