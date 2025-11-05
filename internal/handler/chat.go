package handler

import (
	"customer-support-bot/internal/service"
	"encoding/json"
	"net/http"
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

func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "your error message", http.StatusBadRequest)
		return
	}

	response, errr := h.service.RespGenerator(r.Context(), req.Prompt)

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
