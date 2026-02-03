package main

import (
	"go/ast"
	"go/token"
	"go/types"
)

// zeroValueExpr возвращает AST-выражение нулевого значения для заданного типа.
func zeroValueExpr(t types.Type) ast.Expr {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return &ast.Ident{Name: "false"}
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
			return &ast.Ident{Name: "0"}
		case types.Float32, types.Float64:
			return &ast.Ident{Name: "0"}
		case types.Complex64, types.Complex128:
			return &ast.Ident{Name: "0"}
		case types.String:
			return &ast.BasicLit{Kind: token.STRING, Value: `""`}
		default:
			return &ast.Ident{Name: "0"}
		}
	default:
		return &ast.Ident{Name: "nil"}
	}
}

// fieldSelector строит AST селектора поля (recv.fieldName).
func fieldSelector(recv, fieldName string) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X:   &ast.Ident{Name: recv},
		Sel: &ast.Ident{Name: fieldName},
	}
}

// nilCheckStmt строит AST условия "if sel != nil { body }".
func nilCheckStmt(sel ast.Expr, body ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  sel,
			Op: token.NEQ,
			Y:  &ast.Ident{Name: "nil"},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{body}},
	}
}

// assignStmt строит AST присваивания lhs = rhs.
func assignStmt(lhs, rhs ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{lhs},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{rhs},
	}
}

// callResetStmt строит AST вызова recv.Reset().
func callResetStmt(recv ast.Expr) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun:  &ast.SelectorExpr{X: recv, Sel: &ast.Ident{Name: "Reset"}},
			Args: nil,
		},
	}
}

// typeAssertResetStmt строит AST проверки типа interface{ Reset() } и вызова Reset().
func typeAssertResetStmt(sel ast.Expr) *ast.IfStmt {
	interfaceType := &ast.InterfaceType{
		Methods: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "Reset"}},
				Type:  &ast.FuncType{Params: &ast.FieldList{}, Results: &ast.FieldList{}},
			}},
		},
	}
	typeAssert := &ast.TypeAssertExpr{X: sel, Type: interfaceType}
	assign := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "r"}, &ast.Ident{Name: "ok"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{typeAssert},
	}
	return &ast.IfStmt{
		Init: assign,
		Cond: &ast.Ident{Name: "ok"},
		Body: &ast.BlockStmt{List: []ast.Stmt{callResetStmt(&ast.Ident{Name: "r"})}},
	}
}

// typeAssertResetIfStmt оборачивает typeAssertResetStmt в проверку на nil.
func typeAssertResetIfStmt(sel ast.Expr) *ast.IfStmt {
	return nilCheckStmt(sel, typeAssertResetStmt(sel))
}
