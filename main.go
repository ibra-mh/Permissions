package main

import (
	"database/sql"
	"log"
	"main/app"
	"os"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS roles (
			id SERIAL PRIMARY KEY,
			role_key VARCHAR UNIQUE NOT NULL,
			description VARCHAR NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS user_roles (
			id SERIAL PRIMARY KEY,
			email VARCHAR NOT NULL,
			role_id INT NOT NULL REFERENCES roles(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	
	app.InitializeRoute(db)
}
