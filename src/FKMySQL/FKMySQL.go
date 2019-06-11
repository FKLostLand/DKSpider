package FKMySQL

import (
	"FKConfig"
	"FKLog"
	"database/sql"
	"errors"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// sql转换结构体
type FKSQLTable struct {
	tableName        string        // 表名
	columnNames      [][2]string   // 标题字段
	rowsCount        int           // 当前缓存的待插入数据的行数
	args             []interface{} // 数据
	sqlCode          string        // SQL语句
	customPrimaryKey bool
	size             int // 内容大小的近似值
}

var (
	err              error
	db               *sql.DB
	syncOnceOpenSQL  sync.Once
	maxAllowedPacket = FKConfig.MYSQL_MAX_ALLOWED_PACKET - 1024
	maxConnChan      = make(chan bool, FKConfig.MYSQL_CONNECT_POOL_CAP) //最大执行数限制
)

func CreateSQLTable() *FKSQLTable {
	return &FKSQLTable{}
}

func DB() (*sql.DB, error) {
	return db, err
}

func Refresh() {
	syncOnceOpenSQL.Do(func() {
		dbSource := FKConfig.CONFIG_MYSQL_CONNECT_STRING+"/"+FKConfig.CONFIG_DATABASE_NAME+"?charset=utf8"
		db, err = sql.Open("mysql", dbSource)
		if err != nil {
			FKLog.G_Log.Error("Mysql %s：%v\n", dbSource, err)
			return
		}
		db.SetMaxOpenConns(FKConfig.MYSQL_CONNECT_POOL_CAP)
		db.SetMaxIdleConns(FKConfig.MYSQL_CONNECT_POOL_CAP)
	})
	if err = db.Ping(); err != nil {
		FKLog.G_Log.Error("Mysql %s：%v\n", err)
	}
}

func (t *FKSQLTable) Clone() *FKSQLTable {
	return &FKSQLTable{
		tableName:        t.tableName,
		columnNames:      t.columnNames,
		customPrimaryKey: t.customPrimaryKey,
	}
}

// 设置表名
func (t *FKSQLTable) SetTableName(name string) *FKSQLTable {
	t.tableName = wrapSqlKey(name)
	return t
}

// 设置表单列
func (t *FKSQLTable) AddColumn(names ...string) *FKSQLTable {
	for _, name := range names {
		name = strings.Trim(name, " ")
		idx := strings.Index(name, " ")
		t.columnNames = append(t.columnNames, [2]string{wrapSqlKey(name[:idx]), name[idx+1:]})
	}
	return t
}

// 设置主键的语句（可选）
func (t *FKSQLTable) CustomPrimaryKey(primaryKeyCode string) *FKSQLTable {
	t.AddColumn(primaryKeyCode)
	t.customPrimaryKey = true
	return t
}

// 生成"创建表单"的语句，执行前须保证SetTableName()、AddColumn()已经执行
func (t *FKSQLTable) Create() error {
	if len(t.columnNames) == 0 {
		return errors.New("ColumnNames shouldn't be empty")
	}
	t.sqlCode = `CREATE TABLE IF NOT EXISTS ` + t.tableName + " ("
	if !t.customPrimaryKey {
		t.sqlCode += `id INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT,`
	}
	for _, title := range t.columnNames {
		t.sqlCode += title[0] + ` ` + title[1] + `,`
	}
	t.sqlCode = t.sqlCode[:len(t.sqlCode)-1] + `) ENGINE=MyISAM DEFAULT CHARSET=utf8;`

	maxConnChan <- true
	defer func() {
		t.sqlCode = ""
		<-maxConnChan
	}()

	_, err := db.Exec(t.sqlCode)
	return err
}

// 清空表单，执行前须保证SetTableName()已经执行
func (t *FKSQLTable) Truncate() error {
	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	_, err := db.Exec(`TRUNCATE TABLE ` + t.tableName)
	return err
}

// 设置插入的1行数据
func (t *FKSQLTable) addRow(value []string) *FKSQLTable {
	for i, count := 0, len(value); i < count; i++ {
		t.args = append(t.args, value[i])
	}
	t.rowsCount++
	return t
}

// 智能插入数据，每次1行
func (t *FKSQLTable) AutoInsert(value []string) *FKSQLTable {
	if t.rowsCount > 100 {
		FKLog.CheckErr(t.FlushInsert())
		return t.AutoInsert(value)
	}
	var nsize int
	for _, v := range value {
		nsize += len(v)
	}
	if nsize > maxAllowedPacket {
		FKLog.G_Log.Error("%v", "packet for query is too large. Try adjusting the 'maxallowedpacket'variable in the 'config.ini'")
		return t
	}
	t.size += nsize
	if t.size > maxAllowedPacket {
		FKLog.CheckErr(t.FlushInsert())
		return t.AutoInsert(value)
	}
	return t.addRow(value)
}

//向sqlCode添加"插入数据"的语句，执行前须保证Create()、AutoInsert()已经执行
func (t *FKSQLTable) FlushInsert() error {
	if t.rowsCount == 0 {
		return nil
	}

	colCount := len(t.columnNames)
	if colCount == 0 {
		return nil
	}

	t.sqlCode = `INSERT INTO ` + t.tableName + `(`

	for _, v := range t.columnNames {
		t.sqlCode += v[0] + ","
	}

	t.sqlCode = t.sqlCode[:len(t.sqlCode)-1] + `) VALUES `

	blank := ",(" + strings.Repeat(",?", colCount)[1:] + ")"
	t.sqlCode += strings.Repeat(blank, t.rowsCount)[1:] + `;`

	defer func() {
		// 清空临时数据
		t.args = []interface{}{}
		t.rowsCount = 0
		t.size = 0
		t.sqlCode = ""
	}()

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()

	_, err := db.Exec(t.sqlCode, t.args...)
	return err
}

// 获取全部数据
func (t *FKSQLTable) SelectAll() (*sql.Rows, error) {
	if t.tableName == "" {
		return nil, errors.New("表名不能为空")
	}
	t.sqlCode = `SELECT * FROM ` + t.tableName + `;`

	maxConnChan <- true
	defer func() {
		<-maxConnChan
	}()
	return db.Query(t.sqlCode)
}

func wrapSqlKey(s string) string {
	return "`" + strings.Replace(s, "`", "", -1) + "`"
}
