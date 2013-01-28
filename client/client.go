package main

import (
	"bufio"
	// "encoding/json"
	"flag"
	"fmt"
	"github.com/go-gl/glfw"
	. "github.com/sixthgear/thewar/gamelib"
	"log"
	"math"
	"net"
	"time"
)

const (
	TURN_TICKS = 12
	M_WIDTH    = 64
	M_DEPTH    = 64
)

var (
	world                  *Map
	renderer               *MapRenderer
	running                bool
	pathCache              map[int][]int
	lastPath               int
	prevMouseX, prevMouseY int = glfw.MousePos()
	conn                   net.Conn
)

func main() {

	port := flag.Int("port", 11235, "Port to listen on.")
	ip := flag.String("ip", "0.0.0.0", "Address to connect to.")
	flag.Parse()

	// connect to server
	// download map
	var err error
	conn, err = net.Dial("tcp", *ip+":"+fmt.Sprintf("%d", *port))
	if err != nil {
		log.Fatal("Could not connect. \n", err)
	}

	initGame()
	run()
}

func initGame() {

	pathCache = make(map[int][]int, 32)
	running = true
	world = new(Map)

	initWindow()
	renderer = new(MapRenderer)
	renderer.Init()
	initCallbacks()

	world.Init(M_WIDTH, M_DEPTH)
	// world.Generate()

	data, _ := bufio.NewReader(conn).ReadBytes('\n')
	world.Decode(data)

	for i := range world.Objects {
		o := world.Objects[i]
		world.Lookup(o.X, o.Y).Unit = o
	}

	renderer.buildVertices(world)
	// GenerateObjects(world)
	renderer.buildObjects(world)
	renderer.clearPath()

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
		// either not a hex, or no.Unit here		
		world.Selected = nil
		renderer.clearPath()
	} else {
		// data, _ := json.MarshalIndent(hex.Unit, "", "\t")
		// fmt.Printf("%s\n", data)
		world.Selected = hex
	}
}

func doOrder(mx, my int) {

	if world.Selected != nil {
		hex := hexAt(mx, my)
		if hex != nil && hex.Unit == nil {
			// is a hex and no other.Unit here
			// find a path
			path := make([]int, 0)

			path = FindPath(world, world.Selected, hex)

			world.Selected.Unit.OrderQueue = append(world.Selected.Unit.OrderQueue, Order{OR_MOVE, path})
			world.Selected.Unit = nil
			world.Selected = nil
			renderer.clearPath()
		} else {
			// invalid order
			// do nothing
		}
	}
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

	for i := range world.Objects {
		animate(world.Objects[i])
	}
}

func animate(obj *Obj) {

	if len(obj.OrderQueue) > 0 {

		order := &obj.OrderQueue[0]
		if obj.Dest != nil {
			// look at order at front of queue
			a := world.Lookup(obj.X, obj.Y)
			b := obj.Dest

			// next := world.Index(order.Path[len(order.Path)-1])
			// if theres at least two more nodes in the path
			// and the node two turns from now has the same x
			// and the next node has the same terrain type
			// interpolate obj.fx,fy,fz
			// from prev to next
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

			ts := t * t
			tc := ts * t

			switch {
			case obj.Type == OBJ_AIRCRAFT:
				obj.Fx = x0 + (x1-x0)*t
				obj.Fz = z0 + (z1-z0)*t
			case y1 > y0:
				obj.Fx = x0 + (x1-x0)*ts
				obj.Fz = z0 + (z1-z0)*ts
				obj.Fy = y0 + (y1-y0)*(-2*tc*ts+-0.0025*ts*ts+10*tc+-15*ts+8*t)
			case y1 < y0:
				obj.Fx = x0 + (x1-x0)*ts
				obj.Fz = z0 + (z1-z0)*ts
				obj.Fy = y0 + (y1-y0)*(2*ts*ts+2*tc+-3*ts)
			default:
				obj.Fx = x0 + (x1-x0)*ts
				obj.Fz = z0 + (z1-z0)*ts
			}

			renderer.buildObjects(world)
			obj.AnimCounter++
		}
		if obj.Dest == nil || obj.AnimCounter >= obj.AnimTotal {

			// if obj.Dest == nil {
			// 	order.Path = order.Path[0 : len(order.Path)-1] // pop				
			// }

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
				obj.Dest = world.Index(order.Path[len(order.Path)-1])
				cost := 1 - TMOD[obj.Type][obj.Dest.TerrainType].MOV
				obj.AnimCounter = 0
				if obj.Type == OBJ_AIRCRAFT {

					x := float64(obj.Dest.Index%world.Width - obj.X)
					y := float64(obj.Dest.Index/world.Width - obj.Y)
					h := int(TURN_TICKS*math.Hypot(y, x)) / 2

					obj.AnimTotal = h
				} else {
					obj.AnimTotal = TURN_TICKS * cost
				}
			}
		}

	}
}

func run() {

	t := 0.0
	const dt = 1.0 / 60
	currentTime := float64(time.Now().UnixNano()) / 1000000000
	accumulator := 0.0

	for running {

		newTime := float64(time.Now().UnixNano()) / 1000000000
		frameTime := newTime - currentTime
		currentTime = newTime
		accumulator += frameTime

		for accumulator >= dt {

			kf.PollInput()

			update(dt)
			accumulator -= dt
			t += dt
		}

		renderer.Render(world)

	}

	closeWindow()
}

func handleKeyDown(key, state int) {
	switch {
	case key == 'R' && state == 1:
		world.Generate()
	case key == glfw.KeyEsc && state == 1:
		running = false
	}
}

func handleMousePos(mx, my int) {
	deltaX, deltaY := float64(mx-prevMouseX), float64(my-prevMouseY)
	prevMouseX = mx
	prevMouseY = my
	if glfw.MouseButton(glfw.MouseMiddle) == 1 {
		renderer.camera.x -= deltaX * (renderer.camera.y * 0.0015)
		renderer.camera.z -= deltaY * (renderer.camera.y * 0.0015)
	}
	doHover(mx, my)
}

func handleMouseButton(button, state int) {
	// fmt.Printf("button '%d' -> %d\n", button, state)
	mx, my := glfw.MousePos()
	switch {
	case button == glfw.MouseLeft && state == 1:
		doSelect(mx, my)
	case button == glfw.MouseRight && state == 1:
		doOrder(mx, my)
	}
}

func handleMouseWheel(pos int) {
	// fmt.Println("hello!", delta)
	// println(delta)
	renderer.camera.y = 1200 - float64(pos)*25
}
