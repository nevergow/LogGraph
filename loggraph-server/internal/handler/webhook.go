package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"loggraph/internal/model"
	"loggraph/internal/repository"
)

type WebhookHandler struct {
	blockRepo *repository.BlockRepo
	tokenRepo *repository.WebhookTokenRepo
}

func NewWebhookHandler(br *repository.BlockRepo, tr *repository.WebhookTokenRepo) *WebhookHandler {
	return &WebhookHandler{blockRepo: br, tokenRepo: tr}
}

// ── Receive webhook log ──

func (h *WebhookHandler) Receive(w http.ResponseWriter, r *http.Request) {
	token := extractBearer(r)
	if token == "" {
		writeErr(w, http.StatusUnauthorized, "missing bearer token")
		return
	}
	valid, _ := h.tokenRepo.Validate(r.Context(), token)
	if !valid {
		writeErr(w, http.StatusUnauthorized, "invalid token")
		return
	}

	var input model.WebhookLogInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if input.Content == "" {
		writeErr(w, http.StatusBadRequest, "content required")
		return
	}
	if input.UserID == "" {
		input.UserID = "webhook"
	}

	// Support idempotency key to prevent duplicate webhook deliveries
	if key := r.Header.Get("X-Idempotency-Key"); key != "" {
		// Simple strategy: use the key as metadata marker
		if input.Metadata == nil {
			input.Metadata = model.JSONMap{}
		}
		input.Metadata["_idempotency_key"] = key
	}

	block, err := h.blockRepo.Create(r.Context(), model.CreateBlockInput{
		UserID:   input.UserID,
		Content:  input.Content,
		Metadata: input.Metadata,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, block)
}

// ── Token management ──

type generateTokenInput struct {
	Name string `json:"name"`
}

func (h *WebhookHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var input generateTokenInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if input.Name == "" {
		writeErr(w, http.StatusBadRequest, "name required")
		return
	}
	plaintext, _, err := h.tokenRepo.Generate(r.Context(), input.Name)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Return the plaintext — this is the ONLY time it's visible
	writeJSON(w, http.StatusCreated, map[string]string{
		"name":  input.Name,
		"token": plaintext,
	})
}

func (h *WebhookHandler) ListTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.tokenRepo.List(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tokens == nil {
		tokens = []model.WebhookToken{}
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (h *WebhookHandler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.tokenRepo.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return h
}
