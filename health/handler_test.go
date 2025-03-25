package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler(t *testing.T) {
	serverInfo = &ServerInfo{
		StartTime: time.Now().Add(-24 * time.Hour),
		Version:   "1.0.0-test",
	}

	req, err := http.NewRequest("GET", "/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler retornou código de status incorreto: recebido %v esperado %v",
			status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type incorreto: recebido %v esperado %v",
			contentType, "application/json")
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	message, ok := response["message"].(string)
	if !ok || message != "Serviço operacional" {
		t.Errorf("Mensagem incorreta: recebido %v esperado %v",
			message, "Serviço operacional")
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Dados não encontrados na resposta")
	}

	status, ok := data["status"].(string)
	if !ok || status != "UP" {
		t.Errorf("Status incorreto: recebido %v esperado %v",
			status, "UP")
	}

	version, ok := data["version"].(string)
	if !ok || version != "1.0.0-test" {
		t.Errorf("Versão incorreta: recebido %v esperado %v",
			version, "1.0.0-test")
	}

	_, ok = data["uptime"].(string)
	if !ok {
		t.Errorf("Uptime não encontrado na resposta")
	}
}
