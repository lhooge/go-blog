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

type Sender interface {
	Send(m Mail) error
	SendAsync(m Mail) error
}

//SMTPConfig holds the configuration for the SMTP server
type SMTPConfig struct {
	Address  string
	Port     int
	User     string
	Helo     string
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

	buf.WriteString("From: ")
	buf.WriteString(s.From)
	buf.WriteString("\r\n")
	buf.WriteString("To: ")
	buf.WriteString(m.To)
	buf.WriteString("\r\n")
	buf.WriteString("Subject: ")

	if len(s.SubjectPrefix) > 0 {
		buf.WriteString(s.SubjectPrefix)
	}

	buf.WriteString(m.Subject)
	buf.WriteString("\r\n")
	buf.WriteString(m.Body)

	return buf.Bytes()
}

func (s Service) SendAsync(m Mail) error {
	go func() {
		buffer <- m
	}()

	return <-errc
}

//Send sends a mail over the configured SMTP server
func (s Service) Send(m Mail) error {
	if len(s.SMTPConfig.User) > 0 && len(s.SMTPConfig.Password) > 0 {
		auth := smtp.PlainAuth("", s.SMTPConfig.User, string(s.SMTPConfig.Password), s.SMTPConfig.Address)

		err := smtp.SendMail(fmt.Sprintf("%s:%d", s.SMTPConfig.Address, s.SMTPConfig.Port), auth, s.From, []string{m.To}, m.buildMessage(s))

		if err != nil {
			return err
		}

		return nil
	} else {
		//anonymous
		c, err := smtp.Dial(fmt.Sprintf("%s:%d", s.SMTPConfig.Address, s.SMTPConfig.Port))

		if err != nil {
			return err
		}

		if len(s.SMTPConfig.Helo) > 0 {
			if err := c.Hello(s.SMTPConfig.Helo); err != nil {
				return err
			}
		}

		// Set the sender and recipient first
		if err := c.Mail(s.From); err != nil {
			return err
		}
		if err := c.Rcpt(m.To); err != nil {
			return err
		}

		wc, err := c.Data()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(wc, string(m.buildMessage(s)))

		if err != nil {
			return err
		}

		err = wc.Close()

		if err != nil {
			return err
		}

		return c.Quit()
	}
}

func (m Mail) validate() error {
	if len(m.To) == 0 {
		return errors.New("no recipient specified")
	}

	return nil
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

		return errc
	}
}
