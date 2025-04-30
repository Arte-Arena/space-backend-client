package orders

import (
	"api/middlewares"
	"api/schemas"
	"api/utils"
	"encoding/json"
	"net/http"
)

func getAllOrders(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIDKey)
	if userId == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Usuário não autorizado",
		})
		return
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAllOrders(w, r)
	case http.MethodPatch:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
