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
}

type Entity struct {
	PkgPath string
	Name    string
	Plural  string
	DBName  string
	Fields  []Field
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
					jsonTag, bsonTag, validate := parseTags(f.Tag)
					fields = append(fields, Field{Name: name, Type: typ, JSONTag: jsonTag, BSONTag: bsonTag, Validate: validate})
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
			})
		}
		return true
	})
	return out, nil
}

func relModelPath(path string) string {
	path = filepath.ToSlash(path)
	idx := strings.Index(path, "/model/")
	if idx < 0 {
		return "model"
	}
	return strings.TrimSuffix(path[idx+1:], "/data.go")
}

func parseTags(tag *ast.BasicLit) (jsonTag, bsonTag, validate string) {
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
