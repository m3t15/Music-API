package music

import (
	"Euterpe/internal/logging"
	"Euterpe/internal/shared"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/gabriel-vasile/mimetype"
)

// get an array of music file paths from the music dir
func musicDirContent(musicDir string) ([]string, []string, error) {
	var musicFiles []string
	var failedFiles []string
	var supportedAudioTypes = map[string]bool{
		"audio/mpeg": true,
		"audio/flac": true,
		"audio/ogg":  true,
		"audio/wav":  true,
		"audio/mp4":  true,
	}

	entries, err := os.ReadDir(musicDir)
	if err != nil {
		return musicFiles, failedFiles, err
	}

	for _, item := range entries {
		fullPath := filepath.Join(musicDir, item.Name())

		// recursive directory searching
		if item.IsDir() {
			subMusic, subFailed, err := musicDirContent(fullPath)
			if err != nil {
				failedFiles = append(failedFiles, fullPath)
				continue
			}
			musicFiles = append(musicFiles, subMusic...)
			failedFiles = append(failedFiles, subFailed...)
			continue
		}

		// find MIME type based on the file content
		mtype, err := mimetype.DetectFile(fullPath)
		if err != nil {
			failedFiles = append(failedFiles, fullPath)
			continue
		}

		// check if MIME is audio file
		if supportedAudioTypes[mtype.String()] {
			musicFiles = append(musicFiles, fullPath)
		}
	}
	// log any failures
	for _, item := range failedFiles {
		message := fmt.Sprintf("Failed to ingest file: %s", item)
		logging.Logger("warn", message, shared.GetTime())
	}

	return musicFiles, failedFiles, nil
}

// itterates over the entire music directory CURRENTLY ONLY ON startup and adds the songs to the database
func IngestMusicDir(musicDir string) ([]shared.MusicMetaData, error) {
	musicFiles, _, err := musicDirContent(musicDir)
	if err != nil {
		return nil, err
	}

	var songList []shared.MusicMetaData

	for _, item := range musicFiles {
		musicData, err := extractMetaData(item)
		if err != nil {
			logging.Logger("warn", fmt.Sprintf("Failed to extract metadata for %s: %v", item, err), shared.GetTime())
			continue // Skip bad files instead of killing the whole loop
		}

		// Convert map[string]any to shared.MusicMetaData
		metaData := mapToMusicMetaData(musicData, item)
		songList = append(songList, metaData)
	}

	return songList, nil
}

func extractMetaData(item string) (map[string]any, error) {
	musicData := make(map[string]any)

	// read id3 metadata
	tag, err := id3v2.Open(item, id3v2.Options{Parse: true})
	if err != nil {
		return musicData, fmt.Errorf("id3 error: %w", err)
	}
	defer tag.Close()

	title := tag.Title()
	album := tag.Album()
	artist := tag.Artist()
	trackStr := tag.GetTextFrame("TRCK").Text // track number (TRCK) idk why it is like this but it is
	year := tag.Year()
	genre := tag.Genre()
	var formattedTime string
	if tlen := tag.GetTextFrame("TLEN").Text; tlen != "" {
		if ms, err := strconv.Atoi(tlen); err == nil {
			duration := time.Duration(ms) * time.Millisecond
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) - (minutes * 60)
			formattedTime = fmt.Sprintf("%02d:%02d", minutes, seconds)
		}
	}
	// get track number (using regex)
	re := regexp.MustCompile(`^\d+`)
	match := re.FindString(trackStr)
	trackNum := 0
	if match != "" {
		if i, err := strconv.Atoi(match); err == nil {
			trackNum = i
		}
	}

	// populate musicData map
	musicData["title"] = title
	musicData["album"] = album
	musicData["artist"] = artist
	musicData["track"] = trackNum
	musicData["year"] = year
	musicData["genre"] = genre
	musicData["time"] = formattedTime

	return musicData, nil
}

// helper to move map to struct
func mapToMusicMetaData(data map[string]any, filePath string) shared.MusicMetaData {
	meta := shared.MusicMetaData{
		FilePath: filePath,
	}

	if val, ok := data["title"].(string); ok {
		meta.Title = val
	}
	if val, ok := data["album"].(string); ok {
		meta.Album = val
	}
	if val, ok := data["artist"].(string); ok {
		meta.Artist = val
	}
	if val, ok := data["genre"].(string); ok {
		meta.Genres = val
	}
	if val, ok := data["time"].(string); ok {
		meta.Time = val
	}
	if val, ok := data["track"].(int); ok {
		meta.Track = val
	}

	return meta
}
