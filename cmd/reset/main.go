package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/tools/go/packages"
)

// genSuffix - суффикс сгенерированного файла (имя будет <исходный_файл>.gen.go).
const genSuffix = ".gen.go"

// Reset генерирует методы Reset() для структур с комментарием // generate:reset.
func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	root, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reset: %v\n", err)
		os.Exit(1)
	}

	patterns := findPackagesWithMarker(root)
	if len(patterns) == 0 {
		return
	}

	cfg := &packages.Config{
		Dir:  root,
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reset: load packages: %v\n", err)
		os.Exit(1)
	}
	packages.PrintErrors(pkgs)

	type work struct {
		pkg     *packages.Package
		targets []*targetStruct
	}
	var toProcess []work
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			continue
		}
		rel, err := filepath.Rel(root, pkg.Dir)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		targets := scanPackage(pkg)
		if len(targets) == 0 {
			continue
		}
		toProcess = append(toProcess, work{pkg: pkg, targets: targets})
	}

	var wrote atomic.Int32
	var wg sync.WaitGroup
	for _, w := range toProcess {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := processPackage(w.pkg, w.targets)
			wrote.Add(int32(n))
		}()
	}
	wg.Wait()

	if n := wrote.Load(); n > 0 {
		fmt.Fprintf(os.Stdout, "reset: wrote %d file(s)\n", n)
	}
}

// findPackagesWithMarker обходит дерево .go файлов и находит каталоги с generate:reset без вызова go list.
// Возвращает паттерны для packages.Load.
func findPackagesWithMarker(root string) []string {
	marker := []byte("generate:reset")
	seen := make(map[string]bool)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if !bytes.Contains(data, marker) {
			return nil
		}
		dir := filepath.Dir(path)
		rel, err := filepath.Rel(root, dir)
		if err != nil || strings.HasPrefix(rel, "..") {
			return nil
		}
		pattern := filepath.ToSlash(rel)
		if pattern != "." && !strings.HasPrefix(pattern, "./") {
			pattern = "./" + pattern
		}
		if !seen[pattern] {
			seen[pattern] = true
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "reset: scan dirs: %v\n", err)
		os.Exit(1)
	}
	patterns := make([]string, 0, len(seen))
	for p := range seen {
		patterns = append(patterns, p)
	}
	return patterns
}

// processPackage группирует цели по файлу, генерирует и записывает .gen.go файлы; возвращает их количество.
func processPackage(pkg *packages.Package, targets []*targetStruct) int {
	bySource := make(map[string][]*targetStruct)
	for _, t := range targets {
		base := t.SourceBase
		if base == "" {
			base = "reset"
		}
		bySource[base] = append(bySource[base], t)
	}
	var n int
	for baseName, group := range bySource {
		out, err := generateFile(pkg, group)
		if err != nil {
			fmt.Fprintf(os.Stderr, "reset: generate %s: %v\n", pkg.PkgPath, err)
			os.Exit(1)
		}
		genName := baseName + genSuffix
		path := filepath.Join(pkg.Dir, genName)
		if err := os.WriteFile(path, out, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "reset: write %s: %v\n", path, err)
			os.Exit(1)
		}
		n++
	}
	return n
}
