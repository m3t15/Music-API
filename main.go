package main

import (
	"JMAPI/internal/config"
	"JMAPI/internal/database"
	"JMAPI/internal/logging"
	"JMAPI/internal/music"
	"JMAPI/internal/shared"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

// defaults page and pageSize from the URL
func getPagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 10
	}

	return page, pageSize
}

// encodes any struct/data to JSON
func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logging.Logger("error", fmt.Sprintln("JSON encode error:", err), shared.GetTime())
	}
}

// standard for how how API errors are returned to the client (should always be the same, and logging for it)
func respondError(w http.ResponseWriter, status int, message string, err error) {
	logging.Logger("error", fmt.Sprintln(message+": %v", err), shared.GetTime())
	respondJSON(w, status, map[string]string{"error": message})
}

func main() {
	// load config and initialize variables
	fmt.Println("hiii :3")

	config, err := config.LoadConfig(shared.ConfigFilePath)
	if err != nil {
		logging.Logger("fatal", fmt.Sprintln(err), shared.GetTime())
		os.Exit(1)
	}
	fmt.Println(config)

	// connect to db and init db as needed
	db, err := database.ConnectMusicDB(config.Directories.Databasedir)
	if db == nil {
		if err != nil {
			logging.Logger("fatal", "Unable to connect to DB", shared.GetTime())
			os.Exit(1)
		}
	} else if err != nil {
		logging.Logger("error", fmt.Sprintln(err), shared.GetTime())
	}
	// load and check for new songs on startup
	musicData, err := music.IngestMusicDir(config.Directories.Musicdir)
	if err != nil {
		logging.Logger("error", fmt.Sprintln(err), shared.GetTime())
	}

	err = database.BulkSongLoad(db, musicData)
	if err != nil {
		logging.Logger("error", fmt.Sprintln(err), shared.GetTime())
	}
	// GET REQUESTS

	// Super rudimentary health status checker
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Ping the database first
		err := db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Unable to ping DB..."))
			return
		}

		// send a single OK status and message
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))

	})

	// NEED TO FIGURE OUT HOW TO GET A WAY TO QUERY BY ALBUM AND ARTIST ETC!!!

	// GET /v1/music/get-all?page=X&pageSize=Y (max page size of 100)
	// localhost example: http://localhost:8080/v1/music/get-all?page=1&pageSize=30
	http.HandleFunc("/v1/music/get-all", func(w http.ResponseWriter, r *http.Request) {
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		// enforce max page size
		if pageSize > 100 {
			pageSize = 100
		}
		page, pageSize := getPagination(r)

		pagedData, err := database.DBQueryPaginatedMusic(db, page, pageSize)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to fetch music", err)
			return
		}

		respondJSON(w, http.StatusOK, pagedData)
	})
	// // TO DO: GET request for music by ALBUM
	// http.HandleFunc("/v1/music/album", addPerson)
	// // TO DO: GET request for music by ARTIST
	// http.HandleFunc("/v1/music/artist", getPerson)
	// // TO DO: GET request for music by SONG/TITLE
	// http.HandleFunc("/v1/music/song", getAllPersons)
	// // TO DO: GET request for music by GENRE
	// http.HandleFunc("/v1/music/genre", updatePerson)

	// PATCH REQUESTS

	// // TO DO: PATCH request to UPDATE/COgRRECT MetaData
	// http.HandleFunc("/v1/", deletePerson)
	log.Println("Server running at: http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logging.Logger("fatal", fmt.Sprintln(err), shared.GetTime())
		db.Close()
		os.Exit(1)
	}
	db.Close()
}
