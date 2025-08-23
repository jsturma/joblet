package installers

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates
var templateFiles embed.FS

// BaseInstaller provides common functionality for all installers
type BaseInstaller struct {
	templateName string
}

// NewBaseInstaller creates a new base installer with the specified template
func NewBaseInstaller(templateName string) *BaseInstaller {
	return &BaseInstaller{
		templateName: templateName,
	}
}

// RenderTemplate renders the template with the provided data
func (b *BaseInstaller) RenderTemplate(data *TemplateData) (string, error) {
	templatePath := fmt.Sprintf("templates/%s.tmpl", b.templateName)

	templateContent, err := templateFiles.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", b.templateName, err)
	}

	tmpl, err := template.New(b.templateName).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", b.templateName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", b.templateName, err)
	}

	return buf.String(), nil
}

// CreateInstallResult creates a standard install result with command and args
func (b *BaseInstaller) CreateInstallResult(script string, success bool, message string) *InstallResult {
	// For now, use inline execution until we can properly handle temp files in chroot
	return &InstallResult{
		Success: success,
		Command: "bash",
		Args:    []string{"-c", script},
		Message: message,
	}
}
