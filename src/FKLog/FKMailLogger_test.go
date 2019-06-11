package FKLog

import (
	"testing"
	"time"
)

func TestSmtp(t *testing.T) {
	log := CreateDefaultLogger(10000)
	log.SetLogger("smtp", map[string]interface{}{
		"username": "xxxx@gmail.com",
		"password": "DOYOUWANTTOKNOW?",
		"host":     "smtp.gmail.com:587",
		"sendTos": []string{
			"xxxx@gmail.com",
		},
	})
	log.Informational("sendmail infos")
	time.Sleep(time.Second * 30)
}
