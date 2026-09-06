package database

import (
	"Euterpe/internal/shared"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// NEED TO LOAD DB PATH FROM CONFIG
func ConnectMusicDB(musicdbdir string) (*sql.DB, error) {
	musicdbPath := musicdbdir + "/music.db"
	// If music.db doesn't exist create it
	// I do not know if this saves time or resources but I think it might
	if _, err := os.Stat(musicdbPath); err != nil {
		if err := createMusicDB(musicdbdir, musicdbPath); err != nil {
			fmt.Println(err)
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", musicdbPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// inserts a bunch of songs into the db and all relational keys
func BulkSongLoad(db *sql.DB, songs []shared.MusicMetaData) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Safely rolls back if any statement fails before tx.Commit()

	// Prepared statements for high performance and duplicate prevention
	stmtArtist, err := tx.Prepare(`INSERT INTO artistdata (artist_name) VALUES (?) ON CONFLICT(artist_name) DO NOTHING;`)
	if err != nil {
		return err
	}
	defer stmtArtist.Close()

	stmtGenre, err := tx.Prepare(`INSERT INTO genredata (genre) VALUES (?) ON CONFLICT(genre) DO NOTHING;`)
	if err != nil {
		return err
	}
	defer stmtGenre.Close()

	stmtAlbum, err := tx.Prepare(`INSERT INTO albumdata (album_title, artist_id) VALUES (?, ?) ON CONFLICT(album_title, artist_id) DO NOTHING;`)
	if err != nil {
		return err
	}
	defer stmtAlbum.Close()

	stmtSong, err := tx.Prepare(`INSERT INTO songdata (title, album_id, track_number, length, filepath) VALUES (?, ?, ?, ?, ?) ON CONFLICT(filepath) DO NOTHING RETURNING song_id;`)
	if err != nil {
		return err
	}
	defer stmtSong.Close()

	stmtSongArtist, err := tx.Prepare(`INSERT INTO song_artists (song_id, artist_id, role) VALUES (?, ?, 'Main') ON CONFLICT DO NOTHING;`)
	if err != nil {
		return err
	}
	defer stmtSongArtist.Close()

	stmtSongGenre, err := tx.Prepare(`INSERT INTO song_genres (song_id, genre_id) VALUES (?, ?) ON CONFLICT DO NOTHING;`)
	if err != nil {
		return err
	}
	defer stmtSongGenre.Close()

	for _, song := range songs {
		if song.Title == "" || song.FilePath == "" {
			continue // Skip incomplete items
		}

		// 1. Insert or ensure Artist exists
		artistName := song.Artist
		if artistName == "" {
			artistName = "Unknown Artist"
		}
		if _, err := stmtArtist.Exec(artistName); err != nil {
			return fmt.Errorf("artist insert error: %w", err)
		}

		var artistID int64
		err = tx.QueryRow(`SELECT artist_id FROM artistdata WHERE artist_name = ?;`, artistName).Scan(&artistID)
		if err != nil {
			return fmt.Errorf("failed to get artist_id: %w", err)
		}

		// 2. Insert or ensure Album exists (if metadata has an album)
		var albumID sql.NullInt64
		if song.Album != "" {
			if _, err := stmtAlbum.Exec(song.Album, artistID); err != nil {
				return fmt.Errorf("album insert error: %w", err)
			}

			var aID int64
			err = tx.QueryRow(`SELECT album_id FROM albumdata WHERE album_title = ? AND artist_id = ?;`, song.Album, artistID).Scan(&aID)
			if err == nil {
				albumID = sql.NullInt64{Int64: aID, Valid: true}
			}
		}

		// 3. Insert Song
		var songID int64
		err = stmtSong.QueryRow(song.Title, albumID, song.Track, song.Time, song.FilePath).Scan(&songID)
		if err != nil {
			// If file was already imported, grab its ID so we don't break remaining relations
			if err == sql.ErrNoRows {
				_ = tx.QueryRow(`SELECT song_id FROM songdata WHERE filepath = ?;`, song.FilePath).Scan(&songID)
			} else {
				return fmt.Errorf("song insert error: %w", err)
			}
		}

		// 4. Attach Song to Artist in Junction Table
		if songID > 0 {
			if _, err := stmtSongArtist.Exec(songID, artistID); err != nil {
				return fmt.Errorf("song_artist insert error: %w", err)
			}
		}

		// 5. Handle Genres (Splits comma-separated or multi-genre strings like "Rock/Pop")
		if song.Genres != "" && songID > 0 {
			genres := splitGenres(song.Genres)
			for _, g := range genres {
				g = strings.TrimSpace(g)
				if g == "" {
					continue
				}

				if _, err := stmtGenre.Exec(g); err != nil {
					return fmt.Errorf("genre insert error: %w", err)
				}

				var genreID int64
				err = tx.QueryRow(`SELECT genre_id FROM genredata WHERE genre = ?;`, g).Scan(&genreID)
				if err == nil {
					_, _ = stmtSongGenre.Exec(songID, genreID)
				}
			}
		}
	}

	return tx.Commit()
}

// Helper to split multi-genre strings like "Rock/Pop" or "Indie, Alternative"
func splitGenres(raw string) []string {
	f := func(c rune) bool {
		return c == ',' || c == '/' || c == ';'
	}
	return strings.FieldsFunc(raw, f)
}
