package FKLog

import (
	"FKBase"
	"encoding/json"
	"io"
	"log"
	"net"
)

type TcpWriter struct {
	lg             *log.Logger
	innerWriter    io.WriteCloser
	ReconnectOnMsg bool   `json:"reconnectOnMsg"` // 是否发送一条日志就断开连接（是否短连接）
	Net            string `json:"net"`            // 连接远程的网络类型，例如"tcp", "udp", "ip4:1", "ip6:ipv6-icmp"
	Addr           string `json:"addr"`           // 连接远程的地址，例如"golang.org:http"， "192.0.2.1:http"， "198.51.100.1:80"， ":80"， "192.0.2.1"
	Level          int    `json:"level"`          // 连接远程的日志级别
}

func createTCPLogger() ILogger {
	conn := new(TcpWriter)
	conn.Level = FKBase.LevelDebug
	return conn
}

// Init connect logger with json config.
// config like:
//	{
//	"reconnectOnMsg":false,
//	"net":"udp",
//	"addr":"127.0.0.1:8080",
//	"level":3
//	}
func (c *TcpWriter) Init(config map[string]interface{}) error {
	conf, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return json.Unmarshal(conf, c)
}

func (c *TcpWriter) WriteMsg(msg string, level int) error {
	if level > c.Level {
		return nil
	}
	if c.neededConnectOnMsg() {
		err := c.connect()
		if err != nil {
			return err
		}
	}

	if c.ReconnectOnMsg {
		defer c.innerWriter.Close()
	}
	c.lg.Println(msg)
	return nil
}

// destroy connection writer and close tcp listener.
func (c *TcpWriter) Destroy() {
	if c.innerWriter != nil {
		c.innerWriter.Close()
	}
}

func (c *TcpWriter) Flush() {

}

func (c *TcpWriter) connect() error {
	if c.innerWriter != nil {
		c.innerWriter.Close()
		c.innerWriter = nil
	}

	conn, err := net.Dial(c.Net, c.Addr)
	if err != nil {
		return err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}

	c.innerWriter = conn
	c.lg = log.New(conn, "", log.Ldate|log.Ltime)
	return nil
}

func (c *TcpWriter) neededConnectOnMsg() bool {
	if c.innerWriter == nil {
		return true
	}

	return c.ReconnectOnMsg
}
