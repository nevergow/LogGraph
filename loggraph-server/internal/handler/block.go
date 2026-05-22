package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"loggraph/internal/model"
	"loggraph/internal/repository"
)

type BlockHandler struct {
	blockRepo    *repository.BlockRepo
	relationRepo *repository.RelationRepo
}

func NewBlockHandler(br *repository.BlockRepo, rr *repository.RelationRepo) *BlockHandler {
	return &BlockHandler{blockRepo: br, relationRepo: rr}
}

func (h *BlockHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.CreateBlockInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if input.Content == "" {
		writeErr(w, http.StatusBadRequest, "content required")
		return
	}
	if input.UserID == "" {
		input.UserID = "system"
	}

	block, err := h.blockRepo.Create(r.Context(), input)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, block)
}

func (h *BlockHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	block, err := h.blockRepo.GetByID(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "block not found")
		return
	}
	writeJSON(w, http.StatusOK, block)
}

func (h *BlockHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.BlockListQuery{}
	if v := r.URL.Query().Get("cursor"); v != "" {
		q.Cursor = &v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		q.Limit = v
	}
	if v := r.URL.Query().Get("project"); v != "" {
		q.Project = &v
	}
	if v := r.URL.Query().Get("person"); v != "" {
		q.Person = &v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		q.Status = &v
	}
	if v := r.URL.Query().Get("q"); v != "" {
		q.Query = &v
	}
	if v := r.URL.Query().Get("since"); v != "" {
		q.Since = &v
	}
	if v := r.URL.Query().Get("until"); v != "" {
		q.Until = &v
	}

	page, err := h.blockRepo.List(r.Context(), q)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if page.Data == nil {
		page.Data = []model.Block{}
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *BlockHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input model.UpdateBlockInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	block, err := h.blockRepo.Update(r.Context(), id, input)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, block)
}

func (h *BlockHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.blockRepo.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BlockHandler) Graph(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	graph, err := h.relationRepo.GraphForBlock(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "block not found")
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

func (h *BlockHandler) NodeGraph(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	graph, err := h.relationRepo.GraphForNode(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "node not found")
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

// ── helpers ──

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
