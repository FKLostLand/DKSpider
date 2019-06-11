package FKPipeline

import (
	"FKBase"
	"FKLog"
	"FKMySQL"
	"fmt"
	"sync"
)

func init() {
	var (
		mysqlTable     = map[string]*FKMySQL.FKSQLTable{}
		mysqlTableLock sync.RWMutex
	)

	var getMysqlTable = func(name string) (*FKMySQL.FKSQLTable, bool) {
		mysqlTableLock.RLock()
		defer mysqlTableLock.RUnlock()
		tab, ok := mysqlTable[name]
		if ok {
			return tab.Clone(), true
		}
		return nil, false
	}

	var setMysqlTable = func(name string, tab *FKMySQL.FKSQLTable) {
		mysqlTableLock.Lock()
		mysqlTable[name] = tab
		mysqlTableLock.Unlock()
	}

	G_DataOutput["mysql"] = func(self *pipeline) error {
		_, err := FKMySQL.DB()
		if err != nil {
			return fmt.Errorf("Mysql数据库链接失败:  %v", err)
		}
		var (
			mysqls    = make(map[string]*FKMySQL.FKSQLTable)
			namespace = FKBase.ReplaceSignToChineseSign(self.namespace())
		)
		for _, datacell := range self.dataDocker {
			subNamespace := FKBase.ReplaceSignToChineseSign(self.subNamespace(datacell))
			tName := joinNamespaces(namespace, subNamespace)
			table, ok := mysqls[tName]
			if !ok {
				table, ok = getMysqlTable(tName)
				if ok {
					mysqls[tName] = table
				} else {
					table = FKMySQL.CreateSQLTable()
					table.SetTableName(tName)
					for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
						table.AddColumn(title + ` MEDIUMTEXT`)
					}
					if self.Spider.OutDefaultField() {
						table.AddColumn(`Url VARCHAR(255)`, `ParentUrl VARCHAR(255)`, `DownloadTime VARCHAR(50)`)
					}
					if err := table.Create(); err != nil {
						FKLog.G_Log.Error("%v", err)
						continue
					} else {
						setMysqlTable(tName, table)
						mysqls[tName] = table
					}
				}
			}
			var data []string
			for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					data = append(data, v)
				} else {
					data = append(data, FKBase.JsonString(vd[title]))
				}
			}
			if self.Spider.OutDefaultField() {
				data = append(data, datacell["Url"].(string), datacell["ParentUrl"].(string), datacell["DownloadTime"].(string))
			}
			table.AutoInsert(data)
		}
		for _, tab := range mysqls {
			FKLog.CheckErr(tab.FlushInsert())
		}
		mysqls = nil
		return nil
	}
}
