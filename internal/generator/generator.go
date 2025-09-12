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

	// Use the full PkgPath for model files (e.g., "model/user" -> "model/user/")
	modelDir := e.PkgPath

	targets := []struct{ path, tpl string }{
		{path: filepath.Join(cfg.ProjectRoot, modelDir, "init.go"), tpl: templates.ModelInit},
		{path: filepath.Join(cfg.ProjectRoot, modelDir, "repository.go"), tpl: templates.ModelRepository},
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
	if _, err := os.Stat(mainPath); err == nil {
		// File exists, try to update it
		return updateMainGo(mainPath, entities, cfg)
	}

	// File doesn't exist, create new one
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

func updateMainGo(mainPath string, entities []parser.Entity, cfg Config) error {
	// Read existing main.go
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return err
	}

	if cfg.DryRun {
		fmt.Printf("would update: %s\n", mainPath)
		return nil
	}

	contentStr := string(content)

	// Generate imports for new entities
	var newImports []string
	var newInits []string
	var newRoutes []string

	for _, entity := range entities {
		// Generate import - use correct package path
		pkgPath := entity.PkgPath
		importAlias := ""

		// Check if we need alias (when package name != entity name)
		pkgName := filepath.Base(pkgPath)
		if pkgName != strings.ToLower(entity.Name) {
			importAlias = fmt.Sprintf("app%s ", entity.Name)
		}

		newImports = append(newImports, fmt.Sprintf("\t%s\"%s/%s\"", importAlias, cfg.ModulePath, pkgPath))

		// Generate init call
		var initCall string
		if importAlias != "" {
			initCall = fmt.Sprintf("\tapp%s.Init(database)", entity.Name)
		} else {
			initCall = fmt.Sprintf("\t%s.Init(database)", pkgName)
		}
		newInits = append(newInits, initCall)

		// Generate routes
		entityLower := strings.ToLower(entity.Name)
		routes := []string{
			fmt.Sprintf("\t// Register %s API routes", entity.Name),
			fmt.Sprintf("\t// Core CRUD operations"),
			fmt.Sprintf("\tserver.SetHandler(common.APIMethod.POST, \"/v1/%s\", api.Create%s)", entityLower, entity.Name),
			fmt.Sprintf("\tserver.SetHandler(common.APIMethod.GET, \"/v1/%s\", api.Get%sByID)", entityLower, entity.Name),
			fmt.Sprintf("\tserver.SetHandler(common.APIMethod.QUERY, \"/v1/%ss\", api.Query%s)", entityLower, entity.Plural),
			fmt.Sprintf("\tserver.SetHandler(common.APIMethod.PUT, \"/v1/%s\", api.Update%s)", entityLower, entity.Name),
			fmt.Sprintf("\tserver.SetHandler(common.APIMethod.DELETE, \"/v1/%s\", api.Delete%s)", entityLower, entity.Name),
			"",
		}
		newRoutes = append(newRoutes, strings.Join(routes, "\n"))
	}

	// Update imports section (add before the closing parenthesis of imports)
	importMarker := ")"
	if len(newImports) > 0 {
		importSection := strings.Join(newImports, "\n") + "\n"
		contentStr = strings.Replace(contentStr, importMarker, importSection+importMarker, 1)
	}

	// Update database init section
	dbMarker := "/*{{ Register database generate }}*/"
	if strings.Contains(contentStr, dbMarker) && len(newInits) > 0 {
		initSection := strings.Join(newInits, "\n") + "\n\n\t" + dbMarker
		contentStr = strings.Replace(contentStr, "\t"+dbMarker, "\t"+initSection, 1)
	}

	// Update routes section
	routeMarker := "/*{{ Register User API routes generate }}*/"
	if strings.Contains(contentStr, routeMarker) && len(newRoutes) > 0 {
		routeSection := strings.Join(newRoutes, "\n") + "\n\t" + routeMarker
		contentStr = strings.Replace(contentStr, "\t"+routeMarker, "\t"+routeSection, 1)
	}

	fmt.Printf("✅ Updated: %s\n", mainPath)
	return os.WriteFile(mainPath, []byte(contentStr), 0o644)
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
