package main

import (
	"os"
	"os/exec"
	"runtime"
)

var currentVideoCmd *exec.Cmd

func playVideo(input Song) (int, error) {
	// Kill any existing video player process
	if currentVideoCmd != nil && currentVideoCmd.Process != nil {
		currentVideoCmd.Process.Kill()
	}

	// Detect if we have a desktop environment
	hasDesktop := hasDesktopEnvironment()

	var cmd *exec.Cmd
	var err error

	if hasDesktop {
		// Try VLC first, then fallback to ffplay
		cmd = exec.Command("vlc", input.path)
		err = cmd.Run()
		if err == nil {
			currentVideoCmd = cmd
			return 0, nil
		}

		// Fallback to ffplay
		cmd = exec.Command("ffplay", input.path)
		err = cmd.Run()
		if err == nil {
			currentVideoCmd = cmd
			return 0, nil
		}

		return 0, err
	}

	// Console/headless environment: use mpv with DISPLAY=:0
	cmd = exec.Command("mpv", input.path)
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
	err = cmd.Run()
	if err != nil {
		return 0, err
	}

	currentVideoCmd = cmd
	// Return 0 as duration since external player handles playback
	return 0, nil
}

func hasDesktopEnvironment() bool {
	// Windows always has a desktop environment
	if runtime.GOOS == "windows" {
		return true
	}

	// On Linux, check for DISPLAY (X11) or other desktop indicators
	if runtime.GOOS == "linux" {
		// Check for X11
		if os.Getenv("DISPLAY") != "" {
			return true
		}
		// Check for Wayland
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			return true
		}
	}

	// For other platforms, assume desktop environment exists
	if runtime.GOOS == "darwin" {
		return true
	}

	// Console/headless environment
	return false
}
