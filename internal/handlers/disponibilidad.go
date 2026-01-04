package handlers

import (
	db "agendaFacil/db/sqlc"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Slot struct {
	Inicio string `json:"inicio"`
	Fin    string `json:"fin"`
}

func (h *BarberiaHandler) GetDisponibilidad(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	fechaStr := r.URL.Query().Get("fecha")
	servicioIDStr := r.URL.Query().Get("servicio_id")

	if fechaStr == "" || servicioIDStr == "" {
		http.Error(w, "faltan parametros", http.StatusBadRequest)
		return
	}

	fecha, err := time.Parse("2006-01-02", fechaStr)
	if err != nil {
		http.Error(w, "fecha invalida", http.StatusBadRequest)
		return
	}

	servicioID, err := strconv.Atoi(servicioIDStr)
	if err != nil {
		http.Error(w, "servicio_id invalido", http.StatusBadRequest)
		return
	}

	barberia, err := h.Queries.GetBarberiaBySlug(ctx, slug)
	if err != nil {
		http.Error(w, "barberia no encontrada", http.StatusNotFound)
		return
	}

	servicio, err := h.Queries.GetServicioByID(ctx, int32(servicioID))
	if err != nil {
		http.Error(w, "servicio no encontrado", http.StatusNotFound)
		return
	}

	ocupados, err := h.Queries.ListTurnosOcupados(
		ctx,
		db.ListTurnosOcupadosParams{
			BarberiaID: barberia.ID,
			Fecha:      fecha,
		},
	)
	if err != nil {
		http.Error(w, "error obteniendo turnos", http.StatusInternalServerError)
		return
	}

	slots := calcularSlots(
		barberia.HoraApertura,
		barberia.HoraCierre,
		servicio.DuracionMinutos,
		ocupados,
	)

	writeJSON(w, slots)
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func calcularSlots(
	apertura time.Time,
	cierre time.Time,
	duracion int32,
	ocupados []db.ListTurnosOcupadosRow,
) []Slot {

	var disponibles []Slot
	slotDur := time.Duration(duracion) * time.Minute

	actual := apertura

	for actual.Add(slotDur).Before(cierre) || actual.Add(slotDur).Equal(cierre) {
		fin := actual.Add(slotDur)

		if !choca(actual, fin, ocupados) {
			disponibles = append(disponibles, Slot{
				Inicio: actual.Format("15:04"),
				Fin:    fin.Format("15:04"),
			})
		}

		actual = actual.Add(slotDur)
	}

	return disponibles
}
func choca(inicio, fin time.Time, ocupados []db.ListTurnosOcupadosRow) bool {
	for _, t := range ocupados {
		if inicio.Before(t.HoraFin) && fin.After(t.HoraInicio) {
			return true
		}
	}
	return false
}
