package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	models "main/Models"
	"strconv"

	"net/http"

	"github.com/gorilla/mux"
)


func GetUserRoles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		fmt.Println("Email parameter:", email) // Debugging log

		var rows *sql.Rows
		var err error
		var query string

		query = `SELECT user_roles.id, user_roles.email, user_roles.role_id, user_roles.created_at, user_roles.updated_at,
        user_roles.deleted_at, roles.role_key 
		FROM user_roles 
		LEFT JOIN roles ON user_roles.role_id = roles.id  
		WHERE user_roles.deleted_at IS NULL`

		if email != "" {
			query += ` AND user_roles.email = $1`
			rows, err = db.Query(query, email)
		} else {
			rows, err = db.Query(query)
		}
		if err != nil {
			http.Error(w, "Error fetching user roles: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		userRoles := []models.UserRole{}
		for rows.Next() {
			var userRole models.UserRole
			if err := rows.Scan(&userRole.ID, &userRole.Email, &userRole.RoleID, &userRole.CreatedAt,
				&userRole.UpdatedAt, &userRole.DeletedAt, &userRole.RoleKey); err != nil {
				http.Error(w, "Error scanning user roles: "+err.Error(), http.StatusInternalServerError)
				return
			}
			userRoles = append(userRoles, userRole)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating user roles: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("Number of roles found:", len(userRoles)) // Debugging log
		fmt.Println("User roles:", userRoles)                 // Debugging log
		json.NewEncoder(w).Encode(userRoles)
	}
}


func GetUserRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var userRole models.UserRole
		err := db.QueryRow(`
            SELECT user_roles.id, user_roles.email, user_roles.role_id, user_roles.created_at, user_roles.updated_at, 
                user_roles.deleted_at, roles.role_key 
            FROM user_roles
            LEFT JOIN roles ON user_roles.role_id = roles.id
            WHERE user_roles.id = $1 AND user_roles.deleted_at IS NULL`, id).
			Scan(&userRole.ID, &userRole.Email, &userRole.RoleID, &userRole.CreatedAt,
				&userRole.UpdatedAt, &userRole.DeletedAt, &userRole.RoleKey)

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
        
        // Decode JSON body
        if err := json.NewDecoder(r.Body).Decode(&userRole); err != nil {
            http.Error(w, "Invalid JSON format", http.StatusBadRequest)
            return
        }

        // Check if role exists and is not deleted
        var exists bool
        err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM roles WHERE id = $1 AND deleted_at IS NULL)", userRole.RoleID).Scan(&exists)
        if err != nil {
            http.Error(w, "Database error while checking role", http.StatusInternalServerError)
            return
        }
        if !exists {
            http.Error(w, "Role is either deleted or does not exist", http.StatusBadRequest)
            return
        }

        // Insert the user role
        err = db.QueryRow("INSERT INTO user_roles (email, role_id) VALUES ($1, $2) RETURNING id, created_at, updated_at",
            userRole.Email, userRole.RoleID).Scan(&userRole.ID, &userRole.CreatedAt, &userRole.UpdatedAt)

        if err != nil {
            http.Error(w, "Database error while inserting user role", http.StatusInternalServerError)
            return
        }

        // Return 201 Created status with the new user role
        w.WriteHeader(http.StatusCreated)
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

// 		// Check if the role exists and is not deleted
// 		var exists bool
// 		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM roles WHERE id = $1 AND deleted_at IS NULL)", userRole.RoleID).Scan(&exists)
// 		if err != nil {
// 			http.Error(w, "Error validating role ID", http.StatusInternalServerError)
// 			return
// 		}
// 		if !exists {
// 			http.Error(w, "Role is either deleted or does not exist", http.StatusBadRequest)
// 			return
// 		}

// 		// Insert the user role
// 		err = db.QueryRow("INSERT INTO user_roles (email, role_id) VALUES ($1, $2) RETURNING id, created_at, updated_at", userRole.Email, userRole.RoleID).
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
func DeleteUserRole(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        idStr, exists := vars["id"]
        if !exists {
            http.Error(w, "missing id", http.StatusBadRequest)
            return
        }

        // Convert ID to integer
        id, err := strconv.Atoi(idStr)
        if err != nil {
            log.Printf("Invalid ID: %v", err)
            http.Error(w, "invalid id", http.StatusBadRequest)
            return
        }

        // Execute soft delete query
        res, err := db.Exec("UPDATE user_roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
        if err != nil {
            log.Printf("Database error: %v", err)  // Log error
            http.Error(w, "database error", http.StatusInternalServerError)
            return
        }

        rowsAffected, err := res.RowsAffected()
        if err != nil {
            log.Printf("Error getting rows affected: %v", err)
            http.Error(w, "server error", http.StatusInternalServerError)
            return
        }

        if rowsAffected == 0 {
            http.Error(w, "user role not found", http.StatusNotFound)
            return
        }

        w.WriteHeader(http.StatusNoContent) // 204 success
    }
}

// func DeleteUserRole(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		id := vars["id"]

// 		_, err := db.Exec("UPDATE user_roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		w.WriteHeader(http.StatusNoContent)
// 	}
// }

