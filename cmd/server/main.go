package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	db "agendaFacil/db/sqlc"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"agendaFacil/internal/handlers"
)

func main() {
	// Leer variables de entorno
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Setear defaults para desarrollo si no existen
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPass == "" {
		dbPass = "postgres"
	}
	if dbHost == "" {
		dbHost = "db" // nombre del contenedor en Docker
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
	barberiaHandler := handlers.NewBarberiaHandler(queries)
	serviciosHandler := handlers.NewServiciosHandler(queries)
	barberosHandler := handlers.NewBarberosHandler(queries)

	// Router
	r := chi.NewRouter()
	r.Get("/b/{slug}", barberiaHandler.GetBarberiaPublic)
	r.Get("/b/{slug}/agenda", barberiaHandler.GetAgendaPublic)
	r.Get("/b/{slug}/servicios", serviciosHandler.ListServiciosActivos)
	r.Get("/b/{slug}/barberos", barberosHandler.ListBarberos)
	r.Get("/b/{slug}/disponibilidad", barberiaHandler.GetDisponibilidad)
	r.Get("/b/{slug}/agenda", barberiaHandler.AgendaHTML)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API OK"))
	})

	// Puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default en desarrollo
	}

	log.Println("Servidor en puerto", port)
	http.ListenAndServe(":"+port, r)
}
