package discover

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/arvaliullin/metrics-collection-service/cmd/reset/scan"
)

// MarkerFinder находит директории с маркером generate:reset.
type MarkerFinder interface {
	FindPackages(root string) ([]string, error)
}

// DefaultFinder обходит дерево .go файлов и возвращает паттерны для packages.Load.
type DefaultFinder struct{}

// FindPackages обходит дерево и находит каталоги с generate:reset без вызова go list.
// Для каждого файла читает порциями и прекращает чтение после первого вхождения маркера.
func (DefaultFinder) FindPackages(root string) ([]string, error) {
	marker := []byte(scan.GenerateResetMarker)
	seen := make(map[string]bool)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if !fileContainsMarker(path, marker) {
			return nil
		}
		dir := filepath.Dir(path)
		if pattern, ok := dirToPattern(root, dir); ok {
			seen[pattern] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	patterns := make([]string, 0, len(seen))
	for p := range seen {
		patterns = append(patterns, p)
	}
	return patterns, nil
}

func dirToPattern(root, dir string) (pattern string, ok bool) {
	rel, err := filepath.Rel(root, dir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", false
	}
	pattern = filepath.ToSlash(rel)
	if pattern != "." && !strings.HasPrefix(pattern, "./") {
		pattern = "./" + pattern
	}
	return pattern, true
}

// fileContainsMarker читает файл порциями и возвращает true при первом вхождении marker.
func fileContainsMarker(path string, marker []byte) bool {
	if len(marker) == 0 {
		return true
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	const chunk = 4096
	overlap := len(marker) - 1
	buf := make([]byte, chunk+overlap)
	start := 0
	br := bufio.NewReader(f)
	for {
		n, err := br.Read(buf[start : start+chunk])
		if n > 0 {
			end := start + n
			if bytes.Contains(buf[:end], marker) {
				return true
			}
			if end > overlap {
				copy(buf, buf[end-overlap:end])
				start = overlap
			} else {
				start = end
			}
		}
		if err != nil {
			break
		}
	}
	return false
}
