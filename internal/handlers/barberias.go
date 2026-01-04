package handlers

import (
	db "agendaFacil/db/sqlc"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type BarberiaHandler struct {
	Queries *db.Queries
}

func NewBarberiaHandler(q *db.Queries) *BarberiaHandler {
	return &BarberiaHandler{Queries: q}
}

func (h *BarberiaHandler) GetBarberiaPublic(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	slug := chi.URLParam(r, "slug")

	barberia, err := h.Queries.GetBarberiaBySlug(ctx, slug)
	if err != nil {
		http.Error(w, "barbería no encontrada", http.StatusNotFound)
		return
	}

	servicios, err := h.Queries.ListServicios(ctx, barberia.ID)
	if err != nil {
		http.Error(w, "error servicios", http.StatusInternalServerError)
		return
	}

	barberos, err := h.Queries.ListBarberos(ctx, barberia.ID)
	if err != nil {
		http.Error(w, "error barberos", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"barberia":  barberia,
		"servicios": servicios,
		"barberos":  barberos,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
