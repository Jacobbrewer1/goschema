package generation

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/huandu/xstrings"
	"github.com/jacobbrewer1/goschema/pkg/entities"
	"github.com/jacobbrewer1/goschema/pkg/logging"
)

type templateInfo struct {
	OutputDir string
	Table     *entities.Table
}

func RenderTemplates(tables []*entities.Table, templatesLoc, outputLoc, fileExtensionPrefix string) error {
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
func RenderWithTemplates(fs embed.FS, tables []*entities.Table, outputLoc, fileExtensionPrefix string) error {
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

	if err := renderHelpers(fs, outputLoc, fileExtensionPrefix); err != nil {
		return fmt.Errorf("error rendering helpers: %w", err)
	}

	return nil
}

func renderHelpers(fs embed.FS, outputLoc, fileExtensionPrefix string) error {
	wg := new(sync.WaitGroup)
	errs := new(sync.Map)

	wg.Add(1)
	go func() {
		defer wg.Done()
		tmpl, err := template.New("db.tmpl").Funcs(sprig.TxtFuncMap()).Funcs(Helpers).ParseFS(fs, "templates/db.tmpl")
		if err != nil {
			errs.Store("error parsing db template", err)
			return
		}

		if err := generate(&templateInfo{
			OutputDir: outputLoc,
		}, tmpl, outputLoc, fileExtensionPrefix); err != nil {
			errs.Store("error generating db template", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tmpl, err := template.New("errors.tmpl").Funcs(sprig.TxtFuncMap()).Funcs(Helpers).ParseFS(fs, "templates/errors.tmpl")
		if err != nil {
			errs.Store("error parsing errors template", err)
			return
		}

		if err := generate(&templateInfo{
			OutputDir: outputLoc,
		}, tmpl, outputLoc, fileExtensionPrefix); err != nil {
			errs.Store("error generating helpers template", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tmpl, err := template.New("helpers.tmpl").Funcs(sprig.TxtFuncMap()).Funcs(Helpers).ParseFS(fs, "templates/helpers.tmpl")
		if err != nil {
			errs.Store("error parsing helpers template", err)
			return
		}

		if err := generate(&templateInfo{
			OutputDir: outputLoc,
		}, tmpl, outputLoc, fileExtensionPrefix); err != nil {
			errs.Store("error generating helpers template", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tmpl, err := template.New("metrics.tmpl").Funcs(sprig.TxtFuncMap()).Funcs(Helpers).ParseFS(fs, "templates/metrics.tmpl")
		if err != nil {
			errs.Store("error parsing metrics template", err)
			return
		}

		if err := generate(&templateInfo{
			OutputDir: outputLoc,
		}, tmpl, outputLoc, fileExtensionPrefix); err != nil {
			errs.Store("error generating metrics template", err)
		}
	}()

	wg.Wait()

	errs.Range(func(key, value any) bool {
		if value != nil {
			slog.Error("Error rendering helpers", slog.String(logging.KeyError, value.(error).Error()))
		}

		return true
	})

	return nil
}

func generate(t *templateInfo, tmpl *template.Template, outputLoc, fileExtensionPrefix string) error {
	ext := ".go"
	if fileExtensionPrefix != "" {
		// Add a period if it's not already there
		if fileExtensionPrefix[0] != '.' {
			fileExtensionPrefix = "." + fileExtensionPrefix
		}
		ext = fileExtensionPrefix + ext
	}

	filename := strings.ReplaceAll(tmpl.Name(), ".tmpl", "")
	if t.Table != nil {
		filename = xstrings.ToSnakeCase(t.Table.Name)
	}

	fn := filepath.Join(outputLoc, filename+ext)
	if err := os.MkdirAll(filepath.Dir(fn), 0o750); err != nil {
		return err
	}

	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			slog.Warn("Error closing file", slog.String(logging.KeyError, err.Error()))
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

func FmtTemplates(outputLoc string) error {
	cmd := exec.Command("goimports", "-w", outputLoc)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
