package health

import "time"

type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

type ServerInfo struct {
	StartTime time.Time
	Version   string
}

func NewServerInfo() *ServerInfo {
	return &ServerInfo{
		StartTime: time.Now(),
		Version:   "1.0.0",
	}
}
