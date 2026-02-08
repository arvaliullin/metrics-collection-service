package scan

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestDefaultScanner_Scan(t *testing.T) {
	dir := t.TempDir()
	goMod := `module testpkg

go 1.21
`
	withMarker := `package testpkg

// generate:reset
type Foo struct {
	N int
}
`
	noMarker := `package testpkg

type Bar struct {
	S string
}
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "foo.go"), []byte(withMarker), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bar.go"), []byte(noMarker), 0644); err != nil {
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
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		t.Fatalf("package had errors: %v", pkg.Errors)
	}

	var s DefaultScanner
	targets, err := s.Scan(pkg)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	tg := targets[0]
	if tg.TypeName.Obj().Name() != "Foo" {
		t.Errorf("expected type Foo, got %s", tg.TypeName.Obj().Name())
	}
	if tg.SourceBase != "foo" {
		t.Errorf("expected SourceBase foo, got %s", tg.SourceBase)
	}
}

func TestDefaultScanner_Scan_noMarker(t *testing.T) {
	dir := t.TempDir()
	goMod := `module testpkg
go 1.21
`
	noMarker := `package testpkg
type Bar struct { S string }
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bar.go"), []byte(noMarker), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &packages.Config{Dir: dir, Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedDeps}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	var s DefaultScanner
	targets, err := s.Scan(pkgs[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 0 {
		t.Fatalf("expected 0 targets (no marker), got %d", len(targets))
	}
}
