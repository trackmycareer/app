package database

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	pwHash "github.com/trackmycareer/app/internal/password"
)

func SeedAdmin(ctx context.Context, pool *pgxpool.Pool, email, password string) {
	if email == "" || password == "" {
		return
	}

	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		log.Printf("checking admin existence: %v", err)
		return
	}
	if exists {
		return
	}

	hash, err := pwHash.Hash(password)
	if err != nil {
		log.Printf("hashing admin password: %v", err)
		return
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, provider, is_admin)
		VALUES ($1, $2, $3, $4, 'email', true)`,
		uuid.New(), email, hash, "Admin")
	if err != nil {
		log.Printf("seeding admin user: %v", err)
		return
	}
	log.Printf("admin user seeded: %s", email)
}
