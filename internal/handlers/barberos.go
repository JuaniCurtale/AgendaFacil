package handlers

import (
	"encoding/json"
	"net/http"

	db "agendaFacil/db/sqlc"

	"github.com/go-chi/chi/v5"
)

type BarberosHandler struct {
	Queries *db.Queries
}

func NewBarberosHandler(q *db.Queries) *BarberosHandler {
	return &BarberosHandler{Queries: q}
}

func (h *BarberosHandler) ListBarberos(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// 1️⃣ Buscar barbería
	barberia, err := h.Queries.GetBarberiaBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "barbería no encontrada", http.StatusNotFound)
		return
	}

	// 2️⃣ Obtener barberos activos
	barberos, err := h.Queries.ListBarberos(r.Context(), barberia.ID)
	if err != nil {
		http.Error(w, "error obteniendo barberos", http.StatusInternalServerError)
		return
	}

	// 3️⃣ Responder JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(barberos)
}
