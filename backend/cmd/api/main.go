package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"playhack/internal/handler"
	"playhack/internal/middleware"
	"playhack/internal/repository"
	"playhack/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load, relying on environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Printf("Warning: Failed to ping database: %v", err)
	} else {
		log.Println("Successfully connected to PostgreSQL")
	}

	// Run migrations
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Printf("Migration init warning: %v", err)
	} else {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Printf("Migration error: %v", err)
		} else {
			log.Println("Database migrations applied successfully")
		}
	}

	// Wire application layers
	repo := repository.NewPostgresRepository(pool)
	authService := service.NewAuthService(repo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup Go 1.22+ method-based router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/request-otp", authHandler.HandleRequestOTP)
	mux.HandleFunc("POST /auth/verify-otp", authHandler.HandleVerifyOTP)
	mux.Handle("GET /me", authMiddleware.Authenticate(http.HandlerFunc(authHandler.HandleGetMe)))

	log.Printf("Server listening on port :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}
