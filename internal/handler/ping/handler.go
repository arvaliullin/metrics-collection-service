package ping

import (
	"fmt"
	"net/http"
)

var (
	ErrInternalServer = fmt.Errorf("ping не выполняется")
)

type PingHandler struct {
	repository Pinger
}

func NewPingHandler(repository Pinger) *PingHandler {
	return &PingHandler{repository: repository}
}

// @Summary Проверка доступности базы данных
// @Description Проверяет соединение с базой данных
// @Tags health
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Internal Server Error"
// @Router /ping [get]
func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if err := h.repository.Ping(r.Context()); err != nil {
		http.Error(w, ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

}
