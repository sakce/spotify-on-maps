package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/nearby/{lat}:{long}", getNearby).Methods("GET")
	router.HandleFunc("/current_song/{userID}", getCurrentSong).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}

// tbh not sure if this is necessary here yet... we'll see when we introduce the DB
// type Location struct {
// 	ID        string `json:"id"`
// 	Latitude  string `json:"lat"`
// 	Longitude string `json:"long"`
// }

func getNearby(w http.ResponseWriter, r *http.Request) {
	initDb() // Open connection
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) // now we extract params with params["lat"] and params["long"]
	// fmt.Println(params["lat"], params["long"])
	// lat := params["lat"]
	// long := params["long"]

	// result_locations
	// json.NewEncoder(w).Encode(result_locations)

	json.NewEncoder(w).Encode(params)

	defer db.Close()
	return
}

func getCurrentSong(w http.ResponseWriter, r *http.Request) {
	//...
	initDb() // Open connection
	w.Header().Set("Content-Type", "application/json")
	userID := mux.Vars(r)["userID"]

	json.NewEncoder(w).Encode(userID)

	// result_song = ...
	// json.NewEncoder(w).Encode(result_song)

	defer db.Close()
	return
}

func initDb() {
	config := dbConfig()
	// fmt.Println(config)
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

	fmt.Println(psqlInfo)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
}

func dbConfig() map[string]string {
	conf := make(map[string]string)
	host, ok := os.LookupEnv(dbhost)
	if !ok {
		panic("DBHOST environment variable required but not set")
	}
	port, ok := os.LookupEnv(dbport)
	if !ok {
		panic("DBPORT environment variable required but not set")
	}
	user, ok := os.LookupEnv(dbuser)
	if !ok {
		panic("DBUSER environment variable required but not set")
	}
	password, ok := os.LookupEnv(dbpass)
	if !ok {
		panic("DBPASS environment variable required but not set")
	}
	name, ok := os.LookupEnv(dbname)
	if !ok {
		panic("DBNAME environment variable required but not set")
	}
	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name
	return conf
}
