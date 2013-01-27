package client

import "log"
import "github.com/go-gl/glfw"

const (
	W_WIDTH  = 1152
	W_HEIGHT = 720
	W_WMODE  = glfw.Windowed // glfw.Fullscreen 
)

var (
	prevMouseX, prevMouseY int = glfw.MousePos()
)

func initWindow(client *Client) {

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

func initCallbacks(client *Client) {
	glfw.SetWindowCloseCallback(func() int { client.running = false; return 0 })
	glfw.SetMousePosCallback(func(mx, my int) { client.handleMousePos(mx, my) })
	glfw.SetKeyCallback(func(key, state int) { client.handleKeyDown(key, state) })
	glfw.SetMouseWheelCallback(func(pos int) { client.handleMouseWheel(pos) })
	glfw.SetMouseButtonCallback(func(button, state int) { client.handleMouseButton(button, state) })
}

func closeWindow() {
	glfw.Terminate()
}
