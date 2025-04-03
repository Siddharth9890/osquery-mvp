package api

import (
	"encoding/json"
	"net/http"

	"github.com/Siddharth9890/osquery-mvp/internal/database"
	"github.com/Siddharth9890/osquery-mvp/pkg/logger"
	"github.com/Siddharth9890/osquery-mvp/pkg/middleware"
	"go.uber.org/zap"
)

type Handler struct {
	dbService *database.Service
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewHandler(dbService *database.Service) *Handler {
	return &Handler{dbService: dbService}
}

func (h *Handler) GetLatestData(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestIDFromContext(r.Context())
	log := logger.WithRequestID(requestID)

	log.Info("Processing latest data request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr))

	if r.Method != http.MethodGet {
		log.Warn("Method not allowed",
			zap.String("method", r.Method))
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	info, err := h.dbService.GetLatestSystemInfo()
	if err != nil {
		log.Error("Failed to retrieve latest data",
			zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve latest data")
		return
	}

	log.Debug("Retrieved latest system info",
		zap.String("os_version", info.OSVersion),
		zap.String("osquery_version", info.OsqueryVersion),
		zap.Int("app_count", len(info.Apps)))

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    info,
	})

	log.Info("Successfully responded with latest data")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	logger.Log.Warn("Sending error response",
		zap.Int("status_code", code),
		zap.String("message", message))

	respondWithJSON(w, code, Response{
		Success: false,
		Error:   message,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Error("Error marshaling JSON response",
			zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
