package FKPipeline

import (
	"FKBase"
	"FKConfig"
	"FKLog"
	"FKSpider"
	"FKStatus"
	"FKTempDataPool"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// 全局支持的输出方式
	G_DataOutput = make(map[string]func(self *pipeline) error)

	// 全局支持的文本数据输出方式名称列表
	G_DataOutputLib []string
)

// 结果收集与输出
type pipeline struct {
	*FKSpider.Spider                              //绑定的采集规则
	DataChan         chan FKTempDataPool.DataCell //文本数据收集通道
	FileChan         chan FKTempDataPool.FileCell //文件收集通道
	dataDocker       []FKTempDataPool.DataCell    //分批输出结果缓存
	outType          string                       //输出方式
	dataBatch        uint64                       //当前文本输出批次
	fileBatch        uint64                       //当前文件输出批次
	wait             sync.WaitGroup
	sum              [4]uint64 //收集的数据总数[上次输出后文本总数，本次输出后文本总数，上次输出后文件总数，本次输出后文件总数]，非并发安全
	dataSumLock      sync.RWMutex
	fileSumLock      sync.RWMutex
}

func (self *pipeline) CollectData(dataCell FKTempDataPool.DataCell) error {
	var err error
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("输出协程已终止")
		}
	}()
	self.DataChan <- dataCell
	return err
}

func (self *pipeline) CollectFile(fileCell FKTempDataPool.FileCell) error {
	var err error
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("输出协程已终止")
		}
	}()
	self.FileChan <- fileCell
	return err
}

// 停止
func (self *pipeline) Stop() {
	go func() {
		defer func() {
			recover()
		}()
		close(self.DataChan)
	}()
	go func() {
		defer func() {
			recover()
		}()
		close(self.FileChan)
	}()
}

// 启动数据收集/输出管道
func (self *pipeline) Start() {
	// 启动输出协程
	go func() {
		dataStop := make(chan bool)
		fileStop := make(chan bool)

		go func() {
			defer func() {
				recover()
			}()
			for data := range self.DataChan {
				// 缓存分批数据
				self.dataDocker = append(self.dataDocker, data)

				// 未达到设定的分批量时继续收集数据
				if len(self.dataDocker) < FKStatus.GlobalRuntimeTaskConfig.DockerCap {
					continue
				}

				// 执行输出
				self.dataBatch++
				self.outputData()
			}
			// 将剩余收集到但未输出的数据输出
			self.dataBatch++
			self.outputData()
			close(dataStop)
		}()

		go func() {
			defer func() {
				recover()
				// println("FileChanStop$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			}()
			// 只有当收到退出通知并且通道内无数据时，才退出循环
			for file := range self.FileChan {
				atomic.AddUint64(&self.fileBatch, 1)
				self.wait.Add(1)
				go self.outputFile(file)
			}
			close(fileStop)
		}()

		<-dataStop
		<-fileStop

		// 等待所有输出完成
		self.wait.Wait()

		// 返回报告
		self.Report()
	}()
}

func (self *pipeline) resetDataDocker() {
	for _, cell := range self.dataDocker {
		FKTempDataPool.PutDataCell(cell)
	}
	self.dataDocker = self.dataDocker[:0]
}

// 获取文本数据总量
func (self *pipeline) dataSum() uint64 {
	self.dataSumLock.RLock()
	defer self.dataSumLock.RUnlock()
	return self.sum[1]
}

// 更新文本数据总量
func (self *pipeline) addDataSum(add uint64) {
	self.dataSumLock.Lock()
	defer self.dataSumLock.Unlock()
	self.sum[0] = self.sum[1]
	self.sum[1] += add
}

// 获取文件数据总量
func (self *pipeline) fileSum() uint64 {
	self.fileSumLock.RLock()
	defer self.fileSumLock.RUnlock()
	return self.sum[3]
}

// 更新文件数据总量
func (self *pipeline) addFileSum(add uint64) {
	self.fileSumLock.Lock()
	defer self.fileSumLock.Unlock()
	self.sum[2] = self.sum[3]
	self.sum[3] += add
}

// 返回报告
func (self *pipeline) Report() {
	FKStatus.GlobalRuntimeReportChan <- &FKStatus.AppRuntimeReport{
		SpiderName: self.Spider.GetName(),
		Keyword:    self.GetKeywords(),
		DataNum:    self.dataSum(),
		FileNum:    self.fileSum(),
		Time:       time.Since(FKStatus.GlobalAppStartTime),
	}
}

// 主命名空间相对于数据库名，不依赖具体数据内容，可选
func (self *pipeline) namespace() string {
	if self.Spider.Namespace == nil {
		if self.Spider.GetSubName() == "" {
			return self.Spider.GetName()
		}
		return self.Spider.GetName() + "__" + self.Spider.GetSubName()
	}
	return self.Spider.Namespace(self.Spider)
}

// 次命名空间相对于表名，可依赖具体数据内容，可选
func (self *pipeline) subNamespace(dataCell map[string]interface{}) string {
	if self.Spider.SubNamespace == nil {
		return dataCell["RuleName"].(string)
	}
	defer func() {
		if p := recover(); p != nil {
			FKLog.G_Log.Error("subNamespace: %v", p)
		}
	}()
	return self.Spider.SubNamespace(self.Spider, dataCell)
}

// 下划线连接主次命名空间
func joinNamespaces(namespace, subNamespace string) string {
	if namespace == "" {
		return subNamespace
	} else if subNamespace != "" {
		return namespace + "__" + subNamespace
	}
	return namespace
}

// 文本数据输出
func (self *pipeline) outputData() {
	defer func() {
		// 回收缓存块
		self.resetDataDocker()
	}()

	// 输出
	dataLen := uint64(len(self.dataDocker))
	if dataLen == 0 {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			FKLog.G_Log.Informational(" * ")
			FKLog.G_Log.App(" *     Panic  [数据输出：%v | KEYWORDS：%v | 批次：%v]   数据 %v 条！ [ERROR]  %v",
				self.Spider.GetName(), self.Spider.GetKeywords(), self.dataBatch, dataLen, p)
		}
	}()

	// 输出统计
	self.addDataSum(dataLen)

	//FKLog.G_Log.Informational("G_DataOutput len = %d", len(G_DataOutput))
	// 执行输出
	err := G_DataOutput[self.outType](self)

	FKLog.G_Log.Informational(" * ")
	if err != nil {
		FKLog.G_Log.App(" *     Fail  [数据输出：%v | KEYWORDS：%v | 批次：%v]   数据 %v 条！ [ERROR]  %v",
			self.Spider.GetName(), self.Spider.GetKeywords(), self.dataBatch, dataLen, err)
	} else {
		FKLog.G_Log.App(" *     [数据输出：%v | KEYWORDS：%v | 批次：%v]   数据 %v 条！",
			self.Spider.GetName(), self.Spider.GetKeywords(), self.dataBatch, dataLen)
		self.Spider.TryFlushSuccess()
	}
}

// 文件输出
func (self *pipeline) outputFile(file FKTempDataPool.FileCell) {
	// 复用FileCell
	defer func() {
		FKTempDataPool.PutFileCell(file)
		self.wait.Done()
	}()

	// 路径： file/"RuleName"/"time"/"Name"
	p, n := filepath.Split(filepath.Clean(file["Name"].(string)))
	dir := filepath.Join(FKConfig.CONFIG_FILE_OUT_DIR_PATH, FKBase.ReplaceSignToChineseSign(self.namespace()), p)

	// 文件名
	fileName := filepath.Join(dir, FKBase.ReplaceSignToChineseSign(n))

	// 创建/打开目录
	d, err := os.Stat(dir)
	if err != nil || !d.IsDir() {
		if err := os.MkdirAll(dir, 0777); err != nil {
			FKLog.G_Log.Error(
				" *     Fail  [文件下载：%v | KEYWORDS：%v | 批次：%v]   %v [ERROR]  %v",
				self.Spider.GetName(), self.Spider.GetKeywords(), atomic.LoadUint64(&self.fileBatch), fileName, err,
			)
			return
		}
	}

	// 文件不存在就以0777的权限创建文件，如果存在就在写入之前清空内容
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		FKLog.G_Log.Error(
			" *     Fail  [文件下载：%v | KEYWORDS：%v | 批次：%v]   %v [ERROR]  %v",
			self.Spider.GetName(), self.Spider.GetKeywords(), atomic.LoadUint64(&self.fileBatch), fileName, err,
		)
		return
	}

	size, err := io.Copy(f, bytes.NewReader(file["Bytes"].([]byte)))
	f.Close()
	if err != nil {
		FKLog.G_Log.Error(
			" *     Fail  [文件下载：%v | KEYWORDS：%v | 批次：%v]   %v (%s) [ERROR]  %v",
			self.Spider.GetName(), self.Spider.GetKeywords(), atomic.LoadUint64(&self.fileBatch), fileName,
			FKBase.GlobalBytes.FormatUintBytesToString(size), err,
		)
		return
	}

	// 输出统计
	self.addFileSum(1)

	// 打印报告
	FKLog.G_Log.Informational(" * ")
	FKLog.G_Log.App(
		" *     [文件下载：%v | KEYWORDS：%v | 批次：%v]   %v (%s)",
		self.Spider.GetName(), self.Spider.GetKeywords(), atomic.LoadUint64(&self.fileBatch), fileName,
		FKBase.GlobalBytes.FormatUintBytesToString(size),
	)
	FKLog.G_Log.Informational(" * ")
}
