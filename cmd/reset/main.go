package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/arvaliullin/metrics-collection-service/cmd/reset/discover"
	"github.com/arvaliullin/metrics-collection-service/cmd/reset/gen"
	"github.com/arvaliullin/metrics-collection-service/cmd/reset/scan"
	"github.com/arvaliullin/metrics-collection-service/cmd/reset/write"

	"golang.org/x/tools/go/packages"
)

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
	n, errs := Run(root, discover.DefaultFinder{}, scan.DefaultScanner{}, gen.DefaultGenerator{}, write.OSWriter{})
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "reset: %v\n", e)
	}
	if len(errs) > 0 {
		os.Exit(1)
	}
	if n > 0 {
		fmt.Fprintf(os.Stdout, "reset: wrote %d file(s)\n", n)
	}
}

// Run выполняет обнаружение пакетов, сканирование, генерацию и запись. Возвращает число записанных файлов и ошибки.
// Используется из main и из интеграционных тестов.
func Run(root string, finder discover.MarkerFinder, scanner scan.Scanner, g gen.Generator, writer write.Writer) (int, []error) {
	patterns, err := finder.FindPackages(root)
	if err != nil {
		return 0, []error{err}
	}
	if len(patterns) == 0 {
		return 0, nil
	}

	cfg := &packages.Config{
		Dir:  root,
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return 0, []error{err}
	}
	packages.PrintErrors(pkgs)

	toProcess := buildWorkList(root, pkgs, scanner)

	var wrote atomic.Int32
	var errMu sync.Mutex
	var errs []error
	var wg sync.WaitGroup
	for _, job := range toProcess {
		wg.Add(1)
		go func(job pkgTargets) {
			defer wg.Done()
			n, err := processPackage(job.Pkg, job.Targets, g, writer)
			if err != nil {
				errMu.Lock()
				errs = append(errs, err)
				errMu.Unlock()
				return
			}
			wrote.Add(int32(n))
		}(job)
	}
	wg.Wait()

	return int(wrote.Load()), errs
}

type pkgTargets struct {
	Pkg     *packages.Package
	Targets []*scan.Target
}

func buildWorkList(root string, pkgs []*packages.Package, scanner scan.Scanner) []pkgTargets {
	var list []pkgTargets
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			continue
		}
		rel, err := filepath.Rel(root, pkg.Dir)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		targets, err := scanner.Scan(pkg)
		if err != nil || len(targets) == 0 {
			continue
		}
		list = append(list, pkgTargets{Pkg: pkg, Targets: targets})
	}
	return list
}

// processPackage группирует цели по файлу, генерирует и записывает .gen.go файлы.
func processPackage(pkg *packages.Package, targets []*scan.Target, g gen.Generator, w write.Writer) (int, error) {
	bySource := make(map[string][]*scan.Target)
	for _, t := range targets {
		base := t.SourceBase
		if base == "" {
			base = "reset"
		}
		bySource[base] = append(bySource[base], t)
	}
	var n int
	for baseName, group := range bySource {
		out, err := g.Generate(pkg, group)
		if err != nil {
			return 0, fmt.Errorf("generate %s: %w", pkg.PkgPath, err)
		}
		genName := baseName + gen.GenSuffix
		path := filepath.Join(pkg.Dir, genName)
		if err := w.Write(path, out); err != nil {
			return 0, fmt.Errorf("write %s: %w", path, err)
		}
		n++
	}
	return n, nil
}
