package main

import "xiaoxuxiansheng/my-raft/raft"

func main() {
	// 用于提交写请求的 chan
	proposeC := make(chan string)
	// 用于提交配置变更的 chan
	confChangeC := make(chan raft.ConfChange)

	// 创建 raft 集群客户端，获取提交数据的 chan
	commitC := newRaftProxy(1, []string{}, proposeC, confChangeC)
	// 创建 kv 存储应用
	kvStore := newKVStore(proposeC, commitC)

	// 启动 http 服务
	s := newService(kvStore, proposeC, confChangeC)
	serveHTTPAPI(8091, s)
}
