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
	return rad * 180 / math.Pi
}

func deg2rad(deg float64) float64 {
	return deg * math.Pi / 180
}

func getNearby(w http.ResponseWriter, r *http.Request) {
	initDb() // Open connection
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) // now we extract params with params["lat"] and params["long"]
	latVar, _ := params["lat"]
	lonVar, _ := params["long"]
	radVar, _ := params["rad"] // radius of the bounding circle

	lat, _ := strconv.ParseFloat(latVar, 64) // in degrees
	lon, _ := strconv.ParseFloat(lonVar, 64) // in degrees
	rad, _ := strconv.ParseFloat(radVar, 64) // in km

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

func queryLocations(locations *nearbyUsers, minLat, maxLat, minLon, maxLon, latDeg, lonDeg, R, rad float64) error {
	
	// Prepare the statement first, then Query with the variables to avoid vulnerability of a SQL injection

	stmt, err := db.Prepare("WITH consts (minLat, maxLat, minLon, maxLon, latDeg, lonDeg, R, radius) as (values ($1, $2, $3, $4, $5, $6, $7, $8)) " + 
							"SELECT firstCut.\"userID\", firstCut.\"latitude\", firstCut.\"longitude\", " + 
							// Distance
							"acos(" + 
								"sin(CAST(consts.latDeg AS DOUBLE PRECISION)) * sin(radians(firstCut.\"latitude\")) + " + 
								"cos(CAST(consts.latDeg AS DOUBLE PRECISION)) * cos(radians(firstCut.\"latitude\")) * " + 
								"cos(radians(firstCut.\"longitude\") - CAST(consts.lonDeg AS DOUBLE PRECISION))" + 
							") * CAST(consts.R AS DOUBLE PRECISION) AS distance " + 
							// FROM -> sub-query
							"FROM (" + 
								"SELECT l.\"userID\" as \"userID\", " + 
										"l.\"latitude\" as \"latitude\", " +  
										"l.\"longitude\" as \"longitude\" " + 
								"FROM loc AS l, consts " + 
								// defining the rectangle bounding box
								"WHERE l.\"latitude\" BETWEEN CAST(consts.minLat AS DOUBLE PRECISION) AND CAST(consts.maxLat AS DOUBLE PRECISION) AND " +
										"l.\"longitude\" BETWEEN CAST(consts.minLon AS DOUBLE PRECISION) AND CAST(consts.maxLon AS DOUBLE PRECISION) " + 
							") AS firstCut, consts " + 
							// Cosine(?) law
							"WHERE acos(" + 
								"sin(CAST(consts.latDeg AS DOUBLE PRECISION)) * sin(radians(firstCut.\"latitude\")) + " + 
								"cos(CAST(consts.latDeg AS DOUBLE PRECISION)) * cos(radians(firstCut.\"latitude\")) * " + 
								"cos(radians(firstCut.\"longitude\") - CAST(consts.lonDeg AS DOUBLE PRECISION))" + 
							") * CAST(consts.R AS DOUBLE PRECISION) " +
							// WHERE distance less than the radius
							"< CAST(consts.radius AS DOUBLE PRECISION) AND " +
							"acos(" + 
								"sin(CAST(consts.latDeg AS DOUBLE PRECISION)) * sin(radians(firstCut.\"latitude\")) + " + 
								"cos(CAST(consts.latDeg AS DOUBLE PRECISION)) * cos(radians(firstCut.\"latitude\")) * " + 
								"cos(radians(firstCut.\"longitude\") - CAST(consts.lonDeg AS DOUBLE PRECISION))" + 
							// WHERE distance is not 0 - since we don't need the location of the looked up user
							") * CAST(consts.R AS DOUBLE PRECISION) != 0 " +
							"ORDER BY distance;")

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

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
			&nearbyUser.Distance,
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

	current := currentSong{}

	err := queryCurrentSong(&current, userID)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(current)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))

	defer db.Close()
	return
}

type currentSongSummary struct {
	UserID    float64
	SongName  string
	SongArtist string
}

type currentSong struct {
	CurrentSong []currentSongSummary
}

func queryCurrentSong(songs *currentSong, userID string) error {
	stmt, err := db.Prepare("WITH consts (lookupUser) as (values ($1)) " + 
							"SELECT l.\"userID\", m.\"songName\", m.\"artist\" " + 
							"FROM listens l, music m, consts " + 
							"WHERE l.\"songID\" = m.\"songID\" AND CAST(consts.lookupUser AS INTEGER) = l.\"userID\";")

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		currentSong := currentSongSummary{}
		err = rows.Scan(
			&currentSong.UserID,
			&currentSong.SongName,
			&currentSong.SongArtist,
		)
		if err != nil {
			return err
		}
		songs.CurrentSong = append(songs.CurrentSong, currentSong)
	}

	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
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
