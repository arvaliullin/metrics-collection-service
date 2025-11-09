package html

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

// HTMLHandler обрабатывает GET-запрос и возвращает HTML-страницу со списком метрик.
type HTMLHandler struct {
	memStorage MetricStorage
}

// NewHTMLHandler создает обработчик HTML-страницы со списком метрик.
func NewHTMLHandler(memStorage MetricStorage) *HTMLHandler {
	return &HTMLHandler{memStorage: memStorage}
}

// ServeHTTP возвращает HTML-страницу с таблицами Gauge и Counter.
func (h *HTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gauges := h.memStorage.AllGauges(r.Context())
	counters := h.memStorage.AllCounters(r.Context())

	var b strings.Builder
	writeDocStart(&b)
	renderSection(&b, "Gauge", func(w io.StringWriter) {
		for _, m := range gauges {
			renderRow(w, m.ID, func(w io.StringWriter) {
				if m.Value != nil {
					w.WriteString(strconv.FormatFloat(*m.Value, 'f', -1, 64))
				}
			})
		}
	})
	renderSection(&b, "Counter", func(w io.StringWriter) {
		for _, m := range counters {
			renderRow(w, m.ID, func(w io.StringWriter) {
				if m.Value != nil {
					w.WriteString(strconv.FormatInt(int64(*m.Value), 10))
				}
			})
		}
	})
	writeDocEnd(&b)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b.String()))
}

func writeDocStart(w io.StringWriter) {
	w.WriteString("<!DOCTYPE html>")
	w.WriteString("<html><head><meta charset=\"utf-8\"><title>Metrics</title></head><body>")
}

func writeDocEnd(w io.StringWriter) {
	w.WriteString("</body></html>")
}

func renderSection(w io.StringWriter, title string, rows func(io.StringWriter)) {
	w.WriteString("<h1>")
	w.WriteString(title)
	w.WriteString("</h1>")
	w.WriteString("<table border=\"1\">")
	w.WriteString("<thead><tr><th>MType</th><th>Value</th></tr></thead><tbody>")
	rows(w)
	w.WriteString("</tbody></table>")
}

func renderRow(w io.StringWriter, mtype string, writeValue func(io.StringWriter)) {
	w.WriteString("<tr>")
	w.WriteString("<td>")
	w.WriteString(mtype)
	w.WriteString("</td>")
	w.WriteString("<td>")
	writeValue(w)
	w.WriteString("</td>")
	w.WriteString("</tr>")
}
