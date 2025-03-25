package health

import (
	"api/utils"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	serverInfo *ServerInfo
	once       sync.Once
)

func GetServerInfo() *ServerInfo {
	once.Do(func() {
		serverInfo = NewServerInfo()
	})
	return serverInfo
}

func Handler(w http.ResponseWriter, r *http.Request) {
	info := GetServerInfo()

	uptime := time.Since(info.StartTime).String()

	healthStatus := HealthStatus{
		Status:    "UP",
		Timestamp: time.Now(),
		Version:   info.Version,
		Uptime:    uptime,
	}

	response := utils.ApiResponse{
		Message: "Operational service",
		Data:    healthStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erro ao gerar resposta", http.StatusInternalServerError)
		return
	}
}
