package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

type Field struct {
	Name     string
	Type     string
	JSONTag  string
	BSONTag  string
	Validate string
	Index    string // Index definition: "1", "-1", "text", "unique", etc.
}

type Index struct {
	Fields []IndexField // Fields in the index
	Unique bool         // Whether the index is unique
	Sparse bool         // Whether the index is sparse
	Name   string       // Custom index name (optional)
}

type IndexField struct {
	Name      string // Field name
	Direction int    // 1 for ascending, -1 for descending
	Type      string // "text", "2dsphere", etc. (optional)
}

type Entity struct {
	PkgPath string
	Name    string
	Plural  string
	DBName  string
	Fields  []Field
	Indexes []Index // Compound indexes defined via comments
}

func ParseDataGo(path string) ([]Entity, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	pkgRel := relModelPath(path)
	var out []Entity

	ast.Inspect(f, func(n ast.Node) bool {
		gd, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			s, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			var isEntity bool
			var dbName string
			var indexes []Index

			// Check both GenDecl.Doc and TypeSpec.Doc
			var docComments *ast.CommentGroup
			if gd.Doc != nil {
				docComments = gd.Doc
			} else if ts.Doc != nil {
				docComments = ts.Doc
			}

			if docComments != nil {
				for _, c := range docComments.List {
					text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
					if strings.HasPrefix(text, "@entity") {
						isEntity = true
						parts := strings.Fields(text)
						for _, p := range parts {
							if strings.HasPrefix(p, "db:") {
								dbName = strings.TrimPrefix(p, "db:")
							}
						}
					} else if strings.HasPrefix(text, "@index") {
						// Parse index definition: @index field1:1,field2:-1 unique sparse name:custom_name
						idx := parseIndexComment(text)
						if idx != nil {
							indexes = append(indexes, *idx)
						}
					}
				}
			}
			if !isEntity {
				continue
			}

			var fields []Field
			if s.Fields != nil {
				for _, f := range s.Fields.List {
					if len(f.Names) == 0 {
						continue
					}
					name := f.Names[0].Name
					typ := exprString(f.Type)
					jsonTag, bsonTag, validate, index := parseTags(f.Tag)
					fields = append(fields, Field{Name: name, Type: typ, JSONTag: jsonTag, BSONTag: bsonTag, Validate: validate, Index: index})
				}
			}

			entName := ts.Name.Name
			if dbName == "" {
				dbName = defaultDBName(entName)
			}
			out = append(out, Entity{
				PkgPath: pkgRel,
				Name:    entName,
				Plural:  naivePlural(entName),
				DBName:  dbName,
				Fields:  fields,
				Indexes: indexes,
			})
		}
		return true
	})
	return out, nil
}

func relModelPath(path string) string {
	path = filepath.ToSlash(path)

	// Look for /model/ in the path
	idx := strings.Index(path, "/model/")
	if idx >= 0 {
		// Found /model/, extract everything after it
		return strings.TrimSuffix(path[idx+1:], "/data.go")
	}

	// If path starts with model/ (relative path)
	if strings.HasPrefix(path, "model/") {
		return strings.TrimSuffix(path, "/data.go")
	}

	// Fallback: extract directory if it contains model
	dir := filepath.Dir(path)
	if strings.Contains(dir, "model") {
		return dir
	}

	return "model"
}

func parseTags(tag *ast.BasicLit) (jsonTag, bsonTag, validate, index string) {
	if tag == nil {
		return
	}
	raw := strings.Trim(tag.Value, "`")
	for _, kv := range strings.Split(raw, " ") {
		parts := strings.SplitN(kv, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := parts[0]
		v := strings.Trim(parts[1], "\"")
		switch k {
		case "json":
			jsonTag = v
		case "bson":
			bsonTag = v
		case "validate":
			validate = v
		case "index":
			index = v
		}
	}
	return
}

func exprString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprString(t.X)
	case *ast.ArrayType:
		return "[]" + exprString(t.Elt)
	case *ast.SelectorExpr:
		return exprString(t.X) + "." + t.Sel.Name
	default:
		return "interface{}"
	}
}

func defaultDBName(name string) string {
	var b []rune
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b = append(b, '_')
		}
		b = append(b, rune(strings.ToLower(string(r))[0]))
	}
	return string(b) + "s"
}

func naivePlural(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	return s + "s"
}

// parseIndexComment parses index definition from comment
// Format: @index field1:1,field2:-1 unique sparse name:custom_name
func parseIndexComment(comment string) *Index {
	// Remove @index prefix
	comment = strings.TrimSpace(strings.TrimPrefix(comment, "@index"))
	if comment == "" {
		return nil
	}

	parts := strings.Fields(comment)
	if len(parts) == 0 {
		return nil
	}

	index := &Index{}

	// First part should be field definitions
	fieldDefs := parts[0]
	fieldPairs := strings.Split(fieldDefs, ",")

	for _, pair := range fieldPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Parse field:direction or field:type
		colonIdx := strings.Index(pair, ":")
		if colonIdx == -1 {
			// Just field name, default to ascending
			index.Fields = append(index.Fields, IndexField{
				Name:      pair,
				Direction: 1,
			})
		} else {
			fieldName := pair[:colonIdx]
			value := pair[colonIdx+1:]

			indexField := IndexField{Name: fieldName}

			// Try to parse as direction first
			if value == "1" {
				indexField.Direction = 1
			} else if value == "-1" {
				indexField.Direction = -1
			} else {
				// It's a type (text, 2dsphere, etc.)
				indexField.Type = value
				indexField.Direction = 1 // Default direction
			}

			index.Fields = append(index.Fields, indexField)
		}
	}

	// Parse options from remaining parts
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		switch {
		case part == "unique":
			index.Unique = true
		case part == "sparse":
			index.Sparse = true
		case strings.HasPrefix(part, "name:"):
			index.Name = strings.TrimPrefix(part, "name:")
		}
	}

	return index
}
