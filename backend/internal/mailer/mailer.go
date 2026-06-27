package mailer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log/slog"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

// sanitiseHeader removes CR and LF so that user-derived values (for example a
// certification name in the subject) cannot inject additional SMTP headers
// (CWE-93).
func sanitiseHeader(v string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(v)
}

// Mailer defines the interface for sending transactional emails.
type Mailer interface {
	SendVerificationEmail(to, name, token string) error
	SendEmailChangeConfirmation(to, name, token string) error
	SendPasswordResetEmail(to, name, token string) error
	SendOAuthResetNotification(to, name, provider string) error
	SendCertificationReminder(to, name string, data CertReminderData) error
}

// SMTPMailer sends emails via an SMTP server with STARTTLS.
type SMTPMailer struct {
	host        string
	port        string
	username    string
	password    string
	from        string
	frontendURL string
}

// NewSMTPMailer creates a new SMTPMailer with the given SMTP configuration.
func NewSMTPMailer(host, port, username, password, from, frontendURL string) *SMTPMailer {
	return &SMTPMailer{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		from:        from,
		frontendURL: frontendURL,
	}
}

// SendVerificationEmail sends an email asking the user to verify their email address.
func (m *SMTPMailer) SendVerificationEmail(to, name, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", m.frontendURL, token)

	htmlBody, err := renderTemplate(verificationHTMLTmpl, emailData{
		Name: name,
		Link: link,
	})
	if err != nil {
		return fmt.Errorf("rendering verification email HTML: %w", err)
	}

	textBody, err := renderTemplate(verificationTextTmpl, emailData{
		Name: name,
		Link: link,
	})
	if err != nil {
		return fmt.Errorf("rendering verification email text: %w", err)
	}

	return m.send(to, verificationSubject, htmlBody, textBody)
}

// SendEmailChangeConfirmation sends an email asking the user to confirm an email change.
func (m *SMTPMailer) SendEmailChangeConfirmation(to, name, token string) error {
	link := fmt.Sprintf("%s/confirm-email-change?token=%s", m.frontendURL, token)

	htmlBody, err := renderTemplate(emailChangeHTMLTmpl, emailData{
		Name: name,
		Link: link,
	})
	if err != nil {
		return fmt.Errorf("rendering email change HTML: %w", err)
	}

	textBody, err := renderTemplate(emailChangeTextTmpl, emailData{
		Name: name,
		Link: link,
	})
	if err != nil {
		return fmt.Errorf("rendering email change text: %w", err)
	}

	return m.send(to, emailChangeSubject, htmlBody, textBody)
}

// SendPasswordResetEmail sends an email with a link to reset the user's password.
func (m *SMTPMailer) SendPasswordResetEmail(to, name, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", m.frontendURL, token)

	htmlBody, err := renderTemplate(passwordResetHTMLTmpl, emailData{
		Name: name,
		Link: link,
	})
	if err != nil {
		return fmt.Errorf("rendering password reset email HTML: %w", err)
	}

	textBody, err := renderTemplate(passwordResetTextTmpl, emailData{
		Name: name,
		Link: link,
	})
	if err != nil {
		return fmt.Errorf("rendering password reset email text: %w", err)
	}

	return m.send(to, passwordResetSubject, htmlBody, textBody)
}

// SendOAuthResetNotification sends an email informing the user that they
// signed up via an OAuth provider and should use that provider to sign in.
func (m *SMTPMailer) SendOAuthResetNotification(to, name, provider string) error {
	htmlBody, err := renderTemplate(oauthResetHTMLTmpl, oauthEmailData{
		Name:     name,
		Provider: provider,
	})
	if err != nil {
		return fmt.Errorf("rendering OAuth reset notification HTML: %w", err)
	}

	textBody, err := renderTemplate(oauthResetTextTmpl, oauthEmailData{
		Name:     name,
		Provider: provider,
	})
	if err != nil {
		return fmt.Errorf("rendering OAuth reset notification text: %w", err)
	}

	return m.send(to, oauthResetSubject, htmlBody, textBody)
}

// SendCertificationReminder sends a renewal reminder for a certification that
// is expiring soon or has expired.
func (m *SMTPMailer) SendCertificationReminder(to, name string, data CertReminderData) error {
	td := certReminderTemplateData{
		Name:          name,
		CertName:      data.CertName,
		Provider:      data.Provider,
		ExpiryDate:    data.ExpiryDate,
		DaysRemaining: data.DaysRemaining,
		Expired:       data.Expired,
		Link:          fmt.Sprintf("%s/certifications/%s/edit", m.frontendURL, data.CertID),
		CredentialURL: data.CredentialURL,
	}

	htmlBody, err := renderTemplate(certReminderHTMLTmpl, td)
	if err != nil {
		return fmt.Errorf("rendering certification reminder HTML: %w", err)
	}

	textBody, err := renderTemplate(certReminderTextTmpl, td)
	if err != nil {
		return fmt.Errorf("rendering certification reminder text: %w", err)
	}

	return m.send(to, certReminderSubject(data), htmlBody, textBody)
}

// send builds a multipart/alternative MIME message and delivers it via SMTP with STARTTLS.
func (m *SMTPMailer) send(to, subject, htmlBody, textBody string) error {
	addr := net.JoinHostPort(m.host, m.port)

	// Build MIME multipart/alternative message.
	boundary := "----=_Part_trackmy_career_boundary"

	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s\r\n", sanitiseHeader(m.from))
	fmt.Fprintf(&msg, "To: %s\r\n", sanitiseHeader(to))
	fmt.Fprintf(&msg, "Subject: %s\r\n", sanitiseHeader(subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	msg.WriteString("\r\n")

	// Plain text part.
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	msg.WriteString("\r\n")
	textQP := quotedprintable.NewWriter(&msg)
	textQP.Write([]byte(textBody))
	textQP.Close()
	msg.WriteString("\r\n")

	// HTML part.
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	msg.WriteString("\r\n")
	htmlQP := quotedprintable.NewWriter(&msg)
	htmlQP.Write([]byte(htmlBody))
	htmlQP.Close()
	msg.WriteString("\r\n")

	// Closing boundary.
	fmt.Fprintf(&msg, "--%s--\r\n", boundary)

	// Connect to SMTP server.
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connecting to SMTP server: %w", err)
	}

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("creating SMTP client: %w", err)
	}
	defer func() { _ = client.Close() }()

	// Upgrade to TLS (STARTTLS).
	tlsConfig := &tls.Config{
		ServerName: m.host,
		MinVersion: tls.VersionTLS12,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS: %w", err)
	}

	// Authenticate.
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth: %w", err)
	}

	// Set sender and recipient. The envelope sender must be a bare email
	// address; the display name only belongs in the From header.
	envelopeFrom := m.from
	if addr, parseErr := mail.ParseAddress(m.from); parseErr == nil {
		envelopeFrom = addr.Address
	}
	if err := client.Mail(envelopeFrom); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}

	// Write the message body.
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}
	if _, err := w.Write(msg.Bytes()); err != nil {
		return fmt.Errorf("writing email body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("closing email body writer: %w", err)
	}

	return client.Quit()
}

// NoopMailer logs email details instead of sending them. Useful for development
// when SMTP is not configured.
type NoopMailer struct {
	frontendURL string
}

// NewNoopMailer creates a new NoopMailer that logs emails to slog.
func NewNoopMailer(frontendURL string) *NoopMailer {
	return &NoopMailer{frontendURL: frontendURL}
}

// SendVerificationEmail logs the verification email details.
func (m *NoopMailer) SendVerificationEmail(to, name, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", m.frontendURL, token)
	slog.Info("noop mailer: verification email",
		"to", to,
		"subject", verificationSubject,
		"name", name,
		"link", link,
	)
	return nil
}

// SendPasswordResetEmail logs the password reset email details.
func (m *NoopMailer) SendPasswordResetEmail(to, name, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", m.frontendURL, token)
	slog.Info("noop mailer: password reset email",
		"to", to,
		"subject", passwordResetSubject,
		"name", name,
		"link", link,
	)
	return nil
}

// SendOAuthResetNotification logs the OAuth reset notification details.
func (m *NoopMailer) SendOAuthResetNotification(to, name, provider string) error {
	slog.Info("noop mailer: OAuth reset notification",
		"to", to,
		"subject", oauthResetSubject,
		"name", name,
		"provider", provider,
	)
	return nil
}

// SendEmailChangeConfirmation logs the email change confirmation details.
func (m *NoopMailer) SendEmailChangeConfirmation(to, name, token string) error {
	link := fmt.Sprintf("%s/confirm-email-change?token=%s", m.frontendURL, token)
	slog.Info("noop mailer: email change confirmation",
		"to", to,
		"subject", emailChangeSubject,
		"name", name,
		"link", link,
	)
	return nil
}

// SendCertificationReminder logs the certification reminder details.
func (m *NoopMailer) SendCertificationReminder(to, name string, data CertReminderData) error {
	slog.Info("noop mailer: certification reminder",
		"to", to,
		"subject", certReminderSubject(data),
		"name", name,
		"cert", data.CertName,
		"days_remaining", data.DaysRemaining,
		"expired", data.Expired,
	)
	return nil
}

// New returns an SMTPMailer if host is configured, otherwise a NoopMailer.
func New(host, port, username, password, from, frontendURL string) Mailer {
	if host != "" {
		return NewSMTPMailer(host, port, username, password, from, frontendURL)
	}
	slog.Info("SMTP host not configured, using noop mailer (emails will be logged)")
	return NewNoopMailer(frontendURL)
}
