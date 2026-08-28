package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type FacilitySeed struct {
	Name      string
	SportType string
	Courts    []string
}

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/playhack_db?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	log.Println("Starting database seeding...")

	facilities := []FacilitySeed{
		{
			Name:      "Indoor Badminton Complex",
			SportType: "Badminton",
			Courts:    []string{"Court 1", "Court 2"},
		},
		{
			Name:      "IITG Tennis Arena",
			SportType: "Tennis",
			Courts:    []string{"Court A", "Court B"},
		},
		{
			Name:      "Main Football Turf",
			SportType: "Football",
			Courts:    []string{"Ground 1", "Ground 2"},
		},
	}

	now := time.Now().Local()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dates := []time.Time{today, today.AddDate(0, 0, 1)}

	for _, f := range facilities {
		var facilityID string
		err := pool.QueryRow(ctx, `
			INSERT INTO facilities (name, sport_type)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
			RETURNING id;
		`, f.Name, f.SportType).Scan(&facilityID)

		if err != nil {
			// If already exists, fetch ID
			_ = pool.QueryRow(ctx, `SELECT id FROM facilities WHERE name = $1`, f.Name).Scan(&facilityID)
		}

		if facilityID == "" {
			continue
		}

		log.Printf("Seeded Facility: %s (%s) [ID: %s]", f.Name, f.SportType, facilityID)

		for _, courtLabel := range f.Courts {
			var courtID string
			_ = pool.QueryRow(ctx, `
				INSERT INTO courts (facility_id, label)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
				RETURNING id;
			`, facilityID, courtLabel).Scan(&courtID)

			if courtID == "" {
				_ = pool.QueryRow(ctx, `SELECT id FROM courts WHERE facility_id = $1 AND label = $2`, facilityID, courtLabel).Scan(&courtID)
			}

			if courtID == "" {
				continue
			}

			log.Printf("  - Seeded Court: %s [ID: %s]", courtLabel, courtID)

			// Generate hourly slots from 08:00 to 21:00 for today and tomorrow
			for _, date := range dates {
				for hour := 8; hour < 21; hour++ {
					startTime := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, date.Location())
					endTime := startTime.Add(1 * time.Hour)

					_, err := pool.Exec(ctx, `
						INSERT INTO slots (court_id, start_time, end_time)
						VALUES ($1, $2, $3)
						ON CONFLICT (court_id, start_time) DO NOTHING;
					`, courtID, startTime, endTime)

					if err != nil {
						log.Printf("Error inserting slot for court %s at %v: %v", courtID, startTime, err)
					}
				}
			}
		}
	}

	log.Println("Database seeding completed successfully!")
}
