package renderer

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/GunarsK-portfolio/messaging-api/internal/templates"
	"github.com/GunarsK-portfolio/portfolio-common/models"
)

// subjectByType maps email types to subject lines
var subjectByType = map[string]string{
	models.EmailTypeEmailVerification: "Verify your email address",
	models.EmailTypePasswordReset:     "Reset your password",
	models.EmailType2FACode:           "Your verification code",
}

// SubjectForType returns the subject line for an email type
func SubjectForType(emailType string) string {
	if s, ok := subjectByType[emailType]; ok {
		return s
	}
	return ""
}

// Render renders an email template with the given data
func Render(emailType string, data map[string]string) (string, error) {
	filename := emailType + ".html"

	tmpl, err := template.ParseFS(templates.FS, filename)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", filename, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", filename, err)
	}

	return buf.String(), nil
}
