package handler

import (
	"net/http"
	"strconv"

	"loggraph/internal/model"
	"loggraph/internal/repository"
)

type NodeHandler struct {
	repo *repository.NodeRepo
}

func NewNodeHandler(repo *repository.NodeRepo) *NodeHandler {
	return &NodeHandler{repo: repo}
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
