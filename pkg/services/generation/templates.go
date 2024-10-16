package generation

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Jacobbrewer1/goschema/pkg/models"
	"github.com/Masterminds/sprig"
	"github.com/huandu/xstrings"
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

func generate(t *templateInfo, tmpl *template.Template, outputLoc string, fileExtensionPrefix string) error {
	ext := ".go"
	if fileExtensionPrefix != "" {
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
