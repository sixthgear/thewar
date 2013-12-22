package main

import "log"
import glfw "github.com/go-gl/glfw3"

func initWindow(mon *glfw.Monitor) {

	var width, height int

	if mon != nil {
		videoMode, _ := mon.GetVideoMode()
		width = videoMode.Width
		height = videoMode.Height
	} else {
		width = W_WIDTH
		height = W_HEIGHT
	}

	if w, err := glfw.CreateWindow(width, height, "The War", mon, nil); err == nil {
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
		handleKeyDown(key, action, mods)
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

	oldWindow := window

	if currentMonitor == nil {
		currentMonitor, _ = glfw.GetPrimaryMonitor()
	} else {
		currentMonitor = nil
	}

	initWindow(currentMonitor)
	windowWidth, windowHeight := window.GetSize()
	renderer.Init(windowWidth, windowHeight)

	// window.SetShouldClose(true)
	oldWindow.Destroy()

	// initWindow()
	initCallbacks()

}
