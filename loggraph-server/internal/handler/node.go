package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"loggraph/internal/model"
	"loggraph/internal/repository"
)

type NodeHandler struct {
	repo      *repository.NodeRepo
	blockRepo *repository.BlockRepo
}

func NewNodeHandler(repo *repository.NodeRepo, blockRepo *repository.BlockRepo) *NodeHandler {
	return &NodeHandler{repo: repo, blockRepo: blockRepo}
}

func (h *NodeHandler) List(w http.ResponseWriter, r *http.Request) {
	var nodeType *string
	if v := r.URL.Query().Get("type"); v != "" {
		nodeType = &v
	}
	nodes, err := h.repo.List(r.Context(), nodeType)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if nodes == nil {
		nodes = []model.Node{}
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (h *NodeHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []model.Node{})
		return
	}
	limit := 8
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 {
		limit = v
	}
	var nodeType *string
	if v := r.URL.Query().Get("type"); v != "" {
		nodeType = &v
	}

	nodes, err := h.repo.Suggest(r.Context(), q, nodeType, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if nodes == nil {
		nodes = []model.Node{}
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (h *NodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string          `json:"name"`
		Type model.NodeType `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if input.Name == "" {
		writeErr(w, http.StatusBadRequest, "name required")
		return
	}
	if input.Type == "" {
		input.Type = model.NodeTypeProject
	}
	node, err := h.repo.Create(r.Context(), input.Name, input.Type)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, node)
}

func (h *NodeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Name    string `json:"name"`
		Cascade bool   `json:"cascade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if input.Name == "" {
		writeErr(w, http.StatusBadRequest, "name required")
		return
	}

	// Get old name before updating
	old, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "node not found")
		return
	}

	node, err := h.repo.Update(r.Context(), id, input.Name)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Cascade: replace #oldName with #newName in all blocks
	if input.Cascade && old.Name != "" && old.Name != input.Name {
		_, err := h.blockRepo.ReplaceProjectInContent(r.Context(), old.Name, input.Name)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	writeJSON(w, http.StatusOK, node)
}

func (h *NodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
