package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	db "agendaFacil/db/sqlc"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Recuerda: En producción usa os.Getenv("JWT_SECRET")
var jwtKey = []byte("mi_clave_super_secreta_123")

type Credentials struct {
	Username string `json:"username"` // <-- Ahora pedimos Username
	Password string `json:"password"`
}

type Claims struct {
	UserID int32  `json:"user_id"`
	Rol    string `json:"rol"`
	jwt.RegisteredClaims
}

type AuthHandler struct {
	Queries *db.Queries
}

func NewAuthHandler(q *db.Queries) *AuthHandler {
	return &AuthHandler{Queries: q}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Decodificar JSON
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 2. Buscar usuario en DB (Usando la nueva query por Username)
	usuario, err := h.Queries.GetUsuarioByUsername(r.Context(), creds.Username)
	if err != nil {
		http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}

	// 3. Verificar Contraseña
	err = bcrypt.CompareHashAndPassword([]byte(usuario.PasswordHash), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}

	// 4. Generar Token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: usuario.ID,
		Rol:    usuario.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generando token", http.StatusInternalServerError)
		return
	}

	// 5. Responder
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"rol":   usuario.Rol,
	})
}
