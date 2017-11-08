package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	numAveragedFrameLengths = 100
	framerateCap            = 105
)

var (
	times  [numAveragedFrameLengths]float64
	frames int
	i      = 0
	mu     sync.Mutex

	previousTime    float64
	previousElapsed float64
)

func init() {
	previousTime = 0
	previousElapsed = 0

	go sendLogMessagesOnTimer()
}

func sendLogMessagesOnTimer() {
	for range time.Tick(time.Millisecond * 500) {
		if frames < numAveragedFrameLengths {
			continue
		}

		averageFramesPerSecond := calculateFrameDetails()
		messagebus.SendAsync(&messagebus.Message{Type: "console", Data1: "timer_fps", Data2: fmt.Sprintf("%f", averageFramesPerSecond)})
	}
}

func calculateFrameDetails() float64 {
	totalTime := float64(0)
	mu.Lock()
	for _, t := range times {
		totalTime += t
	}
	mu.Unlock()
	averageFrameTime := totalTime / numAveragedFrameLengths

	return averageFrameTime
}

// StartOfFrame is expected to be called at the same point in every frame to work properly.
func StartOfFrame() {
	time := glfw.GetTime()
	previousElapsed = time - previousTime
	previousTime = time
}

// EndOfFrame is intended to be called at the end of the frame with the current time. This function will block to maintain the intended maximum frame rate.
func EndOfFrame() {
	now := glfw.GetTime()
	previousElapsed := now - previousTime
	mu.Lock()
	times[i] = previousElapsed
	frames++
	i++
	if i >= numAveragedFrameLengths {
		i = 0
	}
	mu.Unlock()
	// Sleep for as long as we need to...
	time.Sleep(time.Duration((float64(time.Second) / framerateCap) - (float64(time.Second) * previousElapsed)))
}

// GetPreviousFrameLength returns the time in seconds as a float64 of the previous frame.
func GetPreviousFrameLength() float64 {
	return previousElapsed
}

// GetTime returns the current time.Now().
func GetTime() float64 {
	return glfw.GetTime()
}
