package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	db "agendaFacil/db/sqlc"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
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

type CreateBarberoRequest struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *BarberosHandler) CreateBarbero(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// 1. Buscar Barbería
	barberia, err := h.Queries.GetBarberiaBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "Barbería no encontrada", http.StatusNotFound)
		return
	}

	// 2. Decodificar JSON
	var req CreateBarberoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("CreateBarbero: error decoding JSON:", err)
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Encriptar contraseña (Nunca guardar texto plano)
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error procesando contraseña", http.StatusInternalServerError)
		return
	}

	// 4. Crear usuario en DB
	nuevoBarbero, err := h.Queries.CreateUsuario(r.Context(), db.CreateUsuarioParams{
		BarberiaID:   barberia.ID,
		Nombre:       req.Nombre,
		Apellido:     req.Apellido,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPwd),
		Rol:          "barbero", // Forzamos el rol para que no creen otro admin
	})

	if err != nil {
		// Tip: Si el error es por "username duplicado", aquí podrías manejarlo mejor
		http.Error(w, "Error guardando barbero: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nuevoBarbero)
}
