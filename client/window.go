package main

import "log"
import "github.com/go-gl/glfw"

const (
	W_WIDTH  = 1152
	W_HEIGHT = 720
	W_WMODE  = glfw.Windowed // glfw.Fullscreen 
)

var (
	currentWindowMode = W_WMODE
)

func initWindow() {

	if err := glfw.Init(); err != nil {
		log.Fatal(err.Error())
	}

	if err := glfw.OpenWindow(W_WIDTH, W_HEIGHT, 8, 8, 8, 8, 32, 0, currentWindowMode); err == nil {
		glfw.SetWindowTitle("The War")
		glfw.SetSwapInterval(1)
		glfw.Enable(glfw.MouseCursor)
	} else {
		glfw.Terminate()
		log.Fatal(err.Error())
	}

}

func initCallbacks() {
	glfw.SetWindowCloseCallback(func() int { running = false; return 0 })
	glfw.SetMousePosCallback(func(mx, my int) { handleMousePos(mx, my) })
	glfw.SetKeyCallback(func(key, state int) { handleKeyDown(key, state) })
	glfw.SetMouseWheelCallback(func(pos int) { handleMouseWheel(pos) })
	glfw.SetMouseButtonCallback(func(button, state int) { handleMouseButton(button, state) })
}

func closeWindow() {
	glfw.CloseWindow()
	glfw.Terminate()
}

func toggleFullScreen() {

	if currentWindowMode == glfw.Windowed {
		currentWindowMode = glfw.Fullscreen
	} else {
		currentWindowMode = glfw.Windowed
	}

	glfw.CloseWindow()
	glfw.OpenWindow(W_WIDTH, W_HEIGHT, 8, 8, 8, 8, 32, 0, currentWindowMode)

	initCallbacks()

}
