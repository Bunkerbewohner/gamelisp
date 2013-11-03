package main

import glfw "github.com/go-gl/glfw3"
import gl "github.com/go-gl/gl"

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

	for !gamehost_Window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.LoadIdentity()

		gamehost_Window.SwapBuffers()
		glfw.PollEvents()
	}
}
