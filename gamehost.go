package main

import glfw "github.com/go-gl/glfw3"
import gl "github.com/go-gl/gl"
import "fmt"

var gamehost_Window *glfw.Window

func RunGamehost() {
	if !glfw.Init() {
		panic("Cannot init GLFW")
	}

	defer glfw.Terminate()

	window, err := glfw.CreateWindow(800, 600, "Apollo", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	gamehost_Window = window

	gl.ClearColor(1, 1, 1, 1)
	frame := 0

	for !gamehost_Window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.LoadIdentity()
		gamehost_Window.SetTitle(fmt.Sprintf("Frame #%v", frame))
		frame++

		EvaluateString("(gameloop 0.016)", MainContext)

		graphicsQueue.Process()

		gamehost_Window.SwapBuffers()
		glfw.PollEvents()
	}
}
