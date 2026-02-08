package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/cmd/reset/scan"

	"golang.org/x/tools/go/packages"
)

func TestDefaultGenerator_Generate(t *testing.T) {
	dir := t.TempDir()
	goMod := `module testpkg
go 1.21
`
	src := `package testpkg

// generate:reset
type S struct {
	I int
	Slice []int
	M     map[string]string
	Child *S
}
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "s.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &packages.Config{
		Dir:  dir,
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 || len(pkgs[0].Errors) > 0 {
		t.Fatalf("load: %v", err)
	}
	pkg := pkgs[0]

	var sc scan.DefaultScanner
	targets, err := sc.Scan(pkg)
	if err != nil || len(targets) != 1 {
		t.Fatalf("scan: err=%v targets=%d", err, len(targets))
	}

	var g DefaultGenerator
	out, err := g.Generate(pkg, targets)
	if err != nil {
		t.Fatal(err)
	}
	code := string(out)
	if !strings.Contains(code, "func (s *S) Reset()") {
		t.Errorf("expected method (s *S) Reset(), got:\n%s", code)
	}
	if !strings.Contains(code, "s.I = 0") {
		t.Errorf("expected s.I = 0 in output")
	}
	if !strings.Contains(code, "s.Slice = s.Slice[:0]") {
		t.Errorf("expected slice reset in output")
	}
	if !strings.Contains(code, "clear(s.M)") {
		t.Errorf("expected clear(s.M) in output")
	}
	if !strings.Contains(code, "s.Child.Reset()") {
		t.Errorf("expected recursive s.Child.Reset() in output")
	}
}
