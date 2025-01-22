package app

import (
	"database/sql"
	"log"
	"main/utils"
	"net/http"

	"github.com/gorilla/mux"
)

func InitializeRoute(db *sql.DB) {
	r := mux.NewRouter()
	RoleRoutes(db,r)
	UserRoleRoutes(db,r)
	log.Fatal(http.ListenAndServe(":8000", utils.JsonContentTypeMiddleware(r)))
}