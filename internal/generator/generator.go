package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	// Skip main.go generation - library will not interact with main.go anymore
	// if err := genMainGo(entities, cfg); err != nil {
	//     return err
	// }

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
		"Indexes":      e.Indexes,
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

	// Skip generating router and init snippets for main.go - library no longer interacts with main.go
	// routerSnippetPath := filepath.Join(cfg.ProjectRoot, "generated", "router_"+strings.ToLower(e.Name)+".go.snippet")
	// initSnippetPath := filepath.Join(cfg.ProjectRoot, "generated", "init_"+strings.ToLower(e.Name)+".go.snippet")
	//
	// targets = append(targets,
	//     struct{ path, tpl string }{path: routerSnippetPath, tpl: templates.MainRouter},
	//     struct{ path, tpl string }{path: initSnippetPath, tpl: templates.MainInit},
	// )

	for _, t := range targets {
		if err := writeIfNeeded(t.path, t.tpl, ctx, cfg); err != nil {
			return err
		}
	}

	// Generate/update constants file
	if err := updateConstantsFile(e, cfg); err != nil {
		return err
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
	t := template.Must(template.New("tpl").Funcs(template.FuncMap{
		"lower":              strings.ToLower,
		"generateValidation": generateValidation,
		"hasRequiredFields":  hasRequiredFields,
		"generateIndexes":    generateIndexes,
		"hasIndexes":         hasIndexes,
	}).Parse(tpl))
	if err := t.Execute(&buf, ctx); err != nil {
		return err
	}

	fmt.Printf("✅ Generated: %s\n", path)
	return os.WriteFile(path, buf.Bytes(), 0o644)
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

// hasRequiredFields checks if any field has required validation
func hasRequiredFields(fields []parser.Field) bool {
	for _, field := range fields {
		if strings.Contains(field.Validate, "required") {
			return true
		}
	}
	return false
}

// generateValidation generates validation code for fields with validate tags
func generateValidation(fields []parser.Field, entityLower string) string {
	var validations []string

	for _, field := range fields {
		if field.Validate == "" {
			continue
		}

		// Parse validation rules
		rules := strings.Split(field.Validate, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)

			switch {
			case rule == "required":
				validations = append(validations, generateRequiredValidation(field, entityLower))
			case strings.HasPrefix(rule, "min="):
				validations = append(validations, generateMinValidation(field, entityLower, rule))
			case strings.HasPrefix(rule, "max="):
				validations = append(validations, generateMaxValidation(field, entityLower, rule))
			case rule == "email":
				validations = append(validations, generateEmailValidation(field, entityLower))
			}
		}
	}

	if len(validations) == 0 {
		return ""
	}

	return "\t// Field validation\n" + strings.Join(validations, "\n")
}

func generateRequiredValidation(field parser.Field, entityLower string) string {
	fieldName := field.Name
	jsonTag := field.JSONTag
	if jsonTag == "" {
		jsonTag = strings.ToLower(fieldName)
	}

	switch field.Type {
	case "string":
		return fmt.Sprintf("\tif %sData.%s == \"\" {\n\t\treturn res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, \"VALIDATION_FAILED\", \"%s is required\"))\n\t}", entityLower, fieldName, jsonTag)
	case "int", "int32", "int64":
		return fmt.Sprintf("\tif %sData.%s == 0 {\n\t\treturn res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, \"VALIDATION_FAILED\", \"%s is required\"))\n\t}", entityLower, fieldName, jsonTag)
	default:
		// For other types, check for zero value using reflection-like approach
		return fmt.Sprintf("\t// TODO: Add validation for %s.%s (type: %s)", entityLower, fieldName, field.Type)
	}
}

func generateMinValidation(field parser.Field, entityLower string, rule string) string {
	minValue := strings.TrimPrefix(rule, "min=")
	fieldName := field.Name
	jsonTag := field.JSONTag
	if jsonTag == "" {
		jsonTag = strings.ToLower(fieldName)
	}

	switch field.Type {
	case "string":
		return fmt.Sprintf("\tif len(%sData.%s) < %s {\n\t\treturn res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, \"VALIDATION_FAILED\", \"%s must be at least %s characters\"))\n\t}", entityLower, fieldName, minValue, jsonTag, minValue)
	case "int", "int32", "int64":
		return fmt.Sprintf("\tif %sData.%s < %s {\n\t\treturn res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, \"VALIDATION_FAILED\", \"%s must be at least %s\"))\n\t}", entityLower, fieldName, minValue, jsonTag, minValue)
	default:
		return fmt.Sprintf("\t// TODO: Add min validation for %s.%s (type: %s)", entityLower, fieldName, field.Type)
	}
}

func generateMaxValidation(field parser.Field, entityLower string, rule string) string {
	maxValue := strings.TrimPrefix(rule, "max=")
	fieldName := field.Name
	jsonTag := field.JSONTag
	if jsonTag == "" {
		jsonTag = strings.ToLower(fieldName)
	}

	switch field.Type {
	case "string":
		return fmt.Sprintf("\tif len(%sData.%s) > %s {\n\t\treturn res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, \"VALIDATION_FAILED\", \"%s must be at most %s characters\"))\n\t}", entityLower, fieldName, maxValue, jsonTag, maxValue)
	case "int", "int32", "int64":
		return fmt.Sprintf("\tif %sData.%s > %s {\n\t\treturn res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, \"VALIDATION_FAILED\", \"%s must be at most %s\"))\n\t}", entityLower, fieldName, maxValue, jsonTag, maxValue)
	default:
		return fmt.Sprintf("\t// TODO: Add max validation for %s.%s (type: %s)", entityLower, fieldName, field.Type)
	}
}

func generateEmailValidation(field parser.Field, entityLower string) string {
	fieldName := field.Name
	jsonTag := field.JSONTag
	if jsonTag == "" {
		jsonTag = strings.ToLower(fieldName)
	}

	return fmt.Sprintf("\tif %sData.%s != \"\" && !isValidEmail(%sData.%s) {\n\t\treturn res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, \"VALIDATION_FAILED\", \"%s must be a valid email address\"))\n\t}", entityLower, fieldName, entityLower, fieldName, jsonTag)
}

// hasIndexes checks if entity has any indexes defined
func hasIndexes(fields []parser.Field, indexes []parser.Index) bool {
	// Check field-level indexes
	for _, field := range fields {
		if field.Index != "" {
			return true
		}
	}
	// Check compound indexes
	return len(indexes) > 0
}

// generateIndexes generates index creation code
func generateIndexes(fields []parser.Field, indexes []parser.Index, entityLower string) string {
	var indexCreations []string
	hasIndexes := false

	// Generate field-level indexes
	for _, field := range fields {
		if field.Index == "" {
			continue
		}

		bsonField := field.BSONTag
		if bsonField == "" {
			bsonField = strings.ToLower(field.Name)
		}

		indexCode := generateFieldIndex(field, bsonField, entityLower)
		if indexCode != "" {
			indexCreations = append(indexCreations, indexCode)
			hasIndexes = true
		}
	}

	// Generate compound indexes
	for _, index := range indexes {
		indexCode := generateCompoundIndex(index, entityLower)
		if indexCode != "" {
			indexCreations = append(indexCreations, indexCode)
			hasIndexes = true
		}
	}

	if !hasIndexes {
		return "\t// No indexes defined"
	}

	// Add error variable declaration at the beginning
	result := "\tvar err error\n\n" + strings.Join(indexCreations, "\n\n")
	return result
}

func generateFieldIndex(field parser.Field, bsonField, entityLower string) string {
	indexType := field.Index

	var indexDoc string
	var options []string

	switch indexType {
	case "1":
		indexDoc = fmt.Sprintf("bson.D{{Key: \"%s\", Value: 1}}", bsonField)
	case "-1":
		indexDoc = fmt.Sprintf("bson.D{{Key: \"%s\", Value: -1}}", bsonField)
	case "text":
		indexDoc = fmt.Sprintf("bson.D{{Key: \"%s\", Value: \"text\"}}", bsonField)
	case "unique":
		indexDoc = fmt.Sprintf("bson.D{{Key: \"%s\", Value: 1}}", bsonField)
		options = append(options, "Unique: utils.GetPointer(true)")
	case "sparse":
		indexDoc = fmt.Sprintf("bson.D{{Key: \"%s\", Value: 1}}", bsonField)
		options = append(options, "Sparse: utils.GetPointer(true)")
	default:
		// Try to parse as direction
		if indexType == "1" || indexType == "-1" {
			indexDoc = fmt.Sprintf("bson.D{{Key: \"%s\", Value: %s}}", bsonField, indexType)
		} else {
			return fmt.Sprintf("\t// TODO: Unsupported index type '%s' for field %s", indexType, field.Name)
		}
	}

	var optionsStr string
	if len(options) > 0 {
		optionsStr = fmt.Sprintf(", &options.IndexOptions{\n\t\t%s,\n\t}", strings.Join(options, ",\n\t\t"))
	} else {
		optionsStr = ", nil"
	}

	return fmt.Sprintf("\t// Index for %s field\n\terr = %sCollection.CreateIndex(%s%s)\n\tif err != nil {\n\t\treturn err\n\t}", field.Name, entityLower, indexDoc, optionsStr)
}

func generateCompoundIndex(index parser.Index, entityLower string) string {
	if len(index.Fields) == 0 {
		return ""
	}

	var indexFields []string
	for _, field := range index.Fields {
		if field.Type != "" {
			// Special index type (text, 2dsphere, etc.)
			indexFields = append(indexFields, fmt.Sprintf("{Key: \"%s\", Value: \"%s\"}", field.Name, field.Type))
		} else {
			// Direction-based index
			indexFields = append(indexFields, fmt.Sprintf("{Key: \"%s\", Value: %d}", field.Name, field.Direction))
		}
	}

	indexDoc := fmt.Sprintf("bson.D{\n\t\t%s,\n\t}", strings.Join(indexFields, ",\n\t\t"))

	var options []string
	if index.Unique {
		options = append(options, "Unique: utils.GetPointer(true)")
	}
	if index.Sparse {
		options = append(options, "Sparse: utils.GetPointer(true)")
	}
	if index.Name != "" {
		options = append(options, fmt.Sprintf("Name: utils.GetPointer(\"%s\")", index.Name))
	}

	var optionsStr string
	if len(options) > 0 {
		optionsStr = fmt.Sprintf(", &options.IndexOptions{\n\t\t%s,\n\t}", strings.Join(options, ",\n\t\t"))
	} else {
		optionsStr = ", nil"
	}

	// Generate comment describing the compound index
	var fieldNames []string
	for _, field := range index.Fields {
		direction := "asc"
		if field.Direction == -1 {
			direction = "desc"
		}
		if field.Type != "" {
			fieldNames = append(fieldNames, fmt.Sprintf("%s(%s)", field.Name, field.Type))
		} else {
			fieldNames = append(fieldNames, fmt.Sprintf("%s(%s)", field.Name, direction))
		}
	}

	comment := fmt.Sprintf("Compound index: %s", strings.Join(fieldNames, ", "))
	if index.Unique {
		comment += " (unique)"
	}
	if index.Sparse {
		comment += " (sparse)"
	}

	return fmt.Sprintf("\t// %s\n\terr = %sCollection.CreateIndex(%s%s)\n\tif err != nil {\n\t\treturn err\n\t}", comment, entityLower, indexDoc, optionsStr)
}

// updateConstantsFile adds or updates constants for the entity
func updateConstantsFile(entity parser.Entity, cfg Config) error {
	constantsPath := filepath.Join(cfg.ProjectRoot, "constants", "constants.go")
	constantName := fmt.Sprintf("Param%sID", entity.Name)
	constantValue := fmt.Sprintf("%s_id", strings.ToLower(entity.Name))

	// Check if constants file exists
	if _, err := os.Stat(constantsPath); os.IsNotExist(err) {
		// Create new constants file
		return createConstantsFile(constantsPath, entity, cfg)
	}

	// Read existing constants file
	content, err := os.ReadFile(constantsPath)
	if err != nil {
		return fmt.Errorf("failed to read constants file: %v", err)
	}

	contentStr := string(content)

	// Check if constant already exists using more precise matching
	constantPattern := fmt.Sprintf(`\b%s\s*=`, constantName)
	matched, _ := regexp.MatchString(constantPattern, contentStr)
	if matched {
		if cfg.DryRun {
			fmt.Printf("constant %s already exists in: %s\n", constantName, constantsPath)
		}
		return nil // Constant already exists
	}

	if cfg.DryRun {
		fmt.Printf("would add constant %s to: %s\n", constantName, constantsPath)
		return nil
	}

	// Add constant to the end of the const block
	newConstant := fmt.Sprintf("\t%s = \"%s\"\n", constantName, constantValue)

	// Find the const block and insert before its closing parenthesis
	if strings.Contains(contentStr, "const (") {
		// Find the last closing parenthesis that belongs to a const block
		constIndex := strings.Index(contentStr, "const (")
		if constIndex != -1 {
			// Find the matching closing parenthesis
			afterConst := contentStr[constIndex:]
			parenIndex := strings.LastIndex(afterConst, ")")
			if parenIndex != -1 {
				// Insert before the closing parenthesis
				insertPos := constIndex + parenIndex
				contentStr = contentStr[:insertPos] + newConstant + contentStr[insertPos:]
			} else {
				// No closing parenthesis found, append after const (
				insertPos := constIndex + len("const (") + 1
				contentStr = contentStr[:insertPos] + newConstant + contentStr[insertPos:]
			}
		}
	} else {
		// No const block found, create one or append to end
		if strings.TrimSpace(contentStr) == "" || !strings.Contains(contentStr, "package") {
			// Empty file or no package declaration
			contentStr += fmt.Sprintf("const (\n%s)\n", newConstant)
		} else {
			// Append const block to end of file
			contentStr = strings.TrimRight(contentStr, "\n") + "\n\n" + fmt.Sprintf("const (\n%s)\n", newConstant)
		}
	}

	fmt.Printf("✅ Updated constants: %s (added %s)\n", constantsPath, constantName)
	return os.WriteFile(constantsPath, []byte(contentStr), 0o644)
}

// createConstantsFile creates a new constants file with the entity constant
func createConstantsFile(constantsPath string, entity parser.Entity, cfg Config) error {
	if cfg.DryRun {
		fmt.Printf("would create constants file: %s\n", constantsPath)
		return nil
	}

	// Create constants directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(constantsPath), 0o755); err != nil {
		return fmt.Errorf("failed to create constants directory: %v", err)
	}

	constantName := fmt.Sprintf("Param%sID", entity.Name)
	constantValue := fmt.Sprintf("%s_id", strings.ToLower(entity.Name))

	content := fmt.Sprintf(`package constants

// API parameter constants
const (
	%s = "%s"
)
`, constantName, constantValue)

	fmt.Printf("✅ Created constants file: %s\n", constantsPath)
	return os.WriteFile(constantsPath, []byte(content), 0o644)
}
