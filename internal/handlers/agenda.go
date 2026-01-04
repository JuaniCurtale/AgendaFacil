package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	db "agendaFacil/db/sqlc"
)

func (h *BarberiaHandler) GetAgendaPublic(w http.ResponseWriter, r *http.Request) {
	log.Println("Entro handler a agenda")
	ctx := r.Context()

	slug := chi.URLParam(r, "slug")
	log.Println("slug:", slug)
	fechaStr := r.URL.Query().Get("fecha")
	log.Println("fechaStr:", fechaStr)

	if fechaStr == "" {
		http.Error(w, "fecha requerida (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	fecha, err := time.Parse("2006-01-02", fechaStr)
	if err != nil {
		http.Error(w, "formato de fecha inválido", http.StatusBadRequest)
		return
	}

	// 1️⃣ Buscar barbería
	barberia, err := h.Queries.GetBarberiaBySlug(ctx, slug)
	if err != nil {
		http.Error(w, "barbería no encontrada", http.StatusNotFound)
		return
	}

	// 2️⃣ Buscar turnos (ESTO PUEDE DEVOLVER [])
	turnos, err := h.Queries.ListTurnosByFecha(ctx, db.ListTurnosByFechaParams{
		BarberiaID: barberia.ID,
		Fecha:      fecha,
	})
	if err != nil {
		http.Error(w, "error al obtener agenda", http.StatusInternalServerError)
		return
	}

	// 3️⃣ SIEMPRE responder 200
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(turnos)
}
