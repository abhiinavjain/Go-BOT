package handler

import (
	"customer-support-bot/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Dependency Injection

type ChatRequest struct {
	Prompt string `json:"prompt"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

type ChatHandler struct {
	service *service.GeminiService
}

type SessionResponse struct {
	SessionId string `json:"Session_id"`
}

func (h *ChatHandler) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	sessionId, err := h.service.CreateNewSession(r.Context())

	if err != nil {
		http.Error(w, "Could not create session", http.StatusInternalServerError)
		return
	}

	resp := SessionResponse{SessionId: sessionId}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

}

func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {

	sessionId := chi.URLParam(r, "sessionID")

	var req ChatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "your error message", http.StatusBadRequest)
		return
	}

	response, errr := h.service.RespGenerator(r.Context(), sessionId, req.Prompt)

	if errr != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	resp := ChatResponse{Response: response}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func NewChatHandler(ser *service.GeminiService) *ChatHandler {
	return &ChatHandler{service: ser}
}
