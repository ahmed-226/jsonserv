package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"myserv/internal/config"
	"myserv/internal/query"
	"myserv/internal/store"
)

type Handler struct {
	store  *store.Store
	config *config.Config
}

func New(s *store.Store, cfg *config.Config) *Handler {
	return &Handler{store: s, config: cfg}
}

func (h *Handler) List(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items := h.store.List(entity)
		if items == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "entity not found"})
			return
		}

		ecfg := h.config.EntityConfig(entity)

		if ecfg.Filters.Enabled {
			items = query.Filter(items, r.URL.Query())
		}
		if ecfg.Sort.Enabled {
			items = query.SortItems(items, r.URL.Query(), ecfg.Sort.DefaultField, ecfg.Sort.DefaultOrder)
		}

		var total int
		if ecfg.Paginate.Enabled {
			page, limit := query.ParsePageLimit(r.URL.Query(), ecfg.Paginate.DefaultPage, ecfg.Paginate.DefaultLimit)
			items, total = query.Paginate(items, page, limit)
			w.Header().Set("X-Total-Count", strconv.Itoa(total))
		}

		writeJSON(w, http.StatusOK, items)
	}
}

func (h *Handler) Get(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		item, idx := h.store.Get(entity, id)
		if idx == -1 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

func (h *Handler) Create(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		item := h.store.Create(entity, body)
		writeJSON(w, http.StatusCreated, item)
	}
}

func (h *Handler) Update(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		_, idx := h.store.Get(entity, id)
		if idx == -1 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		item := h.store.Update(entity, idx, body)
		writeJSON(w, http.StatusOK, item)
	}
}

func (h *Handler) Patch(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		_, idx := h.store.Get(entity, id)
		if idx == -1 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		delete(body, "id")
		item := h.store.Patch(entity, idx, body)
		writeJSON(w, http.StatusOK, item)
	}
}

func (h *Handler) Delete(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		_, idx := h.store.Get(entity, id)
		if idx == -1 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		h.store.Delete(entity, idx)
		w.WriteHeader(http.StatusNoContent)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
