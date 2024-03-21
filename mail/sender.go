package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// Constants for SMTP server configuration
const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

// EmailSender interface defines the method for sending emails, for mocking also
type EmailSender interface {
	SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error
}

// GmailSender struct implements the EmailSender interface for sending emails via Gmail
type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

// NewGmailSender creates a new GmailSender instance
func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

// SendEmail sends an email using GmailSender
func (sender *GmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	// Create a new email instance
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	// Attach files to the email
	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	// Configure SMTP authentication
	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, smtpAuthAddress)

	// Send the email
	return e.Send(smtpServerAddress, smtpAuth)
}


/*

Constants:

smtpAuthAddress and smtpServerAddress define the authentication and server addresses for the Gmail SMTP server.
EmailSender Interface (EmailSender):

Declares the SendEmail method, which is implemented by types responsible for sending emails.
GmailSender Struct (GmailSender):

Represents an email sender for Gmail.
Holds information such as the sender's name, email address, and password.
NewGmailSender Function:

Creates and returns a new instance of GmailSender with the provided parameters.
SendEmail Method (GmailSender):

Implements the SendEmail method of the EmailSender interface.
Uses the email package to construct and send an email with HTML content, attachments, and recipients.
Configures SMTP authentication and sends the email via the Gmail SMTP server.
This code defines an interface EmailSender and a concrete implementation GmailSender for sending emails. The SendEmail method is responsible for composing and sending emails using the Gmail SMTP server.
 The code structure promotes separation of concerns and flexibility for using different email sending implementations.

*/