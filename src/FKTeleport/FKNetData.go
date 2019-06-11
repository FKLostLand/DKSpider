package FKTeleport

const (
	// 返回成功
	SUCCESS = 0
	// 返回失败
	FAILURE = -1
	// 返回非法请求
	LLLEGAL = -2
)

// 定义数据传输结构
type NetData struct {
	// 消息体
	Body interface{}
	// 操作代号
	Operation string
	// 发信节点uid
	From string
	// 收信节点uid
	To string
	// 返回状态
	Status int
	// 标识符
	Flag string
}

// 请求处理接口
type Handle interface {
	Process(*NetData) *NetData
}

// 每个API方法需判断status状态，并做相应处理
type API map[string]Handle

func CreateNetData(from, to, operation string, flag string, body interface{}) *NetData {
	return &NetData{
		From:      from,
		To:        to,
		Body:      body,
		Operation: operation,
		Status:    SUCCESS,
		Flag:      flag,
	}
}

// API中生成返回结果的方法
// OpAndToAndFrom[0]参数为空时，系统将指定与对端相同的操作符
// OpAndToAndFrom[1]参数为空时，系统将指定与对端为接收者
// OpAndToAndFrom[2]参数为空时，系统将指定自身为发送者
func ReturnData(body interface{}, OpAndToAndFrom ...string) *NetData {
	data := &NetData{
		Status: SUCCESS,
		Body:   body,
	}
	if len(OpAndToAndFrom) > 0 {
		data.Operation = OpAndToAndFrom[0]
	}
	if len(OpAndToAndFrom) > 1 {
		data.To = OpAndToAndFrom[1]
	}
	if len(OpAndToAndFrom) > 2 {
		data.From = OpAndToAndFrom[2]
	}
	return data
}

// 返回错误，receive建议为直接接收到的*NetData
func ReturnError(receive *NetData, status int, msg string, nodeuid ...string) *NetData {
	receive.Status = status
	receive.Body = msg
	receive.From = ""
	if len(nodeuid) > 0 {
		receive.To = nodeuid[0]
	} else {
		receive.To = ""
	}
	return receive
}
