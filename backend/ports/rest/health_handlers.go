package rest

import (
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	status := HealthResponseStatus("ok")

	if err := s.db.PingContext(r.Context()); err != nil {
		status = HealthResponseStatus("degraded")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    status,
			Timestamp: time.Now().UTC(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
	})
}
