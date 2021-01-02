package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

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
	router.HandleFunc("/nearby/{lat}:{long}:{rad}", getNearby).Methods("GET")
	router.HandleFunc("/current_song/{userID}", getCurrentSong).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func rad2deg(rad float64) float64 {
	return rad * (math.Pi / 180)
}

func deg2rad(deg float64) float64 {
	return deg / (math.Pi / 180)
}

func getNearby(w http.ResponseWriter, r *http.Request) {
	initDb() // Open connection
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) // now we extract params with params["lat"] and params["long"]
	latVar, _ := params["lat"]
	lonVar, _ := params["long"]
	radVar, _ := params["rad"] // radius of the bounding circle

	// FROM HERE
	lat, _ := strconv.ParseFloat(latVar, 64)
	lon, _ := strconv.ParseFloat(lonVar, 64)
	rad, _ := strconv.ParseFloat(radVar, 64)
	rad = rad * 1000

	latDeg := deg2rad(lat) // to use in the query
	lonDeg := deg2rad(lon) // to use in the query

	R := float64(6371)

	radiusOverEarth := rad / R

	maxLat := lat + rad2deg(radiusOverEarth)
	minLat := lat - rad2deg(radiusOverEarth)

	maxLon := lon + rad2deg(math.Asin(radiusOverEarth)/math.Cos(deg2rad(lat)))
	minLon := lon - rad2deg(math.Asin(radiusOverEarth)/math.Cos(deg2rad(lat)))

	nearby := nearbyUsers{}

	err := queryLocations(&nearby, minLat, maxLat, minLon, maxLon, latDeg, lonDeg, R, rad)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(nearby)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))

	defer db.Close()
	return
}

type userLocationSummary struct {
	UserID    float64
	Latitude  float64
	Longitude float64
	Distance  float64
}

type nearbyUsers struct {
	NearbyUsers []userLocationSummary
}

// $1, $2, $3, $4, $5, $6, $7, $8
// ?, ?, ?, ?, ?, ?, ?, ?

func queryLocations(locations *nearbyUsers, minLat, maxLat, minLon, maxLon, latDeg, lonDeg, R, rad float64) error {

	stmt, err := db.Prepare("WITH consts (minLat, maxLat, minLon, maxLon, " +
		"latDeg, lonDeg, R, radius) as (values (CAST($1 AS DOUBLE PRECISION), CAST($2 AS DOUBLE PRECISION), CAST($3 AS DOUBLE PRECISION), CAST($4 AS DOUBLE PRECISION), CAST($5 AS DOUBLE PRECISION), CAST($6 AS DOUBLE PRECISION), CAST($7 AS DOUBLE PRECISION), CAST($8 AS DOUBLE PRECISION))) " +

		"SELECT l.\"userID\", l.\"latitude\", l.\"longitude\", " +
		// calculation for distance
		"acos(sin(consts.latDeg)*sin(radians(\"latitude\")) + " +
		"cos(consts.latDeg)*cos(radians(\"latitude\")) * " +
		"cos(radians(\"longitude\")-consts.lonDeg)) * consts.R AS distance " +

		// Bounding square box - sub-query
		"FROM (SELECT l.\"userID\", l.\"latitude\", l.\"longitude\" " +
		"FROM loc l, consts WHERE \"latitude\" " +
		"BETWEEN consts.minLat AND consts.maxLat AND \"longitude\" " +
		"BETWEEN consts.minLon AND consts.maxLon) AS FirstCut, consts " +

		"WHERE acos(sin(consts.latDeg)*sin(radians(\"latitude\")) + " +
		"cos(consts.latDeg)*cos(radians(\"latitude\")) * " +
		"cos(radians(\"longitude\")-consts.lonDeg)) * consts.R < consts.radius " +
		"ORDER BY distance;")

	// sub-query
	// stmt, err := db.Prepare("WITH consts (minLat, maxLat, minLon, maxLon, " +
	// "R, radius) as (values ($1, $2, $3, $4, $5, $6)) " +
	// "SELECT pg_typeof(CAST(consts.minlat AS FLOAT)), pg_typeof(CAST(consts.maxLon AS FLOAT)), pg_typeof(l.\"longitude\") FROM loc l, consts;")
	// l.\"userID\", l.\"latitude\", l.\"longitude\" " +
	// "FROM loc l, consts WHERE (\"latitude\" " +
	// "BETWEEN consts.minLat AND consts.maxLat) AND (\"longitude\" " +
	// "BETWEEN consts.minLon AND consts.maxLon);")

	// stmt, err := db.Prepare("SELECT latitude, firstName FROM account AS a, loc AS l WHERE l.userID = a.userID")

	if err != nil {
		fmt.Println("Right here mate! ")
		log.Fatal(err)
	}
	defer stmt.Close()

	fmt.Println("Passed the preparation mate! \n")

	// sub-query
	// rows, err := stmt.Query(minLat, maxLat, minLon, maxLon, R, rad)

	rows, err := stmt.Query(minLat, maxLat, minLon, maxLon, latDeg, lonDeg, R, rad)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		nearbyUser := userLocationSummary{}
		err = rows.Scan(
			&nearbyUser.UserID,
			&nearbyUser.Latitude,
			&nearbyUser.Longitude,
		)
		if err != nil {
			return err
		}
		locations.NearbyUsers = append(locations.NearbyUsers, nearbyUser)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func getCurrentSong(w http.ResponseWriter, r *http.Request) {
	//...
	initDb() // Open connection
	w.Header().Set("Content-Type", "application/json")
	userID := mux.Vars(r)["userID"]

	json.NewEncoder(w).Encode(userID)

	// TODO
	// result_song = ...
	// json.NewEncoder(w).Encode(result_song)

	defer db.Close()
	return
}

func initDb() {
	config := dbConfig()

	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

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
