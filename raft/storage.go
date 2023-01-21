package raft

import (
	"errors"
	"sync"
)

var ErrCompacted = errors.New("requested index is unavailable due to compaction")

var ErrUnavailable = errors.New("request entry at index is unavailable")

type Storage interface {
	InitialState() (HardState, ConfState, error)
	// 根据日志索引获取日志，大小不超过 max
	Entries(l, r uint64) ([]Entry, error)
	// 通过日志索引获取对应的任期
	Term(i uint64) (uint64, error)
	// 最后一笔已持久化的日志索引
	LastIndex() (uint64, error)
	// 第一笔已持久化的日志索引
	FirstIndex() (uint64, error)
}

type MemoryStorage struct {
	sync.Mutex
	hardState HardState
	// 持久化的日志
	ents []Entry
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		ents: make([]Entry, 1),
	}
}

func (m *MemoryStorage) InitialState() (HardState, ConfState, error) {
	return m.hardState, ConfState{}, nil
}

func (m *MemoryStorage) Entries(l, r uint64) ([]Entry, error) {
	m.Lock()
	defer m.Unlock()

	offset := m.ents[0].Index
	if l <= offset {
		return nil, ErrCompacted
	}

	if r > m.lastIndex()+1 {
		return nil, ErrUnavailable
	}

	if len(m.ents) == 1 {
		return nil, ErrUnavailable
	}

	return m.ents[l-offset : r-offset], nil

}

func (m *MemoryStorage) Term(i uint64) (uint64, error) {
	m.Lock()
	defer m.Unlock()
	offset := m.ents[0].Index
	if i < offset {
		return 0, ErrCompacted
	}

	if int(i-offset) >= len(m.ents) {
		return 0, ErrUnavailable
	}

	return m.ents[i-offset].Term, nil
}

func (m *MemoryStorage) LastIndex() (uint64, error) {
	m.Lock()
	defer m.Unlock()
	return m.lastIndex(), nil
}

func (m *MemoryStorage) lastIndex() uint64 {
	return m.ents[0].Index + uint64(len(m.ents)) - 1
}

func (m *MemoryStorage) FirstIndex() (uint64, error) {
	m.Lock()
	defer m.Lock()
	return m.firstIndex(), nil
}

func (m *MemoryStorage) firstIndex() uint64 {
	return m.ents[0].Index + 1
}
