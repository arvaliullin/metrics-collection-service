package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// generateResetMarker - маркер в комментарии для включения структуры в генерацию.
const generateResetMarker = "generate:reset"

// hasGenerateResetComment проверяет, содержит ли группа комментариев маркер generate:reset.
func hasGenerateResetComment(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		if strings.Contains(strings.TrimSpace(c.Text), generateResetMarker) {
			return true
		}
	}
	return false
}

// targetStruct описывает структуру, для которой нужно сгенерировать метод Reset.
type targetStruct struct {
	TypeSpec   *ast.TypeSpec
	Struct     *types.Struct
	TypeName   *types.Named
	SourceBase string
}

// scanPackage находит в пакете все структуры с комментарием generate:reset.
func scanPackage(pkg *packages.Package) []*targetStruct {
	if len(pkg.Errors) > 0 || pkg.TypesInfo == nil || pkg.Fset == nil {
		return nil
	}
	var out []*targetStruct
	for _, f := range pkg.Syntax {
		if f == nil {
			continue
		}
		for _, d := range f.Decls {
			gen, ok := d.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			if !hasGenerateResetComment(gen.Doc) {
				continue
			}
			if len(gen.Specs) != 1 {
				continue
			}
			spec, ok := gen.Specs[0].(*ast.TypeSpec)
			if !ok || spec.Name == nil {
				continue
			}
			_, isStruct := spec.Type.(*ast.StructType)
			if !isStruct {
				continue
			}
			obj := pkg.TypesInfo.Defs[spec.Name]
			if obj == nil {
				continue
			}
			typeName, ok := obj.(*types.TypeName)
			if !ok {
				continue
			}
			named, ok := typeName.Type().(*types.Named)
			if !ok {
				continue
			}
			under := named.Underlying()
			st, ok := under.(*types.Struct)
			if !ok {
				continue
			}
			sourceBase := ""
			if pos := spec.Name.Pos(); pos.IsValid() {
				if file := pkg.Fset.File(pos); file != nil {
					sourceBase = strings.TrimSuffix(filepath.Base(file.Name()), ".go")
				}
			}
			if sourceBase == "" {
				sourceBase = "reset"
			}
			out = append(out, &targetStruct{TypeSpec: spec, Struct: st, TypeName: named, SourceBase: sourceBase})
		}
	}
	return out
}
