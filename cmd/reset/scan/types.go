package scan

import (
	"go/ast"
	"go/types"
)

// Target описывает структуру, для которой нужно сгенерировать метод Reset.
type Target struct {
	TypeSpec   *ast.TypeSpec
	Struct     *types.Struct
	TypeName   *types.Named
	SourceBase string
}
