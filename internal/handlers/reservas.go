package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	db "agendaFacil/db/sqlc"

	"github.com/go-chi/chi/v5"
)

// Estructura para recibir los datos del JSON
type CreateReservaRequest struct {
	ServicioID      int32  `json:"servicio_id"`
	BarberoID       int32  `json:"barbero_id"`  // Puede ser 0 si es "cualquiera"
	Fecha           string `json:"fecha"`       // YYYY-MM-DD
	HoraInicio      string `json:"hora_inicio"` // HH:MM
	ClienteNombre   string `json:"cliente_nombre"`
	ClienteTelefono string `json:"cliente_telefono"`
}

func (h *BarberiaHandler) PostReservar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	// 1. Decodificar el body
	var req CreateReservaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 2. Validar Fechas y Horas
	fecha, err := time.Parse("2006-01-02", req.Fecha)
	if err != nil {
		http.Error(w, "Formato de fecha incorrecto", http.StatusBadRequest)
		return
	}

	horaInicio, err := time.Parse("15:04", req.HoraInicio)
	if err != nil {
		http.Error(w, "Formato de hora incorrecto", http.StatusBadRequest)
		return
	}

	// 3. Buscar Barbería y Servicio (para saber duración)
	barberia, err := h.Queries.GetBarberiaBySlug(ctx, slug)
	if err != nil {
		http.Error(w, "Barbería no encontrada", http.StatusNotFound)
		return
	}

	servicio, err := h.Queries.GetServicioByID(ctx, req.ServicioID)
	if err != nil {
		http.Error(w, "Servicio no encontrado", http.StatusNotFound)
		return
	}

	// Calcular Hora Fin
	horaFin := horaInicio.Add(time.Duration(servicio.DuracionMinutos) * time.Minute)

	// 4. Validar que no haya overlap (superposición)
	// Usamos la query que ya tenés en SQLC: HasTurnoOverlap
	overlap, err := h.Queries.HasTurnoOverlap(ctx, db.HasTurnoOverlapParams{
		BarberiaID: barberia.ID,
		BarberoID:  req.BarberoID, // Ojo: asegúrate de enviar el barbero correcto
		Fecha:      fecha,
		HoraInicio: horaInicio,
		HoraFin:    horaFin,
	})

	if err != nil {
		http.Error(w, "Error verificando disponibilidad", http.StatusInternalServerError)
		return
	}

	if overlap {
		http.Error(w, "El turno seleccionado ya no está disponible", http.StatusConflict) // 409 Conflict
		return
	}

	// 5. Guardar en DB
	turno, err := h.Queries.CreateTurno(ctx, db.CreateTurnoParams{
		BarberiaID:      barberia.ID,
		BarberoID:       req.BarberoID, // Ojo: Si es 0 (cualquiera), tu DB podría fallar si la FK es not null.
		ServicioID:      req.ServicioID,
		Fecha:           fecha,
		HoraInicio:      horaInicio, // SQLC maneja time.Time para columnas TIME
		HoraFin:         horaFin,
		ClienteNombre:   req.ClienteNombre,
		ClienteTelefono: toNullString(req.ClienteTelefono), // Helper para convertir string a sql.NullString
		Estado:          toNullString("pendiente"),
	})

	if err != nil {
		http.Error(w, "Error al guardar reserva: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(turno)
}

// Helper simple para SQLC
func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// Nota: Deberás ajustar los imports arriba para incluir database/sql si usas tipos Null
