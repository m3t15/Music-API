package shared

import "time"

const (
	ConfigFilePath = "./config.yml"
)

// Shared Music MetaData
type MusicMetaData struct {
	SongID   *int   `json:"songid,omitempty"`
	Title    string `json:"title"`
	Album    string `json:"album"`
	Track    int    `json:"track"`
	Artist   string `json:"artist"`
	Genres   string `json:"genre"` //MIGHT need to make this a pointer to account for a missing value, or do something with if does not exist set to nothing "" OR hydrate from MusicBrainz?
	Time     string `json:"time"`
	FilePath string `json:"filepath"`
}

// Tiny datetime function becuase I'm kind of lazy and want it all uniform
func GetTime() string {
	RFC1123Z := "Mon, 02 Jan 2006 15:04:05 -0700"
	timeNow := time.Now().Format(RFC1123Z)
	return timeNow
}

// Struct for paging data
type PagedData struct {
	Data       []MusicMetaData `json:"musicData"`
	TotalCount int             `json:"totalCount"`
	TotalPages int             `json:"totalPages"`
	Page       int             `json:"page"`
	PageSize   int             `json:"pageSize"`
}
