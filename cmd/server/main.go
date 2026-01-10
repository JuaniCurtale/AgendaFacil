package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath" // Agregado para rutas de archivos
	"strings"       // Agregado para manipulación de rutas

	db "agendaFacil/db/sqlc"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"agendaFacil/internal/handlers"
)

func main() {
	// Leer variables de entorno
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD") // CORREGIDO: Coincide con docker-compose
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Setear defaults para desarrollo
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPass == "" {
		dbPass = "postgres"
	}
	if dbHost == "" {
		dbHost = "db"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbName == "" {
		dbName = "barberia"
	}

	// Construir DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	// Conectar a la DB
	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error conectando a la DB:", err)
	}

	if err := dbConn.Ping(); err != nil {
		log.Fatal("No se pudo conectar a la DB:", err)
	}

	log.Println("Conectado a la DB correctamente")

	// Inicializar queries y handlers
	queries := db.New(dbConn)

	authHandler := handlers.NewAuthHandler(queries)
	barberiaHandler := handlers.NewBarberiaHandler(queries)
	serviciosHandler := handlers.NewServiciosHandler(queries)
	barberosHandler := handlers.NewBarberosHandler(queries)

	// Router
	r := chi.NewRouter()

	// --- RUTAS API ---
	r.Post("/login", authHandler.Login) // <--- NUEVA RUTA
	r.Get("/b/{slug}", barberiaHandler.GetBarberiaPublic)
	r.Get("/b/{slug}/agenda", barberiaHandler.GetAgendaPublic)
	r.Get("/b/{slug}/servicios", serviciosHandler.ListServiciosActivos)
	r.Get("/b/{slug}/barberos", barberosHandler.ListBarberos)
	r.Get("/b/{slug}/disponibilidad", barberiaHandler.GetDisponibilidad)
	r.Post("/b/{slug}/reservar", barberiaHandler.PostReservar)

	r.Group(func(r chi.Router) {
		// Aquí usamos el middleware que acabamos de crear en auth.go
		r.Use(handlers.AuthMiddleware)

		// Rutas protegidas
		r.Post("/b/{slug}/servicios", serviciosHandler.CreateServicio)
		r.Post("/b/{slug}/barberos", barberosHandler.CreateBarbero)
	})

	// --- ARCHIVOS ESTÁTICOS (CORREGIDO PARA CHI) ---
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))

	// Servir archivos HTML específicos
	r.Get("/handlers.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "web", "HANDLERS_VISUALIZER.html"))
	})

	// Admin dashboard demo
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "web", "admin_dashboard.html"))
	})

	// AGREGA ESTA LÍNEA (Maneja la raíz "/")
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "web", "index.html"))
	})
	r.Get("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "web", "index.html"))
	})

	// FileServer para otros archivos estáticos
	FileServer(r, "/", filesDir)

	// Puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Servidor en puerto", port)
	http.ListenAndServe(":"+port, r)
}

// FileServer configura convenientemente un manejador de servidor de archivos dentro de Chi
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
