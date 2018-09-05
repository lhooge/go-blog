package mail

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
)

//Service holds configuration for the SMTP server
//The sender address and an optional subject prefix
type Service struct {
	SubjectPrefix string
	SMTPConfig    SMTPConfig
	From          string
}

//SMTPConfig holds the configuration for the SMTP server
type SMTPConfig struct {
	Address  string
	Port     int
	User     string
	Password []byte
}

//Mail represents a mail
type Mail struct {
	To      string
	Subject string
	Body    string
}

func (m Mail) buildMessage(s Service) []byte {
	var buf bytes.Buffer
	buf.WriteString("To: ")
	buf.WriteString(m.To)
	buf.WriteString("\r\n")
	buf.WriteString("Subject: ")
	buf.WriteString(s.SubjectPrefix)
	buf.WriteString(m.Subject)
	buf.WriteString("\r\n")
	buf.WriteString(m.Body)

	return buf.Bytes()
}

func (m Mail) validate() error {
	if len(m.To) == 0 {
		return errors.New("no recipient specified")
	}

	return nil
}

//Send sends a mail over the configured SMTP server
func (s Service) Send(m Mail) error {
	auth := smtp.PlainAuth("", s.SMTPConfig.User, string(s.SMTPConfig.Password), s.SMTPConfig.Address)

	return smtp.SendMail(fmt.Sprintf("%s:%d", s.SMTPConfig.Address, s.SMTPConfig.Port), auth, s.From, []string{m.To}, m.buildMessage(s))
}
