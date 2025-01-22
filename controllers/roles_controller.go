package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	models "main/Models"
	"net/http"

	"github.com/gorilla/mux"
)

func GetRoles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM roles WHERE deleted_at IS NULL")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		roles := []models.Role{}
		for rows.Next() {
			var role models.Role
			if err := rows.Scan(&role.ID, &role.RoleKey, &role.Description, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt); err != nil {
				log.Fatal(err)
			}
			roles = append(roles, role)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(roles)
	}
}

func GetRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var role models.Role
		err := db.QueryRow("SELECT * FROM roles WHERE id = $1 AND deleted_at IS NULL", id).
			Scan(&role.ID, &role.RoleKey, &role.Description, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(role)
	}
}

func CreateRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var role models.Role
		if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := db.QueryRow("INSERT INTO roles (role_key, description) VALUES ($1, $2) RETURNING id, created_at, updated_at", role.RoleKey, role.Description).
			Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(role)
	}
}

func UpdateRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var role models.Role
		if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE roles SET role_key = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND deleted_at IS NULL", role.RoleKey, role.Description, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(role)
	}
}

func DeleteRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
