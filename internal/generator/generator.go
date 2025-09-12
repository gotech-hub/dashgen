package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gotech-hub/dashgen/internal/parser"

	"github.com/gotech-hub/dashgen/internal/templates"
)

type Config struct {
	ModulePath  string
	ProjectRoot string
	Force       bool
	DryRun      bool
}

func Generate(entities []parser.Entity, cfg Config) error {
	for _, e := range entities {
		if err := genOne(e, cfg); err != nil {
			return err
		}
	}

	// Generate main.go with all entities
	if err := genMainGo(entities, cfg); err != nil {
		return err
	}

	return nil
}

func genOne(e parser.Entity, cfg Config) error {
	ctx := map[string]any{
		"Module":       cfg.ModulePath,
		"PkgPath":      e.PkgPath,
		"Entity":       e.Name,
		"EntityLower":  strings.ToLower(e.Name[:1]) + e.Name[1:],
		"EntitySnake":  toSnake(e.Name),
		"EntityPlural": e.Plural,
		"DBName":       e.DBName,
		"Fields":       e.Fields,
	}

	modelPath := strings.ReplaceAll(e.PkgPath, "model/", "")

	targets := []struct{ path, tpl string }{
		{path: filepath.Join(cfg.ProjectRoot, "model", modelPath, "init.go"), tpl: templates.ModelInit},
		{path: filepath.Join(cfg.ProjectRoot, "model", modelPath, "repository.go"), tpl: templates.ModelRepository},
		{path: filepath.Join(cfg.ProjectRoot, "internal/action", strings.ToLower(e.Name)+".go"), tpl: templates.Action},
		{path: filepath.Join(cfg.ProjectRoot, "internal/api", strings.ToLower(e.Name)+".go"), tpl: templates.API},
		{path: filepath.Join(cfg.ProjectRoot, "client", strings.ToLower(e.Name)+".go"), tpl: templates.Client},
	}

	// Generate router and init snippets for main.go
	routerSnippetPath := filepath.Join(cfg.ProjectRoot, "generated", "router_"+strings.ToLower(e.Name)+".go.snippet")
	initSnippetPath := filepath.Join(cfg.ProjectRoot, "generated", "init_"+strings.ToLower(e.Name)+".go.snippet")

	targets = append(targets,
		struct{ path, tpl string }{path: routerSnippetPath, tpl: templates.MainRouter},
		struct{ path, tpl string }{path: initSnippetPath, tpl: templates.MainInit},
	)

	for _, t := range targets {
		if err := writeIfNeeded(t.path, t.tpl, ctx, cfg); err != nil {
			return err
		}
	}
	return nil
}

func writeIfNeeded(path, tpl string, ctx map[string]any, cfg Config) error {
	// Check if file already exists (unless force is enabled)
	if !cfg.Force {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("⚠️  File already exists, skipping: %s\n", path)
			return nil
		}
	}

	if cfg.DryRun {
		fmt.Println("would write:", path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	t := template.Must(template.New("tpl").Funcs(template.FuncMap{"lower": strings.ToLower}).Parse(tpl))
	if err := t.Execute(&buf, ctx); err != nil {
		return err
	}

	fmt.Printf("✅ Generated: %s\n", path)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func genMainGo(entities []parser.Entity, cfg Config) error {
	mainPath := filepath.Join(cfg.ProjectRoot, "main.go")

	// Check if main.go already exists
	if _, err := os.Stat(mainPath); err == nil && !cfg.Force {
		fmt.Printf("⚠️  main.go already exists, skipping (use -force to overwrite)\n")
		return nil
	}

	if cfg.DryRun {
		fmt.Println("would write:", mainPath)
		return nil
	}

	ctx := map[string]any{
		"Module":   cfg.ModulePath,
		"Entities": entities,
	}

	var buf bytes.Buffer
	t := template.Must(template.New("main").Funcs(template.FuncMap{"lower": strings.ToLower}).Parse(templates.MainGo))
	if err := t.Execute(&buf, ctx); err != nil {
		return err
	}

	fmt.Printf("✅ Generated: %s\n", mainPath)
	return os.WriteFile(mainPath, buf.Bytes(), 0o644)
}

func toSnake(in string) string {
	var out []rune
	for i, r := range in {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '_')
		}
		out = append(out, rune(strings.ToLower(string(r))[0]))
	}
	return string(out)
}
