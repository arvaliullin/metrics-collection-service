package gen

import (
	"go/token"
	"go/types"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

var resetIface types.Type

func init() {
	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
	m := types.NewFunc(token.NoPos, nil, "Reset", sig)
	resetIface = types.NewInterfaceType([]*types.Func{m}, nil)
	resetIface.(*types.Interface).Complete()
}

func hasResetMethod(typ types.Type) bool {
	return types.Implements(typ, resetIface.(*types.Interface)) ||
		types.Implements(types.NewPointer(typ), resetIface.(*types.Interface))
}

func as[T any](v any) (T, bool) {
	t, ok := v.(T)
	return t, ok
}

func samePkg(typ types.Type, pkg *types.Package) bool {
	if pkg == nil {
		return false
	}
	named, ok := as[*types.Named](typ)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Pkg() == pkg
}

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

func pkgName(pkg *packages.Package) string {
	if pkg.Name != "" {
		return pkg.Name
	}
	if pkg.Types != nil {
		return pkg.Types.Name()
	}
	return "main"
}
