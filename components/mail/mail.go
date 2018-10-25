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

func NewMailService(subjectPrefix, from string, smtpConfig SMTPConfig) Service {
	s := Service{
		SubjectPrefix: subjectPrefix,
		From:          from,
		SMTPConfig:    smtpConfig,
	}

	go s.readBuffer()

	return s

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

var buffer = make(chan Mail, 10)
var errc = make(chan error, 1)

func (s Service) readBuffer() <-chan error {
	for {
		mail := <-buffer
		if err := s.Send(mail); err != nil {
			errc <- err
		}

		close(errc)
	}

	return errc
}

func (s Service) SendAsync(m Mail) error {
	go func() {
		buffer <- m
	}()

	return <-errc
}
