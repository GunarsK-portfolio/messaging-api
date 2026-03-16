package renderer

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/GunarsK-portfolio/messaging-api/internal/templates"
	"github.com/GunarsK-portfolio/portfolio-common/models"
)

// emailTypeMeta holds subject line and required template keys per email type
type emailTypeMeta struct {
	Subject      string
	RequiredKeys []string
}

var typeRegistry = map[string]emailTypeMeta{
	models.EmailTypeEmailVerification: {
		Subject:      "Verify your email address",
		RequiredKeys: []string{"username", "verify_url"},
	},
}

// parsed templates, loaded once at init
var parsedTemplates = map[string]*template.Template{}

func init() {
	for emailType := range typeRegistry {
		filename := emailType + ".html"
		tmpl, err := template.ParseFS(templates.FS, filename)
		if err != nil {
			panic(fmt.Sprintf("renderer: failed to parse template %s: %v", filename, err))
		}
		parsedTemplates[emailType] = tmpl
	}
}

// SubjectForType returns the subject line for an email type
func SubjectForType(emailType string) string {
	if meta, ok := typeRegistry[emailType]; ok {
		return meta.Subject
	}
	return ""
}

// Render renders an email template with the given data.
// Returns an error if the email type is unknown or required keys are missing.
func Render(emailType string, data map[string]string) (string, error) {
	meta, ok := typeRegistry[emailType]
	if !ok {
		return "", fmt.Errorf("unknown email type: %s", emailType)
	}

	for _, key := range meta.RequiredKeys {
		if data[key] == "" {
			return "", fmt.Errorf("missing required template key %q for type %s", key, emailType)
		}
	}

	tmpl := parsedTemplates[emailType]

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", emailType, err)
	}

	return buf.String(), nil
}
