# Mejoras Recomendadas para AgendaFacil

## 🔧 Mejoras Técnicas

### 1. Middleware de Autenticación Faltante

**Problema actual:**
```go
// No hay validación de JWT en rutas protegidas
r.Get("/b/{slug}/disponibilidad", barberiaHandler.GetDisponibilidad)
```

**Solución recomendada:**
```go
// Crear middleware de autenticación
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Sin autorización", http.StatusUnauthorized)
            return
        }
        
        // Validar token JWT
        token, err := jwt.ParseWithClaims(authHeader[7:], &Claims{}, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })
        
        if err != nil || !token.Valid {
            http.Error(w, "Token inválido", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Aplicar a rutas protegidas
r.Route("/admin", func(r chi.Router) {
    r.Use(AuthMiddleware)
    r.Get("/dashboard", adminHandler.Dashboard)
})
```

---

### 2. JWT Secret en Entorno

**Problema actual:**
```go
var jwtKey = []byte("mi_clave_super_secreta_123") // Hardcoded
```

**Solución recomendada:**
```go
func init() {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatal("JWT_SECRET no está configurada")
    }
    jwtKey = []byte(secret)
}
```

---

### 3. Contexto de BD Mejorado

**Problema actual:**
```go
func (h *BarberiaHandler) GetBarberiaPublic(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background() // Sin timeout
    barberia, err := h.Queries.GetBarberiaBySlug(ctx, slug)
```

**Solución recomendada:**
```go
func (h *BarberiaHandler) GetBarberiaPublic(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    barberia, err := h.Queries.GetBarberiaBySlug(ctx, slug)
    if err == context.DeadlineExceeded {
        http.Error(w, "Timeout en BD", http.StatusGatewayTimeout)
        return
    }
```

---

### 4. Refactorizar GetDisponibilidad

**Problema actual:**
La función `GetDisponibilidad` es muy larga y hace demasiadas cosas.

**Solución recomendada:**
```go
func (h *BarberiaHandler) GetDisponibilidad(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    slug := chi.URLParam(r, "slug")
    
    params, err := h.parseDisponibilidadParams(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    barberia, err := h.Queries.GetBarberiaBySlug(ctx, slug)
    if err != nil {
        http.Error(w, "barberia no encontrada", http.StatusNotFound)
        return
    }
    
    servicio, err := h.Queries.GetServicioByID(ctx, params.ServicioID)
    if err != nil {
        http.Error(w, "servicio no encontrado", http.StatusNotFound)
        return
    }
    
    slots := h.calcularDisponibilidad(ctx, barberia, servicio)
    writeJSON(w, slots)
}

// Métodos helper privados
func (h *BarberiaHandler) parseDisponibilidadParams(r *http.Request) (*DisponibilidadParams, error) {
    // Validación separada
}

func (h *BarberiaHandler) calcularDisponibilidad(ctx context.Context, barberia db.Barberia, servicio db.Servicio) []Slot {
    // Lógica centralizada
}
```

---

### 5. Logging Estructurado

**Problema actual:**
```go
log.Println("Entro handler a agenda") // Logging básico
```

**Solución recomendada:**
```go
import "github.com/sirupsen/logrus"

var logger = logrus.New()

func init() {
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetLevel(logrus.InfoLevel)
}

func (h *BarberiaHandler) GetAgendaPublic(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, "slug")
    
    logger.WithFields(logrus.Fields{
        "slug": slug,
        "fecha": r.URL.Query().Get("fecha"),
    }).Info("Obteniendo agenda pública")
    
    // ... resto del código
}
```

---

### 6. Errores Más Específicos

**Problema actual:**
```go
if err != nil {
    http.Error(w, "error al obtener agenda", http.StatusInternalServerError)
}
```

**Solución recomendada:**
```go
func handleQueryError(w http.ResponseWriter, err error, resource string) {
    if err == sql.ErrNoRows {
        http.Error(w, fmt.Sprintf("%s no encontrado", resource), http.StatusNotFound)
        return
    }
    
    if err == context.DeadlineExceeded {
        http.Error(w, "Timeout", http.StatusGatewayTimeout)
        return
    }
    
    logger.WithError(err).Error("Error de BD")
    http.Error(w, "Error interno", http.StatusInternalServerError)
}
```

---

### 7. Validación de Entrada Mejorada

**Problema actual:**
```go
// Validación mínima
if fechaStr == "" || servicioIDStr == "" {
    http.Error(w, "faltan parametros", http.StatusBadRequest)
}
```

**Solución recomendada:**
```go
type DisponibilidadParams struct {
    Fecha      time.Time `query:"fecha" validate:"required"`
    ServicioID int32     `query:"servicio_id" validate:"required,min=1"`
}

func (h *BarberiaHandler) GetDisponibilidad(w http.ResponseWriter, r *http.Request) {
    var params DisponibilidadParams
    
    // Parsear y validar automáticamente
    if err := h.validateQuery(r, &params); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // ...
}
```

---

### 8. Tests de Integración con Docker

**Sugerencia: Crear `internal/handlers/handlers_integration_test.go`**

```go
package handlers

import (
    "testing"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestWithRealDatabase(t *testing.T) {
    // Usar testcontainers para PostgreSQL
    container, err := testcontainers.GenericContainer(context.Background(), ...)
    if err != nil {
        t.Fatal(err)
    }
    
    // Conectar y ejecutar tests reales
    dbConn, _ := sql.Open("postgres", dsn)
    queries := db.New(dbConn)
    handler := NewAuthHandler(queries)
    
    // Tests con BD real
}
```

---

### 9. Rate Limiting

**Solución recomendada:**
```go
import "github.com/go-chi/chi/v5/middleware"

r.Use(middleware.ThrottleBackoff)

// O usar un rate limiter más sofisticado
import "golang.org/x/time/rate"

func rateLimitMiddleware(limiter *rate.Limiter) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

### 10. Documentación de API (OpenAPI/Swagger)

**Solución recomendada:**
```bash
# Instalar swagger
go get -u github.com/swaggo/swag/cmd/swag

# Agregar comentarios a handlers
// @Router /b/{slug} [get]
// @Param slug path string true "Slug de la barbería"
// @Success 200 {object} db.Barberia
// @Failure 404 {string} string "No encontrada"
func (h *BarberiaHandler) GetBarberiaPublic(w http.ResponseWriter, r *http.Request) {
    // ...
}

# Generar documentación
swag init -g ./cmd/server/main.go
```

---

## 📊 Checklist de Mejoras

- [ ] Implementar middleware de autenticación
- [ ] Mover JWT_SECRET a variables de entorno
- [ ] Agregar contextos con timeout
- [ ] Refactorizar GetDisponibilidad
- [ ] Implementar logging estructurado
- [ ] Mejorar manejo de errores
- [ ] Agregar validación con tags struct
- [ ] Crear tests de integración
- [ ] Implementar rate limiting
- [ ] Documentar API con Swagger

---

## 🚀 Prioridad Inmediata

1. **JWT_SECRET a .env** (5 minutos)
2. **Middleware de auth** (15 minutos)
3. **Tests de integración** (30 minutos)
4. **Logging estructurado** (20 minutos)

---

*Sugerencias de mejora - Basadas en análisis de seguridad y mejores prácticas Go*
