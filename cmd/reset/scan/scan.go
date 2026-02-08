package scan

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Scanner находит в пакете структуры с маркером generate:reset.
type Scanner interface {
	Scan(pkg *packages.Package) ([]*Target, error)
}

// DefaultScanner реализует Scanner, используя AST и type info пакета.
type DefaultScanner struct{}

// Scan находит в пакете все структуры с комментарием generate:reset.
func (DefaultScanner) Scan(pkg *packages.Package) ([]*Target, error) {
	if len(pkg.Errors) > 0 || pkg.TypesInfo == nil || pkg.Fset == nil {
		return nil, nil
	}
	var out []*Target
	for _, f := range pkg.Syntax {
		if f == nil {
			continue
		}
		for _, d := range f.Decls {
			target := tryTargetFromDecl(d, pkg)
			if target != nil {
				out = append(out, target)
			}
		}
	}
	return out, nil
}

func tryTargetFromDecl(d ast.Decl, pkg *packages.Package) *Target {
	gen, ok := as[*ast.GenDecl](d)
	if !ok || gen.Tok != token.TYPE || !hasGenerateResetComment(gen.Doc) || len(gen.Specs) != 1 {
		return nil
	}
	spec, ok := as[*ast.TypeSpec](gen.Specs[0])
	if !ok || spec.Name == nil {
		return nil
	}
	if _, ok := as[*ast.StructType](spec.Type); !ok {
		return nil
	}
	obj := pkg.TypesInfo.Defs[spec.Name]
	if obj == nil {
		return nil
	}
	typeName, ok := as[*types.TypeName](obj)
	if !ok {
		return nil
	}
	named, ok := as[*types.Named](typeName.Type())
	if !ok {
		return nil
	}
	st, ok := as[*types.Struct](named.Underlying())
	if !ok {
		return nil
	}
	return &Target{
		TypeSpec:   spec,
		Struct:     st,
		TypeName:   named,
		SourceBase: sourceBaseFromPos(pkg.Fset, spec.Name.Pos()),
	}
}

func sourceBaseFromPos(fset *token.FileSet, pos token.Pos) string {
	if !pos.IsValid() {
		return "reset"
	}
	file := fset.File(pos)
	if file == nil {
		return "reset"
	}
	base := strings.TrimSuffix(filepath.Base(file.Name()), ".go")
	if base == "" {
		return "reset"
	}
	return base
}

func as[T any](v any) (T, bool) {
	t, ok := v.(T)
	return t, ok
}

func hasGenerateResetComment(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		if strings.Contains(strings.TrimSpace(c.Text), GenerateResetMarker) {
			return true
		}
	}
	return false
}
