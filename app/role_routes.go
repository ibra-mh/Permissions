package app

import (
	"database/sql"
	// "log"
	"main/controllers"
	// "main/utils"
	// "net/http"

	"github.com/gorilla/mux"
)

func RoleRoutes(db *sql.DB,r *mux.Router) {

	r.HandleFunc("/roles", controllers.GetRoles(db)).Methods("GET")
	r.HandleFunc("/roles/{id}", controllers.GetRole(db)).Methods("GET")
	r.HandleFunc("/roles", controllers.CreateRole(db)).Methods("POST")
	r.HandleFunc("/roles/{id}", controllers.UpdateRole(db)).Methods("PUT")
	r.HandleFunc("/roles/{id}", controllers.DeleteRole(db)).Methods("DELETE")

}
