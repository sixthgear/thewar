package main

import (
	"bufio"
	"fmt"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	. "github.com/sixthgear/thewar/gamelib"
	"log"
	"math"
	"net"
	"runtime"
	"time"
)

const (
	DEBUG = true
)

var (
	window     *glfw.Window
	world      *Map
	renderer   *MapRenderer
	running    bool
	pathCache  map[int][]int
	lastPath   int
	mx, my     int
	conn       net.Conn
	fonts      map[string]*Font
	timer      int
	timerLabel *TextLabel
	roundLabel *TextLabel
	unitLabel  *TextLabel
	user       string
)

func main() {

	runtime.LockOSThread()

	var err error

	// set up GLFW
	if !glfw.Init() {
		log.Fatal("Could not initialize GLFW!")
	}
	defer glfw.Terminate()
	glfw.SwapInterval(1)

	// set up window
	initWindow()

	user = "sixthgear"
	addressList := []string{
		"0.0.0.0:11235",
		"ironman.quitjobmakegames.com:11235",
		"64.46.1.232:11235",
	}

	// initiate connection
	for a := range addressList {
		if conn, err = net.Dial("tcp", addressList[a]); err != nil && DEBUG {
			log.Printf("Could not connect to %s.\n", addressList[a])
			continue
		}
		break
	}

	if conn == nil {
		log.Fatalln("No available servers!")
	}

	// create world
	world, _ = new(Map).Decode(reqMap())
	pathCache = make(map[int][]int, 32)

	// set up renderer
	renderer = new(MapRenderer).Init()
	renderer.buildVertices(world)
	renderer.buildObjects(world)
	renderer.clearPath()

	// load fonts
	fonts = make(map[string]*Font, 5)
	fonts["rockwell24"], _ = new(Font).Load("rockwell24")
	fonts["rockwell36"], _ = new(Font).Load("rockwell36")

	// create UI
	roundLabel = new(TextLabel).Init("ROUND 1", fonts["rockwell36"], 4, 4)
	timerLabel = new(TextLabel).Init("TIME 2:00", fonts["rockwell24"], 4, 40)
	unitLabel = new(TextLabel).Init(".", fonts["rockwell24"], 4, W_HEIGHT-28)

	// timer
	timer = 120

	// set up interface callbacks
	initCallbacks()
	run()
}

func hexAt(mx, my int) *Hex {
	tx, _, tz := renderer.camera.WorldCoords(mx, my)
	tz = tz/HEX_HEIGHT + 0.5
	row := float64(int(tz) % 2)
	tx = tx/HEX_WIDTH + 0.5 - row*0.5
	x, z := int(tx), int(tz)
	if x < BOUNDARY || z < BOUNDARY || x >= world.Width-BOUNDARY || z >= world.Depth-BOUNDARY {
		return nil
	}
	return world.Lookup(x, z)
}

func doHover(mx, my int) {

	if world.Selected != nil {
		// calc path
		hex := hexAt(mx, my)
		if hex != nil && hex.Unit == nil {
			// is a hex and no other.Unit here
			// find a path
			i0 := world.Selected.Index * world.Width * world.Depth
			i1 := hex.Index

			if lastPath != i0+i1 {
				path, ok := pathCache[i0+i1]
				lastPath = i0 + i1
				if !ok {
					path = FindPath(world, world.Selected, hex)
					pathCache[i0+i1] = path
				}
				renderer.buildPath(world, path)
			}
		} else {
			renderer.clearPath()
		}
	}
}

func doSelect(mx, my int) {
	hex := hexAt(mx, my)
	if hex == nil || hex.Unit == nil {
		// either not a hex, or no unit here
		world.Selected = nil
		renderer.clearPath()
		unitLabel.SetText(".")
	} else {
		world.Selected = hex
		switch hex.Unit.Type {
		case OBJ_INFANTRY:
			unitLabel.SetText("INFANTRY")
		case OBJ_VEHICLE:
			unitLabel.SetText("VEHICLE")
		case OBJ_BOAT:
			unitLabel.SetText("BOAT")
		case OBJ_AIRCRAFT:
			unitLabel.SetText("AIRCRAFT")
		}
	}
}

func reqMap() []byte {
	o := Order{OR_REQMAP, 0, nil}
	conn.Write(o.Encode())
	data, _ := bufio.NewReader(conn).ReadBytes('\n')
	// world.Decode(data)
	return data
}

func reqOrder(mx, my int) {
	if world.Selected != nil {
		hex := hexAt(mx, my)
		if hex != nil && hex.Unit == nil {
			// is a hex and no other.Unit here
			// find a path
			unit := world.Selected.Unit
			path := FindPath(world, world.Selected, hex)

			o := Order{OR_MOVE, unit.Id, path[0 : len(path)-1]}
			conn.Write(o.Encode())

			world.Selected = nil
			renderer.clearPath()
			unitLabel.SetText(".")
		} else {
			// invalid order
			// do nothing
		}
	}
}

func handleOrder(o Order) {
	switch o.Order {
	case OR_MOVE:
		// log.Println("MOVE")
		doMove(o)
	default:
		if DEBUG {
			log.Println("Unhandled order: ", o.Order)
		}

	}
}

func doMove(o Order) {
	obj := world.Objects[o.UnitId]
	if world.Selected != nil && world.Selected.Unit == obj {
		world.Selected = nil
		renderer.clearPath()
	}
	world.Lookup(obj.X, obj.Y).Unit = nil
	obj.OrderQueue = append(obj.OrderQueue, o)
	obj.NextDest(world)
}

func update(dt float64) {

	if kf.scrollUp {
		renderer.camera.z -= (renderer.camera.y * 0.02)
	} else if kf.scrollDown {
		renderer.camera.z += (renderer.camera.y * 0.02)
	}
	if kf.scrollLeft {
		renderer.camera.x -= (renderer.camera.y * 0.02)
	} else if kf.scrollRight {
		renderer.camera.x += (renderer.camera.y * 0.02)
	}
	if kf.zoomOut {
		renderer.camera.y += 10
	} else if kf.zoomIn {
		renderer.camera.y -= 10
	}
	if kf.tiltUp {
		renderer.camera.rx -= 1
	} else if kf.tiltDown {
		renderer.camera.rx += 1
	}
	if mx < 10 {
		renderer.camera.x -= 10 * (renderer.camera.y * 0.0015)
	} else if mx > W_WIDTH-10 {
		renderer.camera.x += 10 * (renderer.camera.y * 0.0015)
	}
	if my < 10 {
		renderer.camera.z -= 10 * (renderer.camera.y * 0.0015)
	} else if my > W_HEIGHT-10 {
		renderer.camera.z += 10 * (renderer.camera.y * 0.0015)
	}

	for i := range world.Objects {
		animate(world.Objects[i])
	}
}

func animate(obj *Obj) {

	if len(obj.OrderQueue) > 0 {

		order := &obj.OrderQueue[0]

		if obj.Dest != nil {
			obj.AnimCounter++

			// look at order at front of queue
			a := world.Lookup(obj.X, obj.Y)
			b := obj.Dest

			obj.Facing = world.Direction(a, b)
			x0, y0, z0 := world.HexCenter(a)
			x1, y1, z1 := world.HexCenter(b)

			if obj.Type == OBJ_AIRCRAFT {
				//airplanes fly
				y0, y1 = 100, 100
			}

			if a.Unit != nil {
				y0 += 16
			}
			if b.Unit != nil {
				y1 += 16
			}

			t := float32(0)
			if obj.Type == OBJ_AIRCRAFT {
				t = float32(math.Min(1.0, float64(obj.AnimCounter)/float64(obj.AnimTotal)))
			} else {
				t = float32(math.Min(1.0, float64(obj.AnimCounter)/TURN_TICKS)) // float64(obj.AnimTotal)
			}

			switch {
			case obj.Type == OBJ_AIRCRAFT:
				obj.Fx = x0 + (x1-x0)*t
				obj.Fz = z0 + (z1-z0)*t
			case y1 > y0:
				ts := t * t
				tc := ts * t
				obj.Fx = x0 + (x1-x0)*ts
				obj.Fz = z0 + (z1-z0)*ts
				obj.Fy = y0 + (y1-y0)*(-2*tc*ts+-0.0025*ts*ts+10*tc+-15*ts+8*t)
			case y1 < y0:
				ts := t * t
				tc := ts * t
				obj.Fx = x0 + (x1-x0)*ts
				obj.Fz = z0 + (z1-z0)*ts
				obj.Fy = y0 + (y1-y0)*(2*ts*ts+2*tc+-3*ts)
			default:
				ts := t * t
				obj.Fx = x0 + (x1-x0)*ts
				obj.Fz = z0 + (z1-z0)*ts
			}

			renderer.buildObjects(world)
			if obj.AnimCounter >= obj.AnimTotal {

				newHex := world.Index(order.Path[len(order.Path)-1])
				order.Path = order.Path[0 : len(order.Path)-1] // pop
				obj.X = newHex.Index % world.Width
				obj.Y = newHex.Index / world.Width
				if len(order.Path) == 0 {
					// remove order
					obj.Dest = nil
					obj.OrderQueue = obj.OrderQueue[0 : len(obj.OrderQueue)-1]
					newHex.Unit = obj
					return
				} else {
					obj.NextDest(world)
				}
			}
		}
	}
}

func communicate(channel chan *Order) {

	for {
		data, _ := bufio.NewReader(conn).ReadBytes('\n')
		if len(data) == 0 {
			if DEBUG {
				log.Println("Server closed connection!")
			}
			conn.Close()
			return
		}
		order, _ := new(Order).Decode(data)
		if DEBUG {
			log.Printf("%s", data)
		}
		channel <- order
	}

}

func run() {

	const dt = 1.0 / 60

	// set up network IO loop
	channel := make(chan *Order)
	go communicate(channel)

	t := 0.0
	currentTime := float64(time.Now().UnixNano()) / 1000000000
	accumulator := 0.0
	running = true

	for running {

		// animation loop ala Gaffer
		newTime := float64(time.Now().UnixNano()) / 1000000000
		frameTime := newTime - currentTime
		currentTime = newTime
		accumulator += frameTime

		for accumulator >= dt {

			kf.PollInput()

			var o *Order
			select {
			case o = <-channel:
				handleOrder(*o)
			default:
				// nothing to do
			}

			update(dt)
			accumulator -= dt
			t += dt
		}

		timerFloat := math.Max(0, 120-t)
		if int(timerFloat) != timer {
			timer = int(timerFloat)
			timerLabel.SetText(fmt.Sprintf("TIME %d:%02d", timer/60, timer%60))
			if timer <= 10 {
				timerLabel.Color = [3]float32{1, 0.2, 0.2}
			}
		}

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.LoadIdentity()
		renderer.Render(world)
		timerLabel.Render(renderer.camera)
		roundLabel.Render(renderer.camera)
		unitLabel.Render(renderer.camera)

		window.SwapBuffers()
		glfw.PollEvents()
	}

	closeWindow()
}

func handleKeyDown(key glfw.Key, action glfw.Action) {
	switch {
	case key == glfw.KeyR && action == glfw.Press:
		//
	case key == glfw.KeyF && action == glfw.Press:
		//
	case key == glfw.KeyEscape && action == glfw.Press:
		running = false
	}
}

func handleMousePos(nx, ny float64) {
	dx, dy := nx-float64(mx), ny-float64(my)
	mx, my = int(nx), int(ny)

	if window.GetMouseButton(glfw.MouseButtonMiddle) == glfw.Press {
		renderer.camera.x -= dx * (renderer.camera.y * 0.0015)
		renderer.camera.z -= dy * (renderer.camera.y * 0.0015)
	}
	doHover(mx, my)
}

func handleMouseButton(button glfw.MouseButton, action glfw.Action) {

	// fx, fy := window.GetCursorPosition()
	switch {
	case button == glfw.MouseButtonLeft && action == glfw.Press:
		doSelect(mx, my)
	case button == glfw.MouseButtonRight && action == glfw.Press:
		reqOrder(mx, my)
	}
}

func handleMouseWheel(pos float64) {

	renderer.camera.y += pos * -25
}
