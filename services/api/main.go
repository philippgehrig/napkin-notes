package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
	"github.com/philippgehrig/napkin-notes/services/api/internal/database"
	"github.com/philippgehrig/napkin-notes/services/api/internal/repository"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func runMigrations() {
	dsn := database.BuildDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("migration: failed to connect to database: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("migration: failed to create driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("migration: failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration: failed to run migrations: %v", err)
	}

	log.Println("migrations applied successfully")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("RUN_MIGRATIONS") == "true" {
		runMigrations()
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// Health endpoint
	r.Get("/health", healthHandler)

	// Auth routes - only wire if DB is configured
	if os.Getenv("DB_HOST") != "" {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "dev-secret-change-in-production"
		}

		dsn := database.BuildDSN(
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_SSLMODE"),
		)

		db, err := database.Connect(dsn)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer db.Close()

		jwtSvc := auth.NewJWTService(jwtSecret)
		userRepo := repository.NewPostgresUserRepo(db)
		authSvc := auth.NewAuthService(userRepo, jwtSvc)
		authHandler := auth.NewAuthHandler(authSvc)

		r.Route("/api/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
		})
	}

	log.Printf("API server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
