package main

import "log"
import "github.com/go-gl/glfw"

const (
	W_WIDTH  = 720
	W_HEIGHT = 450
	W_WMODE  = glfw.Windowed // glfw.Fullscreen 
)

func initWindow() {

	if err := glfw.Init(); err != nil {
		log.Fatal(err.Error())
	}

	if err := glfw.OpenWindow(W_WIDTH, W_HEIGHT, 8, 8, 8, 8, 32, 0, W_WMODE); err == nil {
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
	glfw.Terminate()
}
