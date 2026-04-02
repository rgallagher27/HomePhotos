package rest

import (
	"encoding/json"
	"net/http"
)

func (s *Server) PostScannerRun(w http.ResponseWriter, r *http.Request) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	// Check if already running
	status := s.scanner.Status()
	if status.State == "scanning" {
		writeError(w, http.StatusConflict, "scan already in progress")
		return
	}

	// Start scan in background
	go func() {
		_ = s.scanner.Run(r.Context())
	}()

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(ScannerRunResponse{
		Status: "started",
	})
}

func (s *Server) GetScannerStatus(w http.ResponseWriter, r *http.Request) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	status := s.scanner.Status()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ScannerStatusResponse{
		Status:     status.State,
		TotalFiles: status.TotalFiles,
		Processed:  status.Processed,
		Errors:     status.Errors,
		StartedAt:  status.StartedAt,
	})
}
