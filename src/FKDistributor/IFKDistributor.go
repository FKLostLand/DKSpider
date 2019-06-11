package FKDistributor

// 分布式对象
type Distributor interface {
	// 主节点从仓库发送一个任务
	Send(clientNum int) DistributeTask
	// 从节点接收一个任务到仓库
	Receive(task *DistributeTask)
	// 返回与之连接的节点数
	CountNodes() int
}
