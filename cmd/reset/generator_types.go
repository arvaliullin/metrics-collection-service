package main

import (
	"go/token"
	"go/types"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

// resetIface - интерфейс с методом Reset() для проверки типов при генерации.
var resetIface types.Type

func init() {
	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
	m := types.NewFunc(token.NoPos, nil, "Reset", sig)
	resetIface = types.NewInterfaceType([]*types.Func{m}, nil)
	resetIface.(*types.Interface).Complete()
}

// hasResetMethod возвращает true, если тип или указатель на него реализует Reset().
func hasResetMethod(typ types.Type) bool {
	return types.Implements(typ, resetIface.(*types.Interface)) ||
		types.Implements(types.NewPointer(typ), resetIface.(*types.Interface))
}

// samePkg проверяет, принадлежит ли именованный тип указанному пакету.
func samePkg(typ types.Type, pkg *types.Package) bool {
	if pkg == nil {
		return false
	}
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Pkg() == pkg
}

// receiverName возвращает короткое имя приёмника по имени типа (например, ResetableStruct → rs).
func receiverName(typeName string) string {
	var b strings.Builder
	for i, r := range typeName {
		if unicode.IsUpper(r) && (i == 0 || (i > 0 && unicode.IsLower(rune(typeName[i-1])))) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	if b.Len() == 0 {
		return "r"
	}
	return b.String()
}

// pkgName возвращает имя пакета для объявления в сгенерированном файле.
func pkgName(pkg *packages.Package) string {
	if pkg.Name != "" {
		return pkg.Name
	}
	if pkg.Types != nil {
		return pkg.Types.Name()
	}
	return "main"
}
