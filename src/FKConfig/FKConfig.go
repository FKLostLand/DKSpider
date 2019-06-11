package FKConfig

import (
	"FKBase"
	"FKStatus"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// 不可调整的默认配置项
const (
	APP_VERSION           string = "1.0.0"                                                      // 版本号
	APP_AUTHOR            string = "FeeKnight"                                                  // 作者
	APP_NAME              string = "FKAutoSpiderPool"                                           // 软件名
	APP_FULL_NAME         string = APP_NAME + " V" + APP_VERSION + " ( by " + APP_AUTHOR + " )" // 软件名全称
	APP_ICON_PNG          string = "iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAD/SURBVEhL7ZUtjgJBEIUHBRIHjnCJVTjEKo4Bd8OgUBjEsslKzrAEQ9gzLO/NdCXVNf0z02P5ki+prn7VJSBQvenLJ/yF/4Vylm9EucHQYB+5JIqESsnO68A3vDSlVxPJ6Tyx5xY68OUkuiaS03lizy2ygQydFww1SihcYpRsYCglCzbw4WSdpGTBHcoc6yR9FozgHsqMyB7vgkgohX6six7BpkEP0y0UdtDeewSbEfihMrusTw2s2eNdkC4LJCNOoDCG9t4j9XMt2P4CCnNo7z1SfziWH8j+uj41rCB71/pUiF5Kz3AGp/Doetre2AdyDoLfmgP8g094gh/QUVUvMGrSh1mUbY0AAAAASUVORK5CYII="                                                           // ICON https://www.base64-image.de/
	WORK_ROOT_DIR         string = "workdir"                                                    // 工作路径
	CONFIG_FILE_PATH      string = WORK_ROOT_DIR + "/config.ini"                                // 配置文件路径
	CACHE_DIR_PATH        string = WORK_ROOT_DIR + "/cache"                                     // 缓存文件路径
	LOG_DIR_PATH          string = WORK_ROOT_DIR + "/log"                                       // 日志文件路径
	DEFAULT_LOG_FILE_PATH string = LOG_DIR_PATH + "/default.log"                                // 默认日志文件路径
	PHANTOM_JS_CACHE_PATH string = CACHE_DIR_PATH + "/js"                                       // Phantom的JS文件临时目录
	HISTORY_TAG           string = "history"                                                    // 历史记录标示符
	HISTORY_DIR_PATH      string = WORK_ROOT_DIR + "/history"                                   // 历史记录下载目录
	SPIDER_HTML_EXT       string = ".fkasp.html"                                                // 动态规则扩展名
	IS_ASYNC_LOG          bool   = true                                                         // 是否异步日志输出
)

// 可脚本配置文件修改的默认配置项
const (
	CRAWL_CAP                int    = 50                           // 蜘蛛池最大容量
	LOG_CAP                  int64  = 10000                        // 日志缓存最大容量
	LOG_PRINT_LEVEL          string = "debug"                      // 打印日志输出级别
	LOG_CONSOLE_LEVEL        string = "info"                       // 控制台日志输出级别
	LOG_TO_BACKEND_LEVEL     string = "error"                      // 客户端反馈到服务器的日志级别
	IS_LOG_LINE_INFO         bool   = false                        // 日志是否打印行信息
	IS_LOG_SAVE_TO_LOCAL     bool   = true                         // 是否本地保存日志
	PHANTOM_JS_PATH          string = WORK_ROOT_DIR + "/phantom"   // PhantomJS所在文件路径
	PROXY_LIB_FILE_PATH      string = WORK_ROOT_DIR + "/proxy.lib" // 代理IP文件所在文件路径
	SPIDER_DIR_PATH          string = WORK_ROOT_DIR + "/sprider"   // 动态规则目录路径
	FILE_OUT_DIR_PATH        string = WORK_ROOT_DIR + "/files"     // 文件结果的输出目录（HTML, 图片等）
	TEXT_OUT_DIR_PATH        string = WORK_ROOT_DIR + "/texts"     // 文本结果的输出目录（EXCEL, CSV等）
	DATABASE_NAME            string = "FKSpiderDB"                 // 数据库名
	MYSQL_CONNECT_STRING     string = "root:@tcp(127.0.0.1:3306)"  // SQL连接字符串
	MYSQL_CONNECT_POOL_CAP   int    = 2048                         // SQL连接池容量
	MYSQL_MAX_ALLOWED_PACKET int    = 1048576                      // SQL通讯缓冲区最大长度，1MB
	KAFKA_BROKERS_STRING     string = "127.0.0.1:9002"             // KAFKA BROKER字符串（可都好分割）
)

// 配置文件的配置项
var (
	CONFIG_CRAWL_CAP                int    = GlobalSetting.DefaultInt("全局::蜘蛛池最大容量", CRAWL_CAP)
	CONFIG_PHANTOM_JS_PATH          string = GlobalSetting.String("全局::PhantomJS文件路径")
	CONFIG_PROXY_LIB_FILE_PATH      string = GlobalSetting.String("全局::代理IP文件路径")
	CONFIG_SPIDER_DIR_PATH          string = GlobalSetting.String("全局::动态规则目录路径")
	CONFIG_FILE_OUT_DIR_PATH        string = GlobalSetting.String("全局::HTML图片输出目录路径")
	CONFIG_TEXT_OUT_DIR_PATH        string = GlobalSetting.String("全局::文本文件输出目录路径")
	CONFIG_DATABASE_NAME            string = GlobalSetting.String("全局::数据库名")
	CONFIG_MYSQL_CONNECT_STRING     string = GlobalSetting.String("MYSQL::连接字符串")
	CONFIG_MYSQL_CONNECT_POOL_CAP   int    = GlobalSetting.DefaultInt("MYSQL::连接池容量", MYSQL_CONNECT_POOL_CAP)
	CONFIG_MYSQL_MAX_ALLOWED_PACKET int    = GlobalSetting.DefaultInt("MYSQL::通讯缓冲区最大长度", MYSQL_MAX_ALLOWED_PACKET)
	CONFIG_KAFKA_BROKERS_STRING     string = GlobalSetting.DefaultString("KAFKA::brokers字符串", KAFKA_BROKERS_STRING)
	CONFIG_LOG_CAP                  int64  = GlobalSetting.DefaultInt64("日志::缓存最大容量", LOG_CAP)
	CONFIG_LOG_PRINT_LEVEL          int    = FKBase.LogLevelStringToInt(GlobalSetting.String("日志::打印日志输出级别"))
	CONFIG_LOG_CONSOLE_LEVEL        int    = FKBase.LogLevelStringToInt(GlobalSetting.String("日志::控制台日志输出级别"))
	CONFIG_LOG_TO_BACKEND_LEVEL     int    = FKBase.LogLevelStringToInt(GlobalSetting.String("日志::反馈给服务器的日志级别"))
	CONFIG_IS_LOG_LINE_INFO         bool   = GlobalSetting.DefaultBool("日志::是否打印行信息", IS_LOG_LINE_INFO)
	CONFIG_IS_LOG_SAVE_TO_LOCAL     bool   = GlobalSetting.DefaultBool("日志::是否保存本地日志", IS_LOG_SAVE_TO_LOCAL)
)

// 可UI修改的默认配置项
const (
	APP_MODE               int    = FKStatus.UNSET // 本节点的角色
	MASTER_SERVER_IP       string = "127.0.0.1"    // 主节点IP
	MASTER_SERVER_PORT     int    = 2000           // 主节点端口
	MAX_THREAD_NUM         int    = 20             // 线程最大并发量
	MEDIAN_PAUSE_TIME      int64  = 300            // 间隔时间中位数（单位：ms)【实际间隔时间为 MEDIAN_PAUSE_TIME / 2 ~ MEDIAN_PAUSE_TIME * 2】
	DOCKER_CAP             int    = 10000          // 分段存储容器容量
	OUTPUT_TYPE            string = "csv"          // 文件保存格式
	REQUEST_LIMIT          int64  = 0              // 采集请求上限，0表示不做限制
	UPDATE_PROXY_INTERVALE int64  = 0              // 代理IP更换的间隔时间（单位：分钟）,0 表示不切换IP
	IS_INHERIT_SUCCESS     bool   = true           // 是否继承历史成功纪录
	IS_INHERIT_FAILTURE    bool   = true           // 是否继承历史失败纪录
)

var GlobalSetting = func() Configer {
	Register("ini", &IniConfig{})

	os.MkdirAll(filepath.Clean(HISTORY_DIR_PATH), 0777)
	os.MkdirAll(filepath.Clean(CACHE_DIR_PATH), 0777)
	os.MkdirAll(filepath.Clean(PHANTOM_JS_CACHE_PATH), 0777)

	iniConfiger, err := CreateConfig("ini", CONFIG_FILE_PATH)
	if err != nil {
		file, err := os.Create(CONFIG_FILE_PATH)
		file.Close()
		iniConfiger, err = CreateConfig("ini", CONFIG_FILE_PATH)
		if err != nil {
			panic(err)
		}
		initDefaultConfig(iniConfiger)
	} else {
		checkConfig(iniConfiger)
	}
	iniConfiger.SaveConfigFile(CONFIG_FILE_PATH)

	os.MkdirAll(filepath.Clean(iniConfiger.String("spiderdir")), 0777)
	os.MkdirAll(filepath.Clean(iniConfiger.String("fileoutdir")), 0777)
	os.MkdirAll(filepath.Clean(iniConfiger.String("textoutdir")), 0777)

	return iniConfiger
}()

// 入口
func init() {
	FKStatus.GlobalRuntimeTaskConfig = &FKStatus.AppRuntimeConfig{
		Mode:                 GlobalSetting.DefaultInt("动态配置::本节点角色", APP_MODE),
		MasterPort:           GlobalSetting.DefaultInt("动态配置::主节点端口", MASTER_SERVER_PORT),
		MasterIP:             GlobalSetting.String("动态配置::主节点IP"),
		MaxThreadNum:         GlobalSetting.DefaultInt("动态配置::线程最大并发量", MAX_THREAD_NUM),
		MedianPauseTime:      GlobalSetting.DefaultInt64("动态配置::间隔时间", MEDIAN_PAUSE_TIME),
		OutputType:           GlobalSetting.String("动态配置::文件保存格式"),
		DockerCap:            GlobalSetting.DefaultInt("动态配置::分段存储器容量", DOCKER_CAP),
		RequestLimit:         GlobalSetting.DefaultInt64("动态配置::请求采集上限", REQUEST_LIMIT),
		UpdateProxyIntervale: GlobalSetting.DefaultInt64("动态配置::更变IP代理时间", UPDATE_PROXY_INTERVALE),
		IsInheritSuccess:     GlobalSetting.DefaultBool("动态配置::是否继承历史成功纪录", IS_INHERIT_SUCCESS),
		IsInheritFailure:     GlobalSetting.DefaultBool("动态配置::是否继承历史失败纪录", IS_INHERIT_FAILTURE),
	}
}

// 初始化默认配置
func initDefaultConfig(iniConfiger Configer) {
	iniConfiger.Set("全局::蜘蛛池最大容量", strconv.Itoa(CRAWL_CAP))
	iniConfiger.Set("全局::PhantomJS文件路径", PHANTOM_JS_PATH)
	iniConfiger.Set("全局::代理IP文件路径", PROXY_LIB_FILE_PATH)
	iniConfiger.Set("全局::动态规则目录路径", SPIDER_DIR_PATH)
	iniConfiger.Set("全局::HTML图片输出目录路径", FILE_OUT_DIR_PATH)
	iniConfiger.Set("全局::文本文件输出目录路径", TEXT_OUT_DIR_PATH)
	iniConfiger.Set("全局::数据库名", DATABASE_NAME)
	iniConfiger.Set("日志::缓存最大容量", strconv.FormatInt(LOG_CAP, 10))
	iniConfiger.Set("日志::打印日志输出级别", LOG_PRINT_LEVEL)
	iniConfiger.Set("日志::控制台日志输出级别", LOG_CONSOLE_LEVEL)
	iniConfiger.Set("日志::反馈给服务器的日志级别", LOG_TO_BACKEND_LEVEL)
	iniConfiger.Set("日志::是否打印行信息", fmt.Sprint(IS_LOG_LINE_INFO))
	iniConfiger.Set("日志::是否保存本地日志", fmt.Sprint(IS_LOG_SAVE_TO_LOCAL))
	iniConfiger.Set("MYSQL::连接字符串", MYSQL_CONNECT_STRING)
	iniConfiger.Set("MYSQL::连接池容量", strconv.Itoa(MYSQL_CONNECT_POOL_CAP))
	iniConfiger.Set("MYSQL::通讯缓冲区最大长度", strconv.Itoa(MYSQL_MAX_ALLOWED_PACKET))
	iniConfiger.Set("KAFKA::brokers字符串", KAFKA_BROKERS_STRING)
	iniConfiger.Set("动态配置::本节点角色", strconv.Itoa(APP_MODE))
	iniConfiger.Set("动态配置::主节点IP", MASTER_SERVER_IP)
	iniConfiger.Set("动态配置::主节点端口", strconv.Itoa(MASTER_SERVER_PORT))
	iniConfiger.Set("动态配置::线程最大并发量", strconv.Itoa(MAX_THREAD_NUM))
	iniConfiger.Set("动态配置::间隔时间", strconv.FormatInt(MEDIAN_PAUSE_TIME, 10))
	iniConfiger.Set("动态配置::文件保存格式", OUTPUT_TYPE)
	iniConfiger.Set("动态配置::分段存储器容量", strconv.Itoa(DOCKER_CAP))
	iniConfiger.Set("动态配置::请求采集上限", strconv.FormatInt(REQUEST_LIMIT, 10))
	iniConfiger.Set("动态配置::更变IP代理时间", strconv.FormatInt(UPDATE_PROXY_INTERVALE, 10))
	iniConfiger.Set("动态配置::是否继承历史成功纪录", fmt.Sprint(IS_INHERIT_SUCCESS))
	iniConfiger.Set("动态配置::是否继承历史失败纪录", fmt.Sprint(IS_INHERIT_FAILTURE))
}

// 调整，纠正配置
func checkConfig(iniConfiger Configer) {
	if v, e := iniConfiger.Int("全局::蜘蛛池最大容量"); v <= 0 || e != nil {
		iniConfiger.Set("全局::蜘蛛池最大容量", strconv.Itoa(CRAWL_CAP))
	}
	if v, e := iniConfiger.Int64("日志::缓存最大容量"); v <= 0 || e != nil {
		iniConfiger.Set("日志::缓存最大容量", strconv.FormatInt(LOG_CAP, 10))
	}

	printLevel := iniConfiger.String("日志::打印日志输出级别")
	if FKBase.LogLevelStringToInt(printLevel) == FKBase.LevelUnknown {
		printLevel = LOG_PRINT_LEVEL
	}
	iniConfiger.Set("日志::打印日志输出级别", printLevel)
	consoleLevel := iniConfiger.String("日志::控制台日志输出级别")
	if FKBase.LogLevelStringToInt(consoleLevel) == FKBase.LevelUnknown {
		consoleLevel = LOG_CONSOLE_LEVEL
	}
	iniConfiger.Set("日志::控制台日志输出级别", getLowerLevel(consoleLevel, printLevel))
	backendLevel := iniConfiger.String("日志::反馈给服务器的日志级别")
	if FKBase.LogLevelStringToInt(backendLevel) == FKBase.LevelUnknown {
		backendLevel = LOG_TO_BACKEND_LEVEL
	}
	iniConfiger.Set("日志::反馈给服务器的日志级别", getLowerLevel(backendLevel, printLevel))

	if _, e := iniConfiger.Bool("日志::是否打印行信息"); e != nil {
		iniConfiger.Set("日志::是否打印行信息", fmt.Sprint(IS_LOG_LINE_INFO))
	}
	if _, e := iniConfiger.Bool("日志::是否保存本地日志"); e != nil {
		iniConfiger.Set("日志::是否保存本地日志", fmt.Sprint(IS_LOG_SAVE_TO_LOCAL))
	}
	if v := iniConfiger.String("全局::PhantomJS文件路径"); v == "" {
		iniConfiger.Set("全局::PhantomJS文件路径", PHANTOM_JS_PATH)
	}
	if v := iniConfiger.String("全局::代理IP文件路径"); v == "" {
		iniConfiger.Set("全局::代理IP文件路径", PROXY_LIB_FILE_PATH)
	}
	if v := iniConfiger.String("全局::动态规则目录路径"); v == "" {
		iniConfiger.Set("全局::动态规则目录路径", SPIDER_DIR_PATH)
	}
	if v := iniConfiger.String("全局::HTML图片输出目录路径"); v == "" {
		iniConfiger.Set("全局::HTML图片输出目录路径", FILE_OUT_DIR_PATH)
	}
	if v := iniConfiger.String("全局::文本文件输出目录路径"); v == "" {
		iniConfiger.Set("全局::文本文件输出目录路径", TEXT_OUT_DIR_PATH)
	}
	if v := iniConfiger.String("全局::数据库名"); v == "" {
		iniConfiger.Set("全局::数据库名", DATABASE_NAME)
	}
	if v := iniConfiger.String("MYSQL::连接字符串"); v == "" {
		iniConfiger.Set("MYSQL::连接字符串", MYSQL_CONNECT_STRING)
	}
	if v, e := iniConfiger.Int("MYSQL::连接池容量"); v <= 0 || e != nil {
		iniConfiger.Set("MYSQL::连接池容量", strconv.Itoa(MYSQL_CONNECT_POOL_CAP))
	}
	if v, e := iniConfiger.Int("MYSQL::通讯缓冲区最大长度"); v <= 0 || e != nil {
		iniConfiger.Set("MYSQL::通讯缓冲区最大长度", strconv.Itoa(MYSQL_MAX_ALLOWED_PACKET))
	}
	if v := iniConfiger.String("KAFKA::brokers字符串"); v == "" {
		iniConfiger.Set("KAFKA::brokers字符串", KAFKA_BROKERS_STRING)
	}
	if v, e := iniConfiger.Int("动态配置::本节点角色"); v < FKStatus.UNSET || v > FKStatus.CLIENT || e != nil {
		iniConfiger.Set("动态配置::本节点角色", strconv.Itoa(APP_MODE))
	}
	if v := iniConfiger.String("动态配置::主节点IP"); v == "" {
		iniConfiger.Set("动态配置::主节点IP", MASTER_SERVER_IP)
	}
	if v, e := iniConfiger.Int("动态配置::主节点端口"); v <= 0 || e != nil {
		iniConfiger.Set("动态配置::主节点端口", strconv.Itoa(MASTER_SERVER_PORT))
	}
	if v, e := iniConfiger.Int("动态配置::线程最大并发量"); v <= 0 || e != nil {
		iniConfiger.Set("动态配置::线程最大并发量", strconv.Itoa(MAX_THREAD_NUM))
	}
	if v, e := iniConfiger.Int64("动态配置::间隔时间"); v <= 0 || e != nil {
		iniConfiger.Set("动态配置::间隔时间", strconv.FormatInt(MEDIAN_PAUSE_TIME, 10))
	}
	if v := iniConfiger.String("动态配置::文件保存格式"); v == "" {
		iniConfiger.Set("动态配置::文件保存格式", OUTPUT_TYPE)
	}
	if v, e := iniConfiger.Int("动态配置::分段存储器容量"); v <= 0 || e != nil {
		iniConfiger.Set("动态配置::分段存储器容量", strconv.Itoa(DOCKER_CAP))
	}
	if v, e := iniConfiger.Int64("动态配置::请求采集上限"); v <= 0 || e != nil {
		iniConfiger.Set("动态配置::请求采集上限", strconv.FormatInt(REQUEST_LIMIT, 10))
	}
	if v, e := iniConfiger.Int64("动态配置::更变IP代理时间"); v <= 0 || e != nil {
		iniConfiger.Set("动态配置::更变IP代理时间", strconv.FormatInt(UPDATE_PROXY_INTERVALE, 10))
	}
	if _, e := iniConfiger.Bool("动态配置::是否继承历史成功纪录"); e != nil {
		iniConfiger.Set("动态配置::是否继承历史成功纪录", fmt.Sprint(IS_INHERIT_SUCCESS))
	}
	if _, e := iniConfiger.Bool("动态配置::是否继承历史失败纪录"); e != nil {
		iniConfiger.Set("动态配置::是否继承历史失败纪录", fmt.Sprint(IS_INHERIT_FAILTURE))
	}

	iniConfiger.SaveConfigFile(CONFIG_FILE_PATH)
}

func getLowerLevel(l string, g string) string {
	a, b := FKBase.LogLevelStringToInt(l), FKBase.LogLevelStringToInt(g)
	if a < b {
		return l
	}
	return g
}
