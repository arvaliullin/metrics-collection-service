package write

// Writer записывает сгенерированные данные в файл.
type Writer interface {
	Write(path string, data []byte) error
}
