package FKLog

import (
	"FKBase"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SmtpWriter struct {
	Username           string   `json:"Username"`
	Password           string   `json:"password"`
	Host               string   `json:"Host"`
	Subject            string   `json:"subject"`
	FromAddress        string   `json:"fromAddress"`
	RecipientAddresses []string `json:"sendTos"`
	Level              int      `json:"level"`
}

func createMailLogger() ILogger {
	return &SmtpWriter{
		Level: FKBase.LevelDebug,
	}
}

// init smtp logger with json config.
// config like:
//	{
//		"Username":"example@gmail.com",
//		"password:"password",
//		"host":"smtp.gmail.com:465",
//		"subject":"email title",
//		"fromAddress":"from@example.com",
//		"sendTos":["email1","email2"],
//		"level":LevelError
//	}
func (s *SmtpWriter) Init(config map[string]interface{}) error {
	conf, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return json.Unmarshal(conf, s)
}

func (s *SmtpWriter) WriteMsg(msg string, level int) error {
	if level > s.Level {
		return nil
	}

	hp := strings.Split(s.Host, ":")
	auth := s.getSmtpAuth(hp[0])
	content_type := "Content-Type: text/plain" + "; charset=UTF-8"
	mailmsg := []byte("To: " + strings.Join(s.RecipientAddresses, ";") + "\r\nFrom: " + s.FromAddress + "<" + s.FromAddress +
		">\r\nSubject: " + s.Subject + "\r\n" + content_type + "\r\n\r\n" + fmt.Sprintf(".%s", time.Now().Format("2006-01-02 15:04:05")) + msg)

	return s.sendMail(s.Host, auth, s.FromAddress, s.RecipientAddresses, mailmsg)
}

func (s *SmtpWriter) Flush() {
}

func (s *SmtpWriter) Destroy() {
}

func (s *SmtpWriter) getSmtpAuth(host string) smtp.Auth {
	if len(strings.Trim(s.Username, " ")) == 0 && len(strings.Trim(s.Password, " ")) == 0 {
		return nil
	}
	return smtp.PlainAuth(
		"",
		s.Username,
		s.Password,
		host,
	)
}

func (s *SmtpWriter) sendMail(hostAddressWithPort string, auth smtp.Auth, fromAddress string, recipients []string, msgContent []byte) error {
	client, err := smtp.Dial(hostAddressWithPort)
	if err != nil {
		return err
	}

	host, _, _ := net.SplitHostPort(hostAddressWithPort)
	tlsConn := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	if err = client.StartTLS(tlsConn); err != nil {
		return err
	}

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(fromAddress); err != nil {
		return err
	}

	for _, rec := range recipients {
		if err = client.Rcpt(rec); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msgContent))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = client.Quit()
	if err != nil {
		return err
	}

	return nil
}
