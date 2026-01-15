package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
	"github.com/mitchellh/go-homedir"
)

type Song struct {
	tag.Metadata
	path    string
	isVideo bool
}

var songDir string

func main() {
	var err error
	var fileList []string

	if len(os.Args) > 1 {
		songDir = os.Args[1]
		fileList, err = getSongList(songDir)
		if err != nil {
			log.Fatal("Can't get media list")
		}
	} else {
		// Scan both Music and Videos directories
		musicDir, err := homedir.Expand("~/Music/")
		if err != nil {
			log.Fatal("Can't open ~/Music directory")
		}
		videosDir, err := homedir.Expand("~/Videos/")
		if err != nil {
			log.Fatal("Can't open ~/Videos directory")
		}

		musicFiles, err := getSongList(musicDir)
		if err != nil {
			log.Fatal("Can't get music list")
		}

		videoFiles, err := getSongList(videosDir)
		if err != nil {
			log.Fatal("Can't get video list")
		}

		// Combine both lists
		fileList = append(musicFiles, videoFiles...)
		songDir, err = homedir.Dir()
		if err != nil {
			log.Fatal("Can't get home directory")
		}
	}
	songs := make([]Song, 0, len(fileList))

	for _, fileName := range fileList {
		currentFile, err := os.Open(fileName)
		if err == nil {
			metadata, _ := tag.ReadFrom(currentFile)
			isVideo := isVideoFile(fileName)
			songs = append(songs, Song{
				Metadata: metadata,
				path:     fileName,
				isVideo:  isVideo,
			})
		}
		currentFile.Close()
	}
	if len(songs) == 0 {
		log.Fatal("Could not find any media to play")
	}
	userInterface, err := NewUi(songs, len(songDir))
	if err != nil {
		log.Fatal(err)
	}
	userInterface.OnSelect = selectMedia
	userInterface.OnPause = pauseSong
	userInterface.OnSeek = seek
	userInterface.OnVolume = setVolue
	userInterface.Start()
	defer userInterface.Close()
}

func isVideoFile(path string) bool {
	ext := filepath.Ext(path)
	videoExts := []string{".mp4", ".mkv", ".avi", ".mov", ".webm"}
	for _, v := range videoExts {
		if ext == v {
			return true
		}
	}
	return false
}

func selectMedia(song Song) (int, error) {
	if song.isVideo {
		return playVideo(song)
	}
	return playSong(song)
}
