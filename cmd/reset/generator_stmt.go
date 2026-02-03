package main

import (
	"go/ast"
	"go/token"
	"go/types"
)

// generateResetBody возвращает список AST-операторов тела метода Reset() для структуры.
func generateResetBody(recv string, st *types.Struct, pkg *types.Package) []ast.Stmt {
	var list []ast.Stmt
	list = append(list, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: recv},
			Op: token.EQL,
			Y:  &ast.Ident{Name: "nil"},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{&ast.ReturnStmt{}},
		},
	})
	for f := range st.Fields() {
		if !f.IsField() {
			continue
		}
		fieldName := f.Name()
		sel := fieldSelector(recv, fieldName)
		stmt := stmtForField(recv, fieldName, f.Type(), sel, pkg)
		if stmt != nil {
			list = append(list, stmt)
		}
	}
	return list
}

// stmtForField возвращает AST-оператор сброса для одного поля в зависимости от его типа.
func stmtForField(recv, fieldName string, typ types.Type, sel ast.Expr, pkg *types.Package) ast.Stmt {
	switch t := typ.Underlying().(type) {
	case *types.Basic:
		return assignStmt(sel, zeroValueExpr(t))
	case *types.Slice:
		return assignStmt(sel, &ast.SliceExpr{
			X:    fieldSelector(recv, fieldName),
			High: &ast.BasicLit{Kind: token.INT, Value: "0"},
		})
	case *types.Map:
		return &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun:  &ast.Ident{Name: "clear"},
				Args: []ast.Expr{sel},
			},
		}
	case *types.Pointer:
		elem := t.Elem()
		switch elem.Underlying().(type) {
		case *types.Basic:
			return nilCheckStmt(sel, assignStmt(&ast.StarExpr{X: sel}, zeroValueExpr(elem)))
		default:
			if samePkg(elem, pkg) || hasResetMethod(elem) {
				return nilCheckStmt(sel, callResetStmt(sel))
			}
			return nilCheckStmt(sel, typeAssertResetStmt(sel))
		}
	case *types.Struct:
		if hasResetMethod(typ) {
			return callResetStmt(sel)
		}
		return nil
	case *types.Interface:
		return typeAssertResetIfStmt(sel)
	default:
		return nil
	}
}
