package FKPipeline

import (
	"FKBase"
	"FKKafka"
	"FKLog"
	"fmt"
	"regexp"
	"sync"
)

func init() {
	var (
		kafkaSenders    = map[string]*FKKafka.KafkaSender{}
		kafkaSenderLock sync.RWMutex
	)

	var getKafkaSender = func(name string) (*FKKafka.KafkaSender, bool) {
		kafkaSenderLock.RLock()
		tab, ok := kafkaSenders[name]
		kafkaSenderLock.RUnlock()
		return tab, ok
	}

	var setKafkaSender = func(name string, tab *FKKafka.KafkaSender) {
		kafkaSenderLock.Lock()
		kafkaSenders[name] = tab
		kafkaSenderLock.Unlock()
	}

	var topic = regexp.MustCompile("^[0-9a-zA-Z_-]+$")

	G_DataOutput["kafka"] = func(self *pipeline) error {
		_, err := FKKafka.GetProducer()
		if err != nil {
			return fmt.Errorf("kafka producer失败: %v", err)
		}
		var (
			kafkas    = make(map[string]*FKKafka.KafkaSender)
			namespace = FKBase.ReplaceSignToChineseSign(self.namespace())
		)
		for _, datacell := range self.dataDocker {
			subNamespace := FKBase.ReplaceSignToChineseSign(self.subNamespace(datacell))
			topicName := joinNamespaces(namespace, subNamespace)
			if !topic.MatchString(topicName) {
				FKLog.G_Log.Error("topic格式要求'^[0-9a-zA-Z_-]+$'，当前为：%s", topicName)
				continue
			}
			sender, ok := kafkas[topicName]
			if !ok {
				sender, ok = getKafkaSender(topicName)
				if ok {
					kafkas[topicName] = sender
				} else {
					sender = FKKafka.CreateKafkaSender()
					sender.SetTopic(topicName)
					setKafkaSender(topicName, sender)
					kafkas[topicName] = sender
				}
			}
			data := make(map[string]interface{})
			for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					data[title] = v
				} else {
					data[title] = FKBase.JsonString(vd[title])
				}
			}
			if self.Spider.OutDefaultField() {
				data["url"] = datacell["Url"].(string)
				data["parent_url"] = datacell["ParentUrl"].(string)
				data["download_time"] = datacell["DownloadTime"].(string)
			}
			err := sender.Push(data)
			FKLog.CheckErr(err)
		}
		kafkas = nil
		return nil
	}
}
