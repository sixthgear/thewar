package client

import "github.com/go-gl/glfw"

var kf KeyFlags

type KeyFlags struct {
	scrollLeft  bool
	scrollRight bool
	scrollUp    bool
	scrollDown  bool
	zoomOut     bool
	zoomIn      bool
	tiltUp      bool
	tiltDown    bool
}

var keyBindings = map[int]*bool{
	'W':            &kf.scrollUp,
	'S':            &kf.scrollDown,
	'A':            &kf.scrollLeft,
	'D':            &kf.scrollRight,
	'Z':            &kf.tiltUp,
	'X':            &kf.tiltDown,
	glfw.KeySpace:  &kf.zoomOut,
	glfw.KeyLshift: &kf.zoomIn,
	// glfw.KeyLeft:  &kf.left,
	// glfw.KeyRight: &kf.left,
	// glfw.KeyUp:    &kf.left,
	// glfw.KeyDown:  &kf.left,
}

func (k *KeyFlags) PollInput() {
	for k, v := range keyBindings {
		*v = glfw.Key(k) == glfw.KeyPress
	}
}
