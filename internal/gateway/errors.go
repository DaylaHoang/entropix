package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/DaylaHoang/entropix/pkg/api"
)

func writeError(w http.ResponseWriter, status int, message, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := api.ErrorResponse{
		Error: api.ErrorDetail{
			Message: message,
			Type:    errType,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func writeBadRequest(w http.ResponseWriter, message string) {
	writeError(w, http.StatusBadRequest, message, "invalid_request_error")
}

func writeInternalError(w http.ResponseWriter, message string) {
	writeError(w, http.StatusInternalServerError, message, "server_error")
}

func writeUpstreamError(w http.ResponseWriter, status int, message string) {
	writeError(w, status, message, "upstream_error")
}
