package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	numAveragedFrameLengths = 25
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

		averageFrameTime, averageFramesPerSecond := calculateFrameDetails()
		messagebus.SendSync(&messagebus.Message{System: "FrameRate", Type: "log", Data1: fmt.Sprintf("Length: %.3f ms - Avg FPS: %.1f - Limiting framerate to %d", averageFrameTime*1000, averageFramesPerSecond, framerateCap)})
	}
}

func calculateFrameDetails() (float64, float64) {
	totalTime := float64(0)
	mu.Lock()
	for _, t := range times {
		totalTime += t
	}
	mu.Unlock()
	averageFrameTime := totalTime / numAveragedFrameLengths
	averageFramesPerSecond := 1 / averageFrameTime

	return averageFrameTime, averageFramesPerSecond
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
