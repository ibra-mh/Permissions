package app

import (
	"database/sql"
	"main/controllers"
	"github.com/gorilla/mux"
)

func UserRoleRoutes(db *sql.DB,r *mux.Router) {

	r.HandleFunc("/user-roles", controllers.GetUserRoles(db)).Methods("GET")
	r.HandleFunc("/user-roles/{id}", controllers.GetUserRole(db)).Methods("GET")
	r.HandleFunc("/user-roles", controllers.CreateUserRole(db)).Methods("POST")
	r.HandleFunc("/user-roles/{id}", controllers.UpdateUserRole(db)).Methods("PUT")
	r.HandleFunc("/user-roles/{id}", controllers.DeleteUserRole(db)).Methods("DELETE")

}

