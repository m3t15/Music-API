package database

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// this exists SIMPLY so wicked big SQLite creations aren't in database.go with the normal sized stuff
// basically I just think this looks cleaner

func createMusicDB(musicdbdir string, musicdbPath string) error {
	// Create database location and DB if it doesn't already exist
	if err := os.MkdirAll(musicdbdir, 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", musicdbPath)
	if err != nil {
		return err
	}

	createArtistTable := `
	CREATE TABLE IF NOT EXISTS "artistdata" (
    "artist_id" INTEGER PRIMARY KEY AUTOINCREMENT, 
    "artist_name" TEXT NOT NULL UNIQUE
	);`

	createGenreTable := `
	CREATE TABLE IF NOT EXISTS "genredata" (
    "genre_id" INTEGER PRIMARY KEY AUTOINCREMENT, 
    "genre" TEXT NOT NULL UNIQUE
	);`

	createAlbumTable := `
	CREATE TABLE IF NOT EXISTS "albumdata" (
    "album_id" INTEGER PRIMARY KEY AUTOINCREMENT, 
    "album_title" TEXT NOT NULL, 
    "artist_id" INTEGER NOT NULL, 
    "year" INTEGER NULL, 
    "cover_path" TEXT NULL,
    FOREIGN KEY ("artist_id") REFERENCES "artistdata"("artist_id") ON DELETE CASCADE,
    UNIQUE("album_title", "artist_id")
	);`

	createSongsTable := `CREATE TABLE IF NOT EXISTS "songdata" (
    "song_id" INTEGER PRIMARY KEY AUTOINCREMENT, 
    "title" TEXT NOT NULL, 
    "album_id" INTEGER NULL, -- NULL allowed (in case of singles)
    "artist_id" INTEGER NOT NULL, 
    "track_number" INTEGER NULL,
    "length" TEXT NOT NULL,
    "comment" TEXT NULL, 
    "filepath" TEXT NOT NULL UNIQUE,
    FOREIGN KEY ("artist_id") REFERENCES "artistdata"("artist_id") ON DELETE CASCADE,
    FOREIGN KEY ("album_id") REFERENCES "albumdata"("album_id") ON DELETE SET NULL
	);`

	createJuncSongGenre := `
	CREATE TABLE IF NOT EXISTS "song_genres" (
    "song_id" INTEGER NOT NULL,
    "genre_id" INTEGER NOT NULL,
    PRIMARY KEY ("song_id", "genre_id"),
    FOREIGN KEY ("song_id") REFERENCES "songdata"("song_id") ON DELETE CASCADE,
    FOREIGN KEY ("genre_id") REFERENCES "genredata"("genre_id") ON DELETE CASCADE
	);`

	tableCreateArr := []string{createArtistTable, createGenreTable, createAlbumTable, createSongsTable, createJuncSongGenre}
	for i := 0; i < len(tableCreateArr); i++ {
		_, err = db.Exec(tableCreateArr[i])
		if err != nil {
			return err

		}
	}
	db.Close()
	return nil
}
