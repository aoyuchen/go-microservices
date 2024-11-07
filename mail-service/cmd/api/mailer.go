// generate and send an email

package main

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

// For the mail server we will be communicating with
type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

// For the content of each individual email that we will send
type Message struct {
	From        string // email address ? Is From the same as FromAddress in Mail struct?
	FromName    string // name associated with that email address
	To          string // recepient
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

// Functions that will allow us to send email
func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress // sender address
	}

	if msg.FromName == "" {
		msg.From = m.FromName // sender name
	}

	// Create two template, one for html and one for plain text
	// then pass data to those templates
	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data

	// html version of the message
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	// plain text version of the message
	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	// create the mail server
	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption) // to support as many as mail services possible, make life easiser when swiching mail server
	server.KeepAlive = false                          // ?? what does this mean?
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// connect to the mail server
	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	// create an email message using the client
	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)

	// add attachment
	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	// send the email
	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := "./template/mail.plain.gohtml" // the template we want to render

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()
	return plainMessage, nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := "./template/mail.html.gohtml" // the template we want to render

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}
	return formattedMessage, nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *Mail) getEncryption(s string) mail.Encryption {
	switch s {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
