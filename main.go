package main

import (
	"Euterpe/internal/config"
	"Euterpe/internal/database"
	"Euterpe/internal/logging"
	"Euterpe/internal/music"
	"Euterpe/internal/shared"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// load config and initialize variables
	fmt.Println("hiii :3")

	config, err := config.LoadConfig(shared.ConfigFilePath)
	if err != nil {
		logging.Logger("fatal", fmt.Sprintln(err), shared.GetTime())
	}
	fmt.Println(config)
	// System Checking
	http.HandleFunc("/health", logging.HealthData)

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
	musicData, err := music.IngestMusicDir(config.Directories.Musicdir)
	if err != nil {
		logging.Logger("error", fmt.Sprintln(err), shared.GetTime())
	}

	err = database.BulkSongLoad(db, musicData)
	if err != nil {
		logging.Logger("error", fmt.Sprintln(err), shared.GetTime())
	}
	// GET REQUESTS

	// NEED TO FIGURE OUT HOW TO GET A WAY TO QUERY BY ALBUM AND ARTIST ETC!!!

	// // TO DO: GET request ALL MUSIC (add paging, not sure how I want to sort this once more data gets added in)
	// http.HandleFunc("/v1/music/all", music.GetAllMusic)
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
	log.Fatal(http.ListenAndServe(":8080", nil))

	db.Close()
}
