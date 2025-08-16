package email

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/mail.v2"
)

// EmailService handles email operations
type EmailService struct {
	dialer *mail.Dialer
	from   string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	// Get SMTP configuration from environment variables
	smtpHost := getEnvOrDefault("SMTP_HOST", "smtp.gmail.com")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := getEnvOrDefault("FROM_EMAIL", smtpUser)

	// Log email service configuration
	if smtpUser != "" && smtpPass != "" {
		log.Printf("📧 Email service configured - Host: %s, User: %s", smtpHost, smtpUser)
	} else {
		log.Printf("⚠️  Email service not configured - Set SMTP_USER and SMTP_PASS environment variables")
	}

	// Create dialer
	dialer := mail.NewDialer(smtpHost, 587, smtpUser, smtpPass)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	return &EmailService{
		dialer: dialer,
		from:   fromEmail,
	}
}

// SendPasswordResetEmail sends a password reset email
func (es *EmailService) SendPasswordResetEmail(to, resetLink string) error {
	log.Printf("📧 Attempting to send password reset email to: %s", to)

	subject := "Password Reset Request - Stock Panel"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>You have requested to reset your password for your Stock Panel account.</p>
			<p>Click the link below to reset your password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>This link will expire in 1 hour.</p>
			<p>If you didn't request this password reset, please ignore this email.</p>
			<br>
			<p>Best regards,<br>Stock Panel Team</p>
		</body>
		</html>
	`, resetLink)

	log.Printf("📧 Email details - From: %s, To: %s, Subject: %s", es.from, to, subject)

	return es.SendEmail(to, subject, body)
}

// SendEmail sends an email using the configured SMTP settings
func (es *EmailService) SendEmail(to, subject, body string) error {
	log.Printf("📧 Connecting to SMTP server...")
	log.Printf("📧 SMTP Details - Host: %s, Port: 587, User: %s", es.dialer.Host, es.dialer.Username)

	m := mail.NewMessage()
	m.SetHeader("From", es.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	log.Printf("📧 Attempting to send email - From: %s, To: %s, Subject: %s", es.from, to, subject)

	// Try to send the email
	if err := es.dialer.DialAndSend(m); err != nil {
		log.Printf("❌ Failed to send email to %s: %v", to, err)
		log.Printf("❌ SMTP Error Details: %T - %v", err, err)
		return err
	}

	log.Printf("✅ Email sent successfully to %s", to)
	return nil
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsEmailConfigured checks if email service is properly configured
func (es *EmailService) IsEmailConfigured() bool {
	return os.Getenv("SMTP_USER") != "" && os.Getenv("SMTP_PASS") != ""
}
