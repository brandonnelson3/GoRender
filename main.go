package main

import (
	"log"
	"runtime"

	"github.com/go-gl/glfw/v3.1/glfw"
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()
}
