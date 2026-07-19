package proton

import (
	"fmt"
	"linux-wallpaperengine-gui/src/backend/internal/config"
	"linux-wallpaperengine-gui/src/backend/internal/logger"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	isProtonGameDetected bool
	detectorRunning      bool
	statusChangeCallback func(bool)
)

// IsProton returns whether a game is currently running under Proton
func IsProton() bool {
	return isProtonGameDetected
}

// StartDetector starts monitoring for Proton games
func StartDetector(callback func(bool)) {
	if detectorRunning {
		return
	}

	statusChangeCallback = callback
	detectorRunning = true

	go func() {
		lastStatus := false
		for detectorRunning {
			conf, err := config.GetConfig()
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}

			// If ProtonStop is disabled, we don't need to detect
			if !conf.ProtonStop {
				if isProtonGameDetected {
					isProtonGameDetected = false
					lastStatus = false
					if statusChangeCallback != nil {
						statusChangeCallback(false)
					}
				}
				time.Sleep(5 * time.Second)
				continue
			}

			// Check for Proton window
			currentStatus := checkProtonGame()

			if currentStatus != lastStatus {
				isProtonGameDetected = currentStatus
				lastStatus = currentStatus

				if currentStatus {
					logger.Printf("Proton game detected - pausing playlist and killing wallpapers")
				} else {
					logger.Printf("Proton game closed - resuming playlist and applying wallpapers")
				}

				if statusChangeCallback != nil {
					statusChangeCallback(currentStatus)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()

	logger.Println("Proton detector started")
}

func checkProtonGame() bool {
	// Search for a proton process in /proc
	procs, err := os.ReadDir("/proc")

	if err != nil {
		return false
	}

	for _, proc := range procs {
		if !proc.IsDir() {
			continue
		}
		// Search for all PIDs in /proc (will only be numeric folders)
		if _, err := strconv.Atoi(proc.Name()); err == nil {
			cmdlinePath := fmt.Sprintf("/proc/%s/cmdline", proc.Name())
			cmdlineBytes, err := os.ReadFile(cmdlinePath)
			if err == nil {
				//Check if the process' cmdline includes the text "/proton",
				//which implies the proton script is running
				if strings.Contains(string(cmdlineBytes), "/proton") {
					return true
				}
			}
		}
	}

	return false
}

// StopDetector stops the Proton detection
func StopDetector() {
	detectorRunning = false
	logger.Println("Proton detector stopped")
}
