package main

import "xiaoxuxiansheng/my-raft/raft"

func main() {

	// 用于提交写请求的 chan
	proposeC := make(chan string)
	// 用于提交配置变更的 chan
	confChangeC := make(chan raft.ConfChange)

	commitC := newRaftProxy(1, []string{}, proposeC, confChangeC)
	s := newService(proposeC, confChangeC, commitC)
	serveHttoAPI(8091, s)
}
