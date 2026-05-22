package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"loggraph/internal/config"
)

type AttachmentHandler struct {
	cfg *config.Config
}

func NewAttachmentHandler(cfg *config.Config) *AttachmentHandler {
	return &AttachmentHandler{cfg: cfg}
}

func (h *AttachmentHandler) Presign(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if input.Filename == "" {
		writeErr(w, http.StatusBadRequest, "filename required")
		return
	}
	if input.ContentType == "" {
		input.ContentType = "application/octet-stream"
	}

	// TODO: integrate with S3-compatible storage for real presigned URLs.
	// For now, return a stub that shows the intended shape.
	uploadURL := fmt.Sprintf("%s/%s/%s?presigned=true",
		h.cfg.S3.Endpoint, h.cfg.S3.Bucket, input.Filename)
	publicURL := fmt.Sprintf("%s/%s/%s",
		h.cfg.S3.Endpoint, h.cfg.S3.Bucket, input.Filename)

	writeJSON(w, http.StatusOK, map[string]string{
		"upload_url": uploadURL,
		"public_url": publicURL,
	})
}
