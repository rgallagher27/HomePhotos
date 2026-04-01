package rest

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func httpStatusToCode(status int) string {
	return strings.ReplaceAll(strings.ToUpper(http.StatusText(status)), " ", "_")
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorBody{
			Code:    httpStatusToCode(statusCode),
			Message: message,
		},
	})
}
