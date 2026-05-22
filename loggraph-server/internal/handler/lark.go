package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"loggraph/internal/model"
	"loggraph/internal/repository"
)

type LarkHandler struct {
	blockRepo *repository.BlockRepo
	tokenRepo *repository.WebhookTokenRepo
}

func NewLarkHandler(br *repository.BlockRepo, tr *repository.WebhookTokenRepo) *LarkHandler {
	return &LarkHandler{blockRepo: br, tokenRepo: tr}
}

// larkMsgContent is the inner JSON structure of a Lark text message.
type larkMsgContent struct {
	Text string `json:"text"`
}

// ServeHTTP handles Lark Event Subscription callbacks.
//
// Two modes:
//  1. URL verification — Lark POSTs {"challenge":"...","type":"url_verification"}
//     → respond with {"challenge":"..."}
//  2. Message receive — parse the nested JSON, extract text, create a Block.
//
// Security: validates the bearer token against webhook_tokens table.
func (h *LarkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "cannot read body")
		return
	}

	// Try to parse as a generic JSON object first
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}

	// ── URL verification ──
	if t, _ := raw["type"].(string); t == "url_verification" {
		challenge, _ := raw["challenge"].(string)
		writeJSON(w, http.StatusOK, map[string]string{"challenge": challenge})
		return
	}

	// ── Event message ──
	event, _ := raw["event"].(map[string]any)
	if event == nil {
		writeErr(w, http.StatusBadRequest, "missing event field")
		return
	}

	// Extract sender name
	senderName := "lark"
	if sender, ok := event["sender"].(map[string]any); ok {
		if name, ok := sender["sender_id"].(map[string]any); ok {
			if n, ok := name["user_id"].(string); ok {
				senderName = n
			}
		}
		// Lark v2 schema has sender_id as a string directly
		if sid, ok := sender["sender_id"].(string); ok {
			senderName = sid
		}
	}

	// Extract message content text
	msg, _ := event["message"].(map[string]any)
	if msg == nil {
		writeErr(w, http.StatusBadRequest, "missing message field")
		return
	}

	contentStr, _ := msg["content"].(string)
	if contentStr == "" {
		writeErr(w, http.StatusBadRequest, "empty message content")
		return
	}

	// Lark message content is a JSON string: {"text":"..."}
	var content larkMsgContent
	if err := json.Unmarshal([]byte(contentStr), &content); err != nil {
		// Fallback: use the raw string as content
		content.Text = contentStr
	}

	text := strings.TrimSpace(content.Text)
	if text == "" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	// If the user typed @bot-name, strip the leading @mention
	text = stripLeadingMention(text)

	block, err := h.blockRepo.Create(r.Context(), model.CreateBlockInput{
		UserID:  "lark:" + senderName,
		Content: text,
		Metadata: model.JSONMap{
			"source":     "lark",
			"sender_id":  senderName,
			"chat_id":    msg["chat_id"],
		},
	})
	if err != nil {
		log.Printf("[lark] create block failed: %v", err)
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[lark] block created: %s by %s", block.ID, senderName)
	writeJSON(w, http.StatusCreated, block)
}

// stripLeadingMention removes "@bot-name " prefixes common in group chats.
func stripLeadingMention(s string) string {
	if strings.HasPrefix(s, "@") {
		if idx := strings.IndexByte(s, ' '); idx > 0 {
			return strings.TrimSpace(s[idx+1:])
		}
	}
	return s
}
