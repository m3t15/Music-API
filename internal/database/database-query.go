package database

import (
	"JMAPI/internal/shared"
	"database/sql"
	"math"
	"strings"
)

// MusicFilter holds your optional search parameters
type MusicFilter struct {
	Album  string
	Artist string
	Title  string
	Genre  string
}

func DBQueryPaginatedMusic(db *sql.DB, page int, pageSize int, filter MusicFilter) (shared.PagedData, error) {
	var musicData []shared.MusicMetaData
	offset := (page - 1) * pageSize

	var conditions []string
	var args []any

	// dynamic checks for the query
	if filter.Artist != "" {
		conditions = append(conditions, "ar.artist_name LIKE ?")
		args = append(args, "%"+filter.Artist+"%")
	}
	if filter.Album != "" {
		conditions = append(conditions, "al.album_title LIKE ?")
		args = append(args, "%"+filter.Album+"%")
	}
	if filter.Title != "" {
		conditions = append(conditions, "s.title LIKE ?")
		args = append(args, "%"+filter.Title+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// sql join stuff
	baseJoin := `
		FROM songdata s
		JOIN artistdata ar ON s.artist_id = ar.artist_id
		LEFT JOIN albumdata al ON s.album_id = al.album_id
	`

	// find total records
	countQuery := `SELECT COUNT(*) ` + baseJoin + whereClause
	var totalCount int
	if err := db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return shared.PagedData{}, err
	}

	// limits and offset for page stuff
	args = append(args, pageSize, offset)

	// built query
	query := `
		SELECT 
			s.song_id, 
			ar.artist_name, 
			COALESCE(al.album_title, '') as album_title, 
			s.title 
	` + baseJoin + whereClause + ` ORDER BY s.song_id LIMIT ? OFFSET ?`

	// run main query
	rows, err := db.Query(query, args...)
	if err != nil {
		return shared.PagedData{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var song shared.MusicMetaData
		if err := rows.Scan(&song.SongID, &song.Artist, &song.Album, &song.Title); err != nil {
			return shared.PagedData{}, err
		}
		musicData = append(musicData, song)
	}

	totalPageCount := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return shared.PagedData{
		Data:       musicData,
		TotalPages: totalPageCount,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}
