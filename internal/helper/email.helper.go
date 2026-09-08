package helper

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	emailLogger "studio-book-be-go/internal/repository/email_logger"
	"time"

	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

var templatesFS embed.FS

type EmailResult struct {
	MessageID string
	Rejected  []string
	Accepted  []string
	Raw       string
}

type EmailHelper struct {
	emailLoggerRepo emailLogger.EmailLoggerRepository
	viewPath        string
	db              *gorm.DB
}

func NewSendEmailHelper(emailLoggerRepo emailLogger.EmailLoggerRepository, db *gorm.DB) *EmailHelper {
	return &EmailHelper{
		db:              db,
		emailLoggerRepo: emailLoggerRepo,
		viewPath:        "./views",
	}
}

func (h *EmailHelper) SendEmail(
	ctx context.Context,
	email string,
	subject string,
	templateName string,
	data map[string]interface{},
	userID uint64,
) (*EmailResult, error) {

	body, err := h.renderTemplate(templateName, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	port, err := strconv.Atoi(os.Getenv("SERVICE_EMAIL_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVICE_EMAIL_PORT: %w", err)
	}

	dialer := gomail.NewDialer(
		os.Getenv("SERVICE_EMAIL"),
		port,
		os.Getenv("SENDER_EMAIL"),
		os.Getenv("PASS_EMAIL"),
	)

	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", os.Getenv("SENDER_EMAIL_NAME"), os.Getenv("SENDER_EMAIL")))
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := dialer.DialAndSend(m); err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	result := &EmailResult{
		Accepted: []string{email},
		Rejected: []string{},
	}

	metadataBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email result: %w", err)
	}
	rawMetadata := json.RawMessage(metadataBytes)

	payloadData := emailLogger.EmailLogger{
		UserID:    userID,
		Type:      templateName,
		Metadata:  &rawMetadata,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		Email:     email,
	}

	_, err = h.emailLoggerRepo.Save(payloadData, h.db.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to log email: %w", err)
	}


	return result, nil
}

func (h *EmailHelper) renderTemplate(templateName string, data map[string]interface{}) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Buat path absolut dari root proyek
	tmplPath := filepath.Join(wd, "views", templateName+".hbs")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
