package raft

import "math"

const (
	None    uint64 = 0
	noLimit uint64 = math.MaxUint64
)

type EntryType int32

const (
	// 正常日志
	EntryNormal EntryType = 0
	// 配置变更日志
	EntryConfChange EntryType = 1
)

type Entry struct {
	Term  uint64    `json:"term"`
	Index uint64    `json:"index"`
	Type  EntryType `json:"type"`
	Data  []byte    `json:"data"`
}

type MessageType int32

const (
	// 推进本地节点选举
	MsgHup MessageType = 0
	// 用于 leader 推进自身进行心跳广播
	MsgBeat MessageType = 1
	// 用户向 raft 提交数据
	MsgProp MessageType = 2
	// leader 向集群其他节点同步数据
	MsgApp MessageType = 3
	// 其他节点向 leader 回复同步数据的请求
	MsgAppResp MessageType = 4
	// 投票
	MsgVote     MessageType = 5
	MsgVoteResp MessageType = 6
	// 心跳
	MsgHearbeat     MessageType = 7
	MsgHearbeatResp MessageType = 8
	// 读一致性
	MsgReadIndex     MessageType = 9
	MsgReadIndexResp MessageType = 10
	// 预投票
	MsgPreVote     MessageType = 11
	MsgPreVoteResp MessageType = 12
)

type Message struct {
	Type MessageType `json:"type"`
	To   uint64      `json:"to"`
	From uint64      `json:"from"`
	// 当前任期 term
	Term uint64 `json:"term"`
	// 上一条日志的任期
	LogTerm  uint64 `json:"logTerm"`
	LogIndex uint64 `json:"logIndex"`
	// 同步的日志
	Entries []Entry `json:"entries"`
	// leader 最晚提交日志的索引
	CommitIndex uint64 `json:"commitIndex"`
	// 是否拒绝
	Reject bool `json:"reject"`
	// 拒绝同步日志时，最晚的日志索引
	RejectHint uint `json:"rejectHint"`
	// 上下文
	Context []byte `json:"context"`
}

type StateType int32

const (
	// 跟随者
	StateFollower StateType = 0
	// 候选人
	StateCandidate StateType = 1
	// 领导者
	StateLeader StateType = 2
	// 预竞选
	StatePreCandidate StateType = 3
)

type SoftState struct {
	// 当前集群的 leader
	Lead uint64
	// 当前节点的状态
	RaftState StateType
}

func (s *SoftState) equal(pre *SoftState) bool {
	return s.Lead == pre.Lead && s.RaftState == pre.RaftState
}

var emptyHardState HardState

type HardState struct {
	// 当前的任期
	Term uint64 `json:""`
	// 当前任期把票投给了谁
	Vote uint64 `json:"vote"`
	// 已提交日志的 index
	CommitIndex uint64 `json:"commitIndex"`
}

func isHardStateEqual(a, b HardState) bool {
	return a.Term == b.Term && a.Vote == b.Vote && a.CommitIndex == b.CommitIndex
}

type ConfState struct {
	// 节点信息
	Nodes []uint64
}

type Config struct {
	// 当前节点 id
	ID uint64
	// 记录了其他节点的 id
	peers []uint64
	// 持久化存储接口
	Storage Storage
	// 已应用的日志索引
	Applied uint64
	// 是否处于预竞选状态
	Prevote bool
	// follower 发生选举超时的 tick
	ElectionTick int32
	// leader 发送心跳的 tick
	HearbeatTick int32
}

type Peer struct {
	// 节点 id
	ID uint64
	// 上下文信息
	Context []byte
}

func max(l, r uint64) uint64 {
	if l > r {
		return l
	}
	return r
}

func min(l, r uint64) uint64 {
	if l < r {
		return l
	}
	return r
}

type CampaignType string

const (
	campaignPreElection CampaignType = "prev"
	campaignElection    CampaignType = "nor"
)

type ConfChangeType int32

const (
	ConfChangeAddNode    ConfChangeType = 0
	ConfChangeRemoveNode ConfChangeType = 1
	ConfChangeUpdateNode ConfChangeType = 2
)

type ConfChange struct {
	ID      uint64         `json:"id"`
	Type    ConfChangeType `json:"type"`
	NodeID  uint64         `json:"nodeID"`
	Context []byte         `json:"context"`
}
