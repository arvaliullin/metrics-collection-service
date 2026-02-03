package exitinmain

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer запрещает прямой вызов os.Exit в функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name:     "exitinmain",
	Doc:      "запрещает прямой вызов os.Exit в функции main пакета main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}
	var inMain bool
	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch x := n.(type) {
		case *ast.FuncDecl:
			inMain = x.Name != nil && x.Name.Name == "main"
		case *ast.CallExpr:
			if !inMain {
				return
			}
			sel, ok := x.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel == nil || sel.Sel.Name != "Exit" {
				return
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok || id.Name != "os" {
				return
			}
			if pass.TypesInfo != nil && id.Obj == nil {
				uses := pass.TypesInfo.Uses
				if uses != nil {
					if obj := uses[id]; obj != nil {
						if pkg := obj.Pkg(); pkg != nil && pkg.Path() != "os" {
							return
						}
					}
				}
			}
			pass.Reportf(x.Pos(), "прямой вызов os.Exit в main пакета main запрещён")
		}
	})
	return nil, nil
}
