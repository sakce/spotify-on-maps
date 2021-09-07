package main

import (
	"database/sql"
	"fmt"
	"net/http"

	// "encoding/json"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

// tbh not sure if this is necessary here yet... we'll see when we introduce the DB
type Location struct {
	ID        string `json:"id"`
	Latitude  string `json:"lat"`
	Longitude string `json:"long"`
}

// DB Connection
const (
	DB_HOST = "localhost:5432"
	DB_NAME = "postgres"
	DB_USER = "postgres"
	DB_PASS = "startups2020"
)

func OpenConnection() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/nearby/{lat}:{long}", getNearby).Methods("GET")

	http.ListenAndServe(":8000", router)
}

// vars need to come from the DB, so let's say we have a users var
// which we will use in the function below to get the locations of users

func getNearby(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) // now we extract params with params["lat"] and params["long"]
}
