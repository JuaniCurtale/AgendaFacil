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

type CreateServicioRequest struct {
	Nombre          string `json:"nombre"`
	DuracionMinutos int32  `json:"duracion_minutos"`
	Precio          string `json:"precio"` // Usamos string para Decimal/Numeric
}

func (h *ServiciosHandler) CreateServicio(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// 1. Validar Barbería
	barberia, err := h.Queries.GetBarberiaBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "Barbería no encontrada", http.StatusNotFound)
		return
	}

	// 2. Decodificar JSON
	var req CreateServicioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 3. Crear en DB
	nuevoServicio, err := h.Queries.CreateServicio(r.Context(), db.CreateServicioParams{
		BarberiaID:      barberia.ID,
		Nombre:          req.Nombre,
		DuracionMinutos: req.DuracionMinutos,
		Precio:          req.Precio,
	})

	if err != nil {
		http.Error(w, "Error creando servicio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nuevoServicio)
}
