package raft

type Storage interface {
	InitialState() (HardState, ConfState, error)
	// 根据日志索引获取日志，大小不超过 max
	Entries(l, r, max uint64) ([]Entry, error)
	// 通过日志索引获取对应的任期
	Term(i uint64) (uint64, error)
	// 最后一笔已持久化的日志索引
	LastIndex() (uint64, error)
	// 第一笔已持久化的日志索引
	FirstIndex() (uint64, error)
}
