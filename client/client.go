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
	M_WIDTH = 64
	M_DEPTH = 64
)

var (
	world                  *Map
	renderer               *MapRenderer
	running                bool
	pathCache              map[int][]int
	lastPath               int
	prevMouseX, prevMouseY int = glfw.MousePos()
	conn                   net.Conn
	channel                chan *Order = make(chan *Order)
)

func main() {

	var err error

	port := flag.Int("port", 11235, "Port to listen on.")
	ip := flag.String("ip", "0.0.0.0", "Address to connect to.")
	flag.Parse()

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

	// world.Init(M_WIDTH, M_DEPTH)
	world, _ = new(Map).Decode(reqMap())
	renderer.buildVertices(world)

	// reconect object references
	for i := range world.Objects {
		o := world.Objects[i]
		world.Lookup(o.X, o.Y).Unit = o
	}
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
		// either not a hex, or no unit here		
		world.Selected = nil
		renderer.clearPath()
	} else {
		// data, _ := json.MarshalIndent(hex.Unit, "", "\t")
		// fmt.Printf("%s\n", data)
		world.Selected = hex
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
		log.Println("Unhandled order: ", o.Order)
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
			log.Println("Server closed connection!")
			conn.Close()
			return
		}
		order, _ := new(Order).Decode(data)
		log.Printf("%s", data)
		channel <- order
	}

}

func run() {

	t := 0.0
	const dt = 1.0 / 60
	currentTime := float64(time.Now().UnixNano()) / 1000000000
	accumulator := 0.0

	go communicate(channel)

	for running {

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

		renderer.Render(world)

	}

	closeWindow()
}

func handleKeyDown(key, state int) {
	switch {
	case key == 'R' && state == 1:
		// world.Generate()
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
		reqOrder(mx, my)
	}
}

func handleMouseWheel(pos int) {
	// fmt.Println("hello!", delta)
	// println(delta)
	renderer.camera.y = 1200 - float64(pos)*25
}
