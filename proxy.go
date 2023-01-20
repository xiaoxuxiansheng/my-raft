package main

import (
	"context"
	"time"
	"xiaoxuxiansheng/my-raft/raft"
)

type raftProxy struct {
	// 用户提交写请求
	proposeC <-chan string
	// 用户提交配置变更请求
	confChangeC <-chan raft.ConfChange
	// 提交日志
	commitC chan<- *string
	// 客户端 id
	id uint64
	// 节点列表
	peers []string

	node raft.Node

	// 日志持久化模块
	storage raft.Storage
}

func newRaftProxy(id uint64, peers []string, proposeC <-chan string, confChangeC <-chan raft.ConfChange) <-chan *string {
	commitC := make(chan *string)

	r := raftProxy{
		proposeC:    proposeC,
		confChangeC: confChangeC,
		commitC:     commitC,
		id:          id,
		peers:       peers,
	}

	go r.run()
	return commitC
}

func (r *raftProxy) run() {

}

func (r *raftProxy) listen() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// 监听两个客户端提交请求的 chan
	go r.listenRequest()
	// 主干流程，监听 ready

	for {
		select {
		case <-ticker.C:
			r.node.Tick()

		case <-r.node.Ready():
			// 持久化硬状态和配置信息

			// 持久化日志

			// 发送消息

			// 应用已提交的日志

			// advance
			r.node.Advance()
		}

	}

}

func (r *raftProxy) listenRequest() {
	for {
		select {
		case prop, ok := <-r.proposeC:
			if !ok {
				return
			}
			r.node.Propose(context.Background(), []byte(prop))

		case cc, ok := <-r.confChangeC:
			if !ok {
				return
			}
			r.node.ProposeConfChange(context.Background(), cc)
		}
	}

}
