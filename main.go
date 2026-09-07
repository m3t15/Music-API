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

// defaults page and pageSize and enforces a max size
func getPagination(r *http.Request, maxPageSize int) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	switch {
	case pageSize < 1:
		pageSize = 10
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	}

	return page, pageSize
}

// encodes any data to JSON
func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logging.Logger("error", fmt.Sprintln("JSON encode error:", err), shared.GetTime())
	}
}

// standardize how API errors are returned and logged
func respondError(w http.ResponseWriter, status int, message string, err error) {
	logging.Logger("error", fmt.Sprintf("%s: %v", message, err), shared.GetTime())
	respondJSON(w, status, map[string]string{"error": message})
}

func main() {
	// I'm leaving this in, it was the first line I wrote for this before I had any idea what kinda music thing I was making therefore it stays
	fmt.Println("hiii :3")

	config, err := config.LoadConfig(shared.ConfigFilePath)
	if err != nil {
		logging.Logger("fatal", fmt.Sprintln(err), shared.GetTime())
		os.Exit(1)
	}

	// Connect to db
	db, err := database.ConnectMusicDB(config.Directories.Databasedir)
	if err != nil || db == nil {
		logging.Logger("fatal", fmt.Sprintf("Unable to connect to DB: %v", err), shared.GetTime())
		os.Exit(1)
	}
	// validate that the db is closed when exiting the program
	defer db.Close()

	// Load and check for new songs on startup
	musicData, err := music.IngestMusicDir(config.Directories.Musicdir)
	if err != nil {
		logging.Logger("error", fmt.Sprintln(err), shared.GetTime())
	}

	err = database.BulkSongLoad(db, musicData)
	if err != nil {
		logging.Logger("error", fmt.Sprintln(err), shared.GetTime())
	}

	// GET REQUESTS

	// simple health checker
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Unable to ping DB..."))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// query DB for music
	// example: /v1/music?album=ALBUMNAME&artist=FIRST+LAST&page=1&pageSize=50
	http.HandleFunc("/v1/music", func(w http.ResponseWriter, r *http.Request) {
		page, pageSize := getPagination(r, 100)

		// filters from the URL query to struct
		filter := database.MusicFilter{
			Album:  r.URL.Query().Get("album"),
			Artist: r.URL.Query().Get("artist"),
			Genre:  r.URL.Query().Get("genre"),
			Title:  r.URL.Query().Get("song"),
		}

		pagedData, err := database.DBQueryPaginatedMusic(db, page, pageSize, filter)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to fetch music", err)
			return
		}

		respondJSON(w, http.StatusOK, pagedData)
	})

	// PATCH REQUESTS

	// TO DO: PATCH to update stuffs
	// http.HandleFunc("/v1/music/update", func(w http.ResponseWriter, r *http.Request) { ... })
	// http.HandleFunc("/v1/music/update", func(w http.ResponseWriter, r *http.Request) {
	// 	trackNum, err := strconv.Atoi((URL.Query().Get("updatetrack")))
	// 	if err != nil {
	// 		respondError(w, http.StatusInternalServerError, "Failed to fetch music", err)
	// 		return
	// 	}
	// 	songid, err := strconv.Atoi(r.URL.Query().Get("songid"))
	// 	if err != nil {
	// 		respondError(w, http.StatusInternalServerError, "Failed to fetch music", err)
	// 		return
	// 	}

	// 	filter := shared.MusicMetaData{
	// 		SongID: songid,
	// 		Title:  r.URL.Query().Get("updatetitle"),
	// 		Album:  r.URL.Query().Get("updatealbum"),
	// 		Track:  trackNum,
	// 		Artist: r.URL.Query().Get("updateartist"),
	// 		Genres: r.URL.Query().Get("updategenres"),
	// 		Time:   r.URL.Query().Get("updatetime"),
	// 	}

	// 	// pagedData, err := database.DBUpdateMusic(db, filter)
	// 	// if err != nil {
	// 	// 	respondError(w, http.StatusInternalServerError, "Failed to fetch music", err)
	// 	// 	return
	// 	// }

	// 	respondJSON(w, http.StatusOK, pagedData)
	// })

	log.Println("Server running at: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logging.Logger("fatal", fmt.Sprintln(err), shared.GetTime())
		os.Exit(1)
	}
}
