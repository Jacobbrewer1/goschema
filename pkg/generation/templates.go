package generation

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/huandu/xstrings"
	"github.com/jacobbrewer1/goschema/pkg/models"
)

type templateInfo struct {
	OutputDir string
	Table     *models.Table
}

func RenderTemplates(tables []*models.Table, templatesLoc, outputLoc string, fileExtensionPrefix string) error {
	tmpl, err := template.New("model.tmpl").Funcs(sprig.TxtFuncMap()).Funcs(Helpers).ParseGlob(templatesLoc)
	if err != nil {
		return fmt.Errorf("error parsing templates: %w", err)
	}

	for _, t := range tables {
		if err = generate(&templateInfo{
			OutputDir: outputLoc,
			Table:     t,
		}, tmpl, outputLoc, fileExtensionPrefix); err != nil {
			return fmt.Errorf("error generating template: %w", err)
		}
	}

	return nil
}

// RenderWithTemplates renders templates that are provided as embedded files
func RenderWithTemplates(fs embed.FS, tables []*models.Table, outputLoc string, fileExtensionPrefix string) error {
	tmpl, err := template.New("model.tmpl").Funcs(sprig.TxtFuncMap()).Funcs(Helpers).ParseFS(fs, "templates/*.tmpl")
	if err != nil {
		return fmt.Errorf("error parsing templates: %w", err)
	}

	for _, t := range tables {
		if err := generate(&templateInfo{
			OutputDir: outputLoc,
			Table:     t,
		}, tmpl, outputLoc, fileExtensionPrefix); err != nil {
			return fmt.Errorf("error generating template: %w", err)
		}
	}

	return nil
}

func generate(t *templateInfo, tmpl *template.Template, outputLoc string, fileExtensionPrefix string) error {
	ext := ".go"
	if fileExtensionPrefix != "" {
		// Add a period if it's not already there
		if fileExtensionPrefix[0] != '.' {
			fileExtensionPrefix = "." + fileExtensionPrefix
		}
		ext = fileExtensionPrefix + ext
	}

	fn := filepath.Join(outputLoc, xstrings.ToSnakeCase(t.Table.Name)+ext)
	if err := os.MkdirAll(filepath.Dir(fn), 0750); err != nil {
		return err
	}

	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			slog.Warn("Error closing file", slog.String("error", err.Error()))
		}
	}(f)

	return tmpl.Execute(f, t)
}

func GoimportsInstallIfNeeded() error {
	if !IsGoimportsInstalled() {
		slog.Info("Installing goimports")
		if err := InstallGoimports(); err != nil {
			return fmt.Errorf("error installing goimports: %w", err)
		}
	}

	return nil
}

func IsGoimportsInstalled() bool {
	_, err := exec.LookPath("goimports")
	return err == nil
}

func InstallGoimports() error {
	cmd := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}