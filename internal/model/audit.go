package models

// generate:reset
// AuditEvent представляет событие аудита.
type AuditEvent struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
