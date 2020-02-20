package models

import (
	"fmt"

	"git.hoogi.eu/snafu/go-blog/components/mail"
	"git.hoogi.eu/snafu/go-blog/settings"
	"git.hoogi.eu/snafu/go-blog/utils"
)

type Mailer struct {
	AppConfig *settings.Application
	Sender    mail.Sender
}

func (m Mailer) SendActivationLink(ui *UserInvite) {
	activation := utils.AppendString(m.AppConfig.Domain, "/admin/activate-account/", ui.Hash)

	mail := mail.Mail{
		To:      ui.Email,
		Subject: "You got an invitation",
		Body:    fmt.Sprintf("Hi %s,\n\n you are invited join %s. To activate your account click the following link and enter a password %s", ui.DisplayName, m.AppConfig.Title, activation),
	}

	m.Sender.SendAsync(mail)
}

func (m Mailer) SendPasswordChangeConfirmation(u *User) {
	mail := mail.Mail{
		To:      u.Email,
		Subject: "Password change",
		Body:    fmt.Sprintf("Hi %s,\n\nyour password change was successful.", u.DisplayName),
	}

	m.Sender.SendAsync(mail)
}

func (m Mailer) SendPasswordResetLink(u *User, t *Token) {
	resetLink := utils.AppendString(m.AppConfig.Domain, "/admin/reset-password/", t.Hash)

	mail := mail.Mail{
		To:      u.Email,
		Subject: "Changing password instructions",
		Body:    fmt.Sprintf("Hi %s,\n\nuse the following link to reset your password:\n\n%s", u.DisplayName, resetLink),
	}

	m.Sender.SendAsync(mail)
}
