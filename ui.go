package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/dhowden/tag"
	"github.com/gizak/termui"
)

type uiState int

const (
	Stopped uiState = iota
	Playing
	Paused
)

type uiPlayMode int

const (
	LoopAll uiPlayMode = iota
	LoopOne
	NoLoop
	Random
)

var playModeName = []string{"Loop All", "Loop One", "No Loop", "Random"}

type selectCallback func(Song) (int, error)
type pauseCallback func(bool)
type seekCallback func(int) error
type volumeCallback func(int)

type Ui struct {
	infoList      *termui.List
	playList      *termui.List
	scrollerGauge *termui.Gauge
	volumeGauge   *termui.Gauge
	controlsPar1  *termui.Par
	controlsPar2  *termui.Par

	songs     []Song
	songNames []string

	volume  int
	songNum int // index of playing song
	songSel int // index of the song that is playing
	songPos int // index of position of the song that is being played
	songLen int // length of the song

	OnSelect selectCallback
	OnPause  pauseCallback
	OnSeek   seekCallback
	OnVolume volumeCallback

	state    uiState
	playMode uiPlayMode
}

func NewUi(songList []Song, pathPrefix int) (*Ui, error) {
	err := termui.Init()
	if err != nil {
		return nil, err
	}

	ui := new(Ui)

	ui.volume = 100

	ui.songs = songList
	ui.songNum = -1
	ui.infoList = termui.NewList()
	ui.infoList.BorderLabel = "Song info"
	ui.infoList.BorderFg = termui.ColorGreen
	ui.infoList.Overflow = "wrap"

	ui.playList = termui.NewList()
	ui.playList.BorderLabel = "Playlist"
	ui.playList.BorderFg = termui.ColorGreen

	ui.scrollerGauge = termui.NewGauge()
	ui.scrollerGauge.BorderLabel = "Stopped"
	ui.scrollerGauge.Height = 3

	ui.volumeGauge = termui.NewGauge()
	ui.volumeGauge.BorderLabel = "Volume"
	ui.volumeGauge.Height = 3
	ui.volumeGauge.Percent = ui.volume

	ui.controlsPar1 = termui.NewPar(
		"[ < ](fg-black,bg-white)[Previous](fg-black,bg-green) " +
			"[Left ](fg-black,bg-white)[-10s](fg-black,bg-green) " +
			"[ - ](fg-black,bg-white)[-Volume](fg-black,bg-green) " +
			"[ Up ](fg-black,bg-white)[Move Up  ](fg-black,bg-green) " +
			"[Enter](fg-black,bg-white)[Select](fg-black,bg-green) " +
			"[ m ](fg-black,bg-white)[Change Play Mode](fg-black,bg-green) " +
			"[ " + playModeName[ui.playMode] + " ](fg-black,bg-yellow)")
	ui.controlsPar1.Border = false
	ui.controlsPar1.Height = 1
	ui.controlsPar2 = termui.NewPar(
		"[ > ](fg-black,bg-white)[Next    ](fg-black,bg-green) " +
			"[Right](fg-black,bg-white)[+10s](fg-black,bg-green) " +
			"[ + ](fg-black,bg-white)[+Volume](fg-black,bg-green) " +
			"[Down](fg-black,bg-white)[Move Down](fg-black,bg-green) " +
			"[ Esc ](fg-black,bg-white)[Stop  ](fg-black,bg-green) " +
			"[ Space ](fg-black,bg-white)[Pause/Resume](fg-black,bg-green) " +
			"[ r ](fg-black,bg-white)[Refresh](fg-black,bg-green) " +
			"[ q ](fg-black,bg-white)[Exit](fg-black,bg-green) ")
	ui.controlsPar2.Border = false
	ui.controlsPar2.Height = 1

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(6, 0, ui.infoList, ui.scrollerGauge, ui.volumeGauge),
			termui.NewCol(6, 0, ui.playList)),
		termui.NewRow(termui.NewCol(12, 0, ui.controlsPar1)),
		termui.NewRow(termui.NewCol(12, 0, ui.controlsPar2)),
	)

	ui.realign()

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/Q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/r", func(termui.Event) {
		ui.reloadSongList(pathPrefix)
	})

	termui.Handle("/sys/kbd/R", func(termui.Event) {
		ui.reloadSongList(pathPrefix)
	})

	termui.Handle("/sys/kbd/<space>", func(termui.Event) {
		if ui.songNum != -1 {
			if ui.state == Playing {
				ui.OnPause(true)
				ui.state = Paused
			} else {
				ui.OnPause(false)
				ui.state = Playing

			}
			ui.renderStatus()
		}
	})
	termui.Handle("timer/1s", func(termui.Event) {
		if ui.state == Playing {
			ui.songPos++
			if ui.songLen != 0 {
				ui.scrollerGauge.Percent = int(float32(ui.songPos) / float32(ui.songLen) * 100)
				ui.scrollerGauge.Label = fmt.Sprintf("%d:%.2d / %d:%.2d", ui.songPos/60, ui.songPos%60, ui.songLen/60, ui.songLen%60)
				if ui.scrollerGauge.Percent >= 100 {
					ui.songEndEvent()
				}
				termui.Clear()
				termui.Render(termui.Body)
			}
		} else if ui.state == Stopped {
			ui.songPos = 0
		}
	})

	termui.Handle("/sys/kbd/<right>", func(termui.Event) {
		if ui.songNum != -1 {
			ui.songPos += 10
			ui.OnSeek(ui.songPos)
		}
	})

	termui.Handle("/sys/kbd/<left>", func(termui.Event) {
		if ui.songNum != -1 {
			ui.songPos -= 10
			if ui.songPos < 0 {
				ui.songPos = 0
			}
			ui.OnSeek(ui.songPos)
		}
	})

	termui.Handle("/sys/kbd/<escape>", func(termui.Event) {
		ui.stopPlayer()
	})

	termui.Handle("/sys/kbd/<enter>", func(termui.Event) {
		ui.songNum = ui.songSel
		ui.playSong(ui.songNum)
	})

	termui.Handle("/sys/kbd/<up>", func(termui.Event) {
		ui.songUp()
		termui.Clear()
		termui.Render(termui.Body)
	})

	termui.Handle("/sys/kbd/=", func(termui.Event) {
		ui.volumeUp()
	})

	termui.Handle("/sys/kbd/+", func(termui.Event) {
		ui.volumeUp()
	})

	termui.Handle("/sys/kbd/-", func(termui.Event) {
		ui.volumeDown()
	})

	termui.Handle("/sys/kbd/_", func(termui.Event) {
		ui.volumeDown()
	})

	termui.Handle("/sys/kbd/<down>", func(termui.Event) {
		ui.songDown()
		termui.Clear()
		termui.Render(termui.Body)
	})

	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		ui.realign()
	})

	termui.Handle("/sys/kbd/<", func(e termui.Event) {
		ui.previous()
	})

	termui.Handle("/sys/kbd/,", func(e termui.Event) {
		ui.previous()
	})

	termui.Handle("/sys/kbd/>", func(e termui.Event) {
		ui.next()
	})

	termui.Handle("/sys/kbd/.", func(e termui.Event) {
		ui.next()
	})

	termui.Handle("/sys/kbd/m", func(e termui.Event) {
		ui.togglePlayMode()
	})

	termui.Handle("/sys/kbd/M", func(e termui.Event) {
		ui.togglePlayMode()
	})

	ui.songNames = make([]string, len(ui.songs))
	ui.updateSongNames(pathPrefix)
	ui.playList.Items = ui.songNames
	ui.setSong(0, false)

	return ui, nil
}

func (ui *Ui) Start() {
	termui.Loop()
}

func (ui *Ui) Close() {
	termui.Close()
}

func (ui *Ui) playSong(number int) {
	ui.songPos = 0
	var err error
	ui.songLen, err = ui.OnSelect(ui.songs[number])
	if err == nil {
		ui.state = Playing
		ui.renderSong()
		ui.renderStatus()
	}
}

// Rendering

func (ui *Ui) realign() {
	termHeight := termui.TermHeight()
	parHeight := ui.controlsPar1.Height
	ui.playList.Height = termHeight - parHeight*2
	ui.infoList.Height = termHeight - (parHeight * 2) - ui.scrollerGauge.Height - ui.volumeGauge.Height
	termui.Body.Width = termui.TermWidth()
	termui.Body.Align()
	termui.Clear()
	termui.Render(termui.Body)
}

func (ui *Ui) renderSong() {
	if ui.songSel != -1 {
		lyrics := ui.songs[ui.songSel].Lyrics()
		trackNum, _ := ui.songs[ui.songSel].Track()
		ui.infoList.Items = []string{
			"[Artist:](fg-green) " + ui.songs[ui.songSel].Artist(),
			"[Title:](fg-green)  " + ui.songs[ui.songSel].Title(),
			"[Album:](fg-green)  " + ui.songs[ui.songSel].Album(),
			fmt.Sprintf("[Track:](fg-green)  %d", trackNum),
			"[Genre:](fg-green)  " + ui.songs[ui.songSel].Genre(),
			fmt.Sprintf("[Year:](fg-green)   %d", ui.songs[ui.songSel].Year()),
		}
		if lyrics != "" {
			ui.infoList.Items = append(ui.infoList.Items, "Lyrics:  "+lyrics)
		}
	} else {
		ui.infoList.Items = []string{}
	}
	termui.Clear()
	termui.Render(termui.Body)
}

func (ui *Ui) renderStatus() {
	var status string
	switch ui.state {
	case Playing:
		status = "[(Playing)](fg-black,bg-green)"
	case Paused:
		status = "[(Paused)](fg-black,bg-yellow)"
	case Stopped:
		status = "[(Stopped)](fg-black,bg-red)"
	}
	ui.scrollerGauge.BorderLabel = status
	termui.Clear()
	termui.Render(termui.Body)
}

//Song selection

func (ui *Ui) songDown() {
	if ui.songSel < len(ui.songNames)-1 {
		ui.setSong(ui.songSel+1, true)
	}
}

func (ui *Ui) songUp() {
	if ui.songSel > 0 {
		ui.setSong(ui.songSel-1, true)
	}
}

func (ui *Ui) volumeUp() {
	if ui.volume < 100 {
		ui.volume += 5
	}
	ui.volumeGauge.Percent = ui.volume
	ui.OnVolume(ui.volume)
	termui.Clear()
	termui.Render(termui.Body)
}

func (ui *Ui) volumeDown() {
	if ui.volume > 0 {
		ui.volume -= 5
	}
	ui.volumeGauge.Percent = ui.volume
	ui.OnVolume(ui.volume)
	termui.Clear()
	termui.Render(termui.Body)
}

func (ui *Ui) setSong(num int, unset bool) {
	skip := 0
	for num-skip >= ui.playList.Height-2 {
		skip += ui.playList.Height - 2
	}
	if unset {
		ui.songNames[ui.songSel] = ui.songNames[ui.songSel][1 : len(ui.songNames[ui.songSel])-20]
	}
	ui.songSel = num
	ui.songNames[num] = fmt.Sprintf("[%s](fg-black,bg-green)", ui.songNames[num])
	ui.playList.Items = ui.songNames[skip:]
}

func (ui *Ui) reloadSongList(pathPrefix int) {
	fileList, err := getSongList(songDir)
	if err != nil {
		log.Fatal("Can't get song list")
	}
	songs := make([]Song, 0, len(fileList))

	for _, fileName := range fileList {
		currentFile, err := os.Open(fileName)
		if err == nil {
			metadata, _ := tag.ReadFrom(currentFile)
			songs = append(songs, Song{
				Metadata: metadata,
				path:     fileName,
			})
		}
		currentFile.Close()
	}
	if len(songs) == 0 {
		log.Fatal("Could find any songs to play")
	}

	ui.songs = songs
	ui.songNames = make([]string, len(ui.songs))
	ui.updateSongNames(pathPrefix)
	ui.playList.Items = ui.songNames
	ui.setSong(0, false)
}

func (ui *Ui) updateSongNames(pathPrefix int) {
	for i, v := range ui.songs {
		if v.Metadata != nil && v.Title() != "" {
			if v.Artist() != "" {
				ui.songNames[i] = fmt.Sprintf("[%d] %s - %s", i+1, v.Artist(), v.Title())
			} else {
				ui.songNames[i] = fmt.Sprintf("[%d] %s", i+1, v.Title())
			}
		} else {
			ui.songNames[i] = fmt.Sprintf("[%d] %s", i+1, v.path[pathPrefix+1:])
		}
	}
}

func (ui *Ui) previous() {
	if ui.songSel > 0 {
		ui.songUp()
		ui.playSong(ui.songSel)
		termui.Clear()
		termui.Render(termui.Body)
	}
}

func (ui *Ui) next() {
	if ui.songSel < len(ui.songNames)-1 {
		ui.songDown()
		ui.playSong(ui.songSel)
		termui.Clear()
		termui.Render(termui.Body)
	}
}

func (ui *Ui) togglePlayMode() {
	oldName := playModeName[ui.playMode]
	ui.playMode++
	if ui.playMode > Random {
		ui.playMode = LoopAll
	}
	newName := playModeName[ui.playMode]
	ui.controlsPar1.Text = strings.Replace(ui.controlsPar1.Text, oldName, newName, 1)
	termui.Clear()
	termui.Render(termui.Body)
}

func (ui *Ui) stopPlayer() {
	ui.playSong(ui.songNum)
	ui.OnPause(true)
	ui.state = Stopped
	ui.scrollerGauge.Percent = 0
	ui.scrollerGauge.Label = "0:00 / 0:00"
	ui.renderStatus()
}

func (ui *Ui) songEndEvent() {
	switch ui.playMode {
	case LoopAll:
		ui.songNum++
		if ui.songNum >= len(ui.songs) {
			ui.songNum = 0
			ui.setSong(0, true)
		} else {
			ui.songDown()
		}
		ui.playSong(ui.songNum)
	case LoopOne:
		ui.playSong(ui.songNum)
	case NoLoop:
		if ui.songNum+1 == len(ui.songs) {
			ui.stopPlayer()
		} else {
			ui.songDown()
			ui.playSong(ui.songNum)
		}
	case Random:
		ui.songNum = rand.Intn(len(ui.songs))
		ui.setSong(ui.songNum, true)
		ui.playSong(ui.songNum)
	}
}
