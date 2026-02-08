package gen

import (
	"go/token"
	"go/types"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/tools/go/packages"
)

var (
	resetIface *types.Interface
	resetOnce  sync.Once
)

// ensureResetIface инициализирует resetIface при первом вызове (интерфейс с методом Reset).
func ensureResetIface() {
	resetOnce.Do(func() {
		sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
		m := types.NewFunc(token.NoPos, nil, "Reset", sig)
		resetIface = types.NewInterfaceType([]*types.Func{m}, nil)
		resetIface.Complete()
	})
}

// hasResetMethod проверяет, реализует ли тип или указатель на тип интерфейс с методом Reset.
func hasResetMethod(typ types.Type) bool {
	ensureResetIface()
	return types.Implements(typ, resetIface) ||
		types.Implements(types.NewPointer(typ), resetIface)
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
