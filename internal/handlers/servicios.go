package handlers

import (
	"encoding/json"
	"net/http"

	db "agendaFacil/db/sqlc"

	"github.com/go-chi/chi/v5"
)

type ServiciosHandler struct {
	Queries *db.Queries
}

func NewServiciosHandler(q *db.Queries) *ServiciosHandler {
	return &ServiciosHandler{Queries: q}
}

func (h *ServiciosHandler) ListServiciosActivos(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	//Buscar barbería
	barberia, err := h.Queries.GetBarberiaBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "barbería no encontrada", http.StatusNotFound)
		return
	}

	// Obtener servicios activos
	servicios, err := h.Queries.ListServicios(r.Context(), barberia.ID)
	if err != nil {
		http.Error(w, "error obteniendo servicios", http.StatusInternalServerError)
		return
	}

	// Responder JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servicios)
}
