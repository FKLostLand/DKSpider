package FKHistory

import (
	"FKBase"
	"FKConfig"
	"FKRequest"
)

const (
	SUCCESS_SUFFIX = FKConfig.HISTORY_TAG + "__y"
	FAILURE_SUFFIX = FKConfig.HISTORY_TAG + "__n"
	SUCCESS_FILE   = FKConfig.HISTORY_DIR_PATH + "/" + SUCCESS_SUFFIX
	FAILURE_FILE   = FKConfig.HISTORY_DIR_PATH + "/" + FAILURE_SUFFIX
)

type (
	Historier interface {
		ReadSuccess(provider string, inherit bool) // 读取成功记录
		UpsertSuccess(string) bool                 // 更新或加入成功记录
		HasSuccess(string) bool                    // 检查是否存在某条成功记录
		DeleteSuccess(string)                      // 删除成功记录
		FlushSuccess(provider string)              // I/O输出成功记录，但不清缓存

		ReadFailure(provider string, inherit bool)      // 取出失败记录
		PullFailureList() map[string]*FKRequest.Request // 拉取失败记录并清空
		UpsertFailure(*FKRequest.Request) bool          // 更新或加入失败记录
		DeleteFailure(*FKRequest.Request)               // 删除失败记录
		FlushFailure(provider string)                   // I/O输出失败记录，但不清缓存

		Empty() // 清空缓存，但不输出
	}
)

func CreateHistorier(name string, subName string) Historier {
	successTabName := SUCCESS_SUFFIX + "__" + name
	successFileName := SUCCESS_FILE + "__" + name
	failureTabName := FAILURE_SUFFIX + "__" + name
	failureFileName := FAILURE_FILE + "__" + name
	if subName != "" {
		successTabName += "__" + subName
		successFileName += "__" + subName
		failureTabName += "__" + subName
		failureFileName += "__" + subName
	}
	return &History{
		Success: &Success{
			tabName:  FKBase.ReplaceSignToChineseSign(successTabName),
			fileName: successFileName,
			new:      make(map[string]bool),
			old:      make(map[string]bool),
		},
		Failure: &Failure{
			tabName:  FKBase.ReplaceSignToChineseSign(failureTabName),
			fileName: failureFileName,
			list:     make(map[string]*FKRequest.Request),
		},
	}
}
