package main

import "log"
import glfw "github.com/go-gl/glfw3"

const (
	W_WIDTH  = 1152
	W_HEIGHT = 720
)

var (
	currentMonitor *glfw.Monitor = nil // glfw.Windowed // glfw.Fullscreen
)

func initWindow() {
	if w, err := glfw.CreateWindow(W_WIDTH, W_HEIGHT, "The War", currentMonitor, nil); err == nil {
		window = w
		window.MakeContextCurrent()
		window.SetInputMode(glfw.Cursor, glfw.CursorNormal)
	} else {
		glfw.Terminate()
		log.Fatal(err.Error())
	}

}

func initCallbacks() {
	window.SetCloseCallback(func(w *glfw.Window) {
		running = false
	})
	window.SetCursorPositionCallback(func(w *glfw.Window, mx float64, my float64) {
		handleMousePos(mx, my)
	})
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		handleKeyDown(key, action)
	})
	window.SetScrollCallback(func(w *glfw.Window, xoff float64, yoff float64) {
		handleMouseWheel(yoff)
	})
	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		handleMouseButton(button, action)
	})
}

func closeWindow() {
	window.SetShouldClose(true)
	// glfw.CloseWindow()
}

func toggleFullScreen() {

	if currentMonitor == nil {
		currentMonitor, _ = glfw.GetPrimaryMonitor()
	} else {
		currentMonitor = nil
	}

	window.Destroy()
	initWindow()
	initCallbacks()

}
