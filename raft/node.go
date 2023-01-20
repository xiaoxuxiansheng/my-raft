package raft

import (
	"context"
	"encoding/json"
)

type Node struct {
	// 本地提交数据使用的 chan
	proc chan Message
	// 其他节点提交数据使用的 chan
	recvc chan Message
	// 接受配置变更的 chan
	confc chan Message
	// 传送 raft 节点的处理结果
	readyc chan Ready
	// 推进算法流程向前处理
	advancec chan struct{}
	// 定时器
	tickc chan struct{}
}

func StartNode(conf *Config, peers []Peer) Node {
	r := newRaft(conf)
	r.becomeFollower(1, None)
	// 将所有节点添加到配置中
	// for _, peer := range peers {

	// }
	r.raftLog.commitIndex = r.raftLog.lastIndex()

	// 添加各节点状态信息
	for _, peer := range peers {
		r.addNode(peer.ID)
	}

	n := newNode()
	go n.run(r)
	return n
}

func newNode() Node {
	return Node{
		proc:   make(chan Message),
		recvc:  make(chan Message),
		confc:  make(chan Message),
		readyc: make(chan Ready),
		tickc:  make(chan struct{}),
	}
}

func (n *Node) run(r *raft) {
	var (
		propc    chan Message
		readyc   chan Ready
		advancec chan struct{}
		// 状态信息
		rd       Ready
		prevSoft = r.softState()
		prevHard = emptyHardState
	)

	for {
		// 判断是否有数据更新，如果有的话会取 readyc 进行投递
		if advancec != nil {
			readyc = nil
			// 构造一个新的数据，如果有更新，就投递到 readyc 当中
		} else if rd = newReady(r, prevSoft, prevHard); rd.containsUpdates() {
			readyc = n.readyc
		} else {
			readyc = nil
		}

		select {
		// 接收到用户提交的数据
		case m := <-propc:
			// 处理本地提交的消息
			m.From = r.id
			r.Step(m)
		case m := <-n.recvc:
			// 非集群内节点的消息直接忽略
			if _, ok := r.prs[m.From]; !ok || !IsResponseMsg(m.Type) {
				break
			}
			r.Step(m)
		case <-n.confc:
		// 定时器到点
		case <-n.tickc:
			r.tick()
			// 有数据需要投递
		case readyc <- rd:
		case <-advancec:
		}
	}
}

func (n *Node) Tick() {
	select {
	case n.tickc <- struct{}{}:
	default:
	}
}

func (n *Node) Campaign(ctx context.Context) error {
	return n.step(ctx, Message{Type: MsgHup})
}

func (n *Node) Propose(ctx context.Context, data []byte) error {
	return n.step(ctx, Message{Type: MsgProp, Entries: []Entry{{Data: data}}})
}

func (n *Node) ProposeConfChange(ctx context.Context, cc ConfChange) error {
	data, _ := json.Marshal(cc)

	return n.step(ctx, Message{Type: MsgProp, Entries: []Entry{{Data: data}}})
}

func (n *Node) Ready() <-chan Ready {
	return n.readyc
}

func (n *Node) Advance() {
	n.advancec <- struct{}{}
}

func (n *Node) step(ctx context.Context, m Message) error {
	ch := n.recvc
	if m.Type == MsgProp {
		ch = n.proc
	}

	select {
	case ch <- m:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
