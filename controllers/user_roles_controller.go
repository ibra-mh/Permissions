package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	models "main/Models"
	"net/http"

	"github.com/gorilla/mux"
)

func GetUserRoles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM user_roles WHERE deleted_at IS NULL")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		userRoles := []models.UserRole{}
		for rows.Next() {
			var userRole models.UserRole
			if err := rows.Scan(&userRole.ID, &userRole.Email, &userRole.RoleID, &userRole.CreatedAt, &userRole.UpdatedAt, &userRole.DeletedAt); err != nil {
				log.Fatal(err)
			}
			userRoles = append(userRoles, userRole)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(userRoles)
	}
}

func GetUserRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var userRole models.UserRole
		err := db.QueryRow("SELECT * FROM user_roles WHERE id = $1 AND deleted_at IS NULL", id).
			Scan(&userRole.ID, &userRole.Email, &userRole.RoleID, &userRole.CreatedAt, &userRole.UpdatedAt, &userRole.DeletedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(userRole)
	}
}
func CreateUserRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userRole models.UserRole
		if err := json.NewDecoder(r.Body).Decode(&userRole); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if the role exists and is not deleted
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM roles WHERE id = $1 AND deleted_at IS NULL)", userRole.RoleID).Scan(&exists)
		if err != nil {
			http.Error(w, "Error validating role ID", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Role is either deleted or does not exist", http.StatusBadRequest)
			return
		}

		// Insert the user role
		err = db.QueryRow("INSERT INTO user_roles (email, role_id) VALUES ($1, $2) RETURNING id, created_at, updated_at", userRole.Email, userRole.RoleID).
			Scan(&userRole.ID, &userRole.CreatedAt, &userRole.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(userRole)
	}
}

// func CreateUserRole(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var userRole models.UserRole
// 		if err := json.NewDecoder(r.Body).Decode(&userRole); err != nil {
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}

// 		err := db.QueryRow("INSERT INTO user_roles (email, role_id) VALUES ($1, $2) RETURNING id, created_at, updated_at", userRole.Email, userRole.RoleID).
// 			Scan(&userRole.ID, &userRole.CreatedAt, &userRole.UpdatedAt)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		json.NewEncoder(w).Encode(userRole)
// 	}
// }

func UpdateUserRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var userRole models.UserRole
		if err := json.NewDecoder(r.Body).Decode(&userRole); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if the role exists and is not deleted
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM roles WHERE id = $1 AND deleted_at IS NULL)", userRole.RoleID).Scan(&exists)
		if err != nil {
			http.Error(w, "Error validating role ID", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Role is either deleted or does not exist", http.StatusBadRequest)
			return
		}

		// Update the user role
		_, err = db.Exec("UPDATE user_roles SET email = $1, role_id = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND deleted_at IS NULL", 
			userRole.Email, userRole.RoleID, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(userRole)
	}
}



// func UpdateUserRole(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		id := vars["id"]

// 		var userRole models.UserRole
// 		if err := json.NewDecoder(r.Body).Decode(&userRole); err != nil {
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}

// 		_, err := db.Exec("UPDATE user_roles SET email = $1, role_id = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND deleted_at IS NULL", userRole.Email, userRole.RoleID, id)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		json.NewEncoder(w).Encode(userRole)
// 	}
// }

func DeleteUserRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE user_roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}