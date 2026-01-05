package handlers

import (
	db "agendaFacil/db/sqlc"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"time"

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

func (h *BarberiaHandler) AgendaHTML(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	fechaStr := r.URL.Query().Get("fecha")
	if fechaStr == "" {
		http.Error(w, "falta fecha", http.StatusBadRequest)
		return
	}

	fecha, err := time.Parse("2006-01-02", fechaStr)
	if err != nil {
		http.Error(w, "fecha invalida", http.StatusBadRequest)
		return
	}

	barberia, err := h.Queries.GetBarberiaBySlug(ctx, slug)
	if err != nil {
		http.Error(w, "barberia no encontrada", http.StatusNotFound)
		return
	}

	turnos, err := h.Queries.ListTurnosOcupados(
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

	tmpl := template.Must(template.New("agenda").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Agenda {{.Barberia}}</title>
</head>
<body>
	<h1>Agenda de {{.Barberia}}</h1>
	<p>Fecha: {{.Fecha}}</p>

	<table border="1" cellpadding="5">
		<tr>
			<th>Barbero</th>
			<th>Inicio</th>
			<th>Fin</th>
		</tr>
		{{range .Turnos}}
		<tr>
			<td>{{.BarberoID}}</td>
			<td>{{.HoraInicio.Format "15:04"}}</td>
			<td>{{.HoraFin.Format "15:04"}}</td>
		</tr>
		{{else}}
		<tr>
			<td colspan="3">Sin turnos</td>
		</tr>
		{{end}}
	</table>
</body>
</html>
	`))

	data := map[string]any{
		"Barberia": barberia.Nombre,
		"Fecha":    fecha.Format("2006-01-02"),
		"Turnos":   turnos,
	}

	tmpl.Execute(w, data)
}

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
