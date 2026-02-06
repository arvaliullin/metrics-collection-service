package write

import "os"

// OSWriter записывает в файловую систему через os.WriteFile.
type OSWriter struct{}

// Write записывает data в path с режимом 0644.
func (OSWriter) Write(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
