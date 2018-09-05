package models

import (
	"fmt"

	"git.hoogi.eu/go-blog/components/mail"
	"git.hoogi.eu/go-blog/settings"
	"git.hoogi.eu/go-blog/utils"
)

type Mailer struct {
	AppConfig   *settings.Application
	MailService *mail.Service
}

func (m Mailer) SendPasswordChangeConfirmation(u *User) error {
	mail := mail.Mail{
		To:      u.Email,
		Subject: "Password change",
		Body:    fmt.Sprintf("Hi %s, \n\n your password change was sucessfully.", u.DisplayName),
	}

	return m.MailService.Send(mail)
}

func (m Mailer) SendPasswordResetLink(u *User, t *Token) error {
	resetLink := utils.AppendString(m.AppConfig.Domain, "/admin/reset-password/", t.Hash)

	mail := mail.Mail{
		To:      u.Email,
		Subject: "Changing password instructions",
		Body:    fmt.Sprintf("Hi %s, \n\n use the following link to reset your password: \n\n. %s", u.DisplayName, resetLink),
	}

	return m.MailService.Send(mail)
}
