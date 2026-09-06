package database

import (
	"JMAPI/internal/shared"
	"database/sql"
)

func DBQueryPaginatedMusic(db *sql.DB, page int, pageSize int) (shared.PagedData, error) {
	var musicData []shared.MusicMetaData
	offset := (page - 1) * pageSize

	// get total records
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM songdata`
	err := db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return shared.PagedData{}, err
	}

	// get records
	query := `
        SELECT 
            s.song_id, 
            ar.artist_name, 
            COALESCE(al.album_title, '') as album_title, 
            s.title 
        FROM songdata s
        JOIN artistdata ar ON s.artist_id = ar.artist_id
        LEFT JOIN albumdata al ON s.album_id = al.album_id
        ORDER BY s.song_id 
        LIMIT $1 OFFSET $2
    `
	rows, err := db.Query(query, pageSize, offset)
	if err != nil {
		return shared.PagedData{}, err
	}

	defer rows.Close()
	for rows.Next() {

		var songData shared.MusicMetaData
		if err := rows.Scan(&songData.SongID, &songData.Artist, &songData.Album, &songData.Title); err != nil {
			return shared.PagedData{}, err
		}

		musicData = append(musicData, songData)
	}

	totalPageCount := (totalCount / pageSize)

	return shared.PagedData{
		Data:       musicData,
		TotalPages: totalPageCount,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil

}
