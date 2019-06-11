package FKDistributor

// 任务仓库
type DistributeTaskList struct {
	Tasks chan *DistributeTask
}

func CreateFKDistributeTaskList() *DistributeTaskList {
	return &DistributeTaskList{
		Tasks: make(chan *DistributeTask, 1024),
	}
}

// 服务器向仓库添加一个任务
func (l *DistributeTaskList) Push(task *DistributeTask) {
	id := len(l.Tasks)
	task.Id = id
	l.Tasks <- task
}

// 客户端从本地仓库获取一个任务
func (l *DistributeTaskList) Pull() *DistributeTask {
	return <-l.Tasks
}

// 仓库任务总数
func (l *DistributeTaskList) Len() int {
	return len(l.Tasks)
}

// 主节点从仓库发送一个任务
func (self *DistributeTaskList) Send(clientNum int) DistributeTask {
	return *<-self.Tasks
}

// 从节点接收一个任务到仓库
func (self *DistributeTaskList) Receive(task *DistributeTask) {
	self.Tasks <- task
}
