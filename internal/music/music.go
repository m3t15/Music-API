package music

import (
	"JMAPI/internal/logging"
	"JMAPI/internal/shared"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/dhowden/tag"
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

		// recursive dir search stuff
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

		// find MIME type via file content
		mtype, err := mimetype.DetectFile(fullPath)
		if err != nil {
			failedFiles = append(failedFiles, fullPath)
			continue
		}

		if supportedAudioTypes[mtype.String()] {
			musicFiles = append(musicFiles, fullPath)
		}
	}
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

// get music metadata w/ fallback
func extractMetaData(item string) (map[string]any, error) {
	musicData := make(map[string]any)
	musicData["filepath"] = item

	// try for quick and easy parsing (can grab duration and stuff on standard MP3s)
	id3tag, err := id3v2.Open(item, id3v2.Options{Parse: true})

	// if the file has some sort of issue being read, or if it is not an MP3 fallback to using dhowden/tag
	if err != nil {
		f, openErr := os.Open(item)
		if openErr != nil {
			return musicData, fmt.Errorf("failed to open file for fallback: %w", openErr)
		}
		defer f.Close()

		m, tagErr := tag.ReadFrom(f)
		if tagErr != nil {
			return musicData, fmt.Errorf("native parse and fallback tag parse failed: %w", tagErr)
		}

		if m.Title() != "" {
			musicData["title"] = m.Title()
		}
		if m.Album() != "" {
			musicData["album"] = m.Album()
		}
		if m.Artist() != "" {
			musicData["artist"] = m.Artist()
		}
		if m.Genre() != "" {
			musicData["genre"] = m.Genre()
		}
		musicData["year"] = m.Year()

		trackNum, _ := m.Track()
		musicData["track"] = trackNum

		return musicData, nil
	}
	defer id3tag.Close()

	musicData["title"] = id3tag.Title()
	musicData["album"] = id3tag.Album()
	musicData["artist"] = id3tag.Artist()
	musicData["genre"] = id3tag.Genre()
	musicData["year"] = id3tag.Year()

	trackStr := id3tag.GetTextFrame("TRCK").Text
	re := regexp.MustCompile(`^\d+`)
	if match := re.FindString(trackStr); match != "" {
		if i, err := strconv.Atoi(match); err == nil {
			musicData["track"] = i
		}
	}

	if tlen := id3tag.GetTextFrame("TLEN").Text; tlen != "" {
		if ms, err := strconv.Atoi(tlen); err == nil {
			duration := time.Duration(ms) * time.Millisecond
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) - (minutes * 60)
			musicData["time"] = fmt.Sprintf("%02d:%02d", minutes, seconds)
		}
	}

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
