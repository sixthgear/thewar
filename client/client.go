package client

import (
	// "fmt"
	"github.com/go-gl/glfw"
	"github.com/sixthgear/thewar/gamelib"
	"math"
	"math/rand"
	"time"
)

const (
	TURN_TICKS = 12
	M_WIDTH    = 64
	M_DEPTH    = 64
)

type Client struct {
	// window   *Window	
	renderer *MapRenderer
	world    *gamelib.Map

	running   bool
	headless  bool
	pathCache map[int][]int
	lastPath  int
}

func (g *Client) Init(headless bool) {

	g.pathCache = make(map[int][]int, 32)
	g.running = true
	g.headless = headless
	g.world = new(gamelib.Map)

	// rebuild vertex lists
	if !headless {
		initWindow(g)
		g.renderer = new(MapRenderer)
		g.renderer.Init()
		initCallbacks(g)
	}

	g.GenerateMap()

}

func (g *Client) GenerateMap() {

	g.world.Init(M_WIDTH, M_DEPTH)
	if !g.headless {
		g.renderer.buildVertices(g.world)
	}

	// generate random objects
	for i := 0; i < 40; i++ {

		o := new(gamelib.Obj)
		o.Team = rand.Int() % 4
		o.Type = rand.Int() % 4
		o.Facing = rand.Int() % 6
		o.OrderQueue = make([]gamelib.Order, 0)

		for {
			x, y := rand.Int()%(g.world.Width-8)+4, rand.Int()%(g.world.Width-8)+4
			hex := g.world.Lookup(x, y)
			t := uint32(0)
			switch o.Type {
			case gamelib.OBJ_INFANTRY:
				t = gamelib.T_FOREST
			case gamelib.OBJ_VEHICLE:
				t = gamelib.T_OUTDOOR
			case gamelib.OBJ_BOAT:
				t = gamelib.T_RIVER
			case gamelib.OBJ_AIRCRAFT:
				t = hex.TerrainType
			}

			if hex.TerrainType == t && hex.Unit == nil {
				g.world.Objects = append(g.world.Objects, o)
				hex.Unit = o
				o.X, o.Y = x, y
				o.Fx, o.Fy, o.Fz = g.renderer.hexCenter(g.world, hex)
				if o.Type == gamelib.OBJ_AIRCRAFT {
					o.Fy = 100
				}
				break
			}
		}

	}

	if !g.headless {
		g.renderer.buildObjects(g.world)
		g.renderer.clearPath()
	}

}

func (g *Client) HexAt(mx, my int) *gamelib.Hex {
	tx, _, tz := g.renderer.camera.WorldCoords(mx, my)
	tz = tz/HEX_HEIGHT + 0.5
	row := float64(int(tz) % 2)
	tx = tx/HEX_WIDTH + 0.5 - row*0.5
	x, z := int(tx), int(tz)
	if x < gamelib.BOUNDARY || z < gamelib.BOUNDARY || x >= g.world.Width-gamelib.BOUNDARY || z >= g.world.Depth-gamelib.BOUNDARY {
		return nil
	}
	return g.world.Lookup(x, z)
}

func (g *Client) Hover(mx, my int) {
	if g.world.Selected != nil {
		// calc path
		hex := g.HexAt(mx, my)
		if hex != nil && hex.Unit == nil {
			// is a hex and no other.Unit here
			// find a path
			i0 := g.world.Selected.Index * g.world.Width * g.world.Depth
			i1 := hex.Index

			if g.lastPath != i0+i1 {
				path, ok := g.pathCache[i0+i1]
				g.lastPath = i0 + i1
				if !ok {
					path = gamelib.FindPath(g.world, g.world.Selected, hex)
					g.pathCache[i0+i1] = path
				}
				g.renderer.buildPath(g.world, path)
			}
		} else {
			g.renderer.clearPath()
		}
	}
}

func (g *Client) Select(mx, my int) {
	hex := g.HexAt(mx, my)
	if hex == nil || hex.Unit == nil {
		// either not a hex, or no.Unit here
		g.world.Selected = nil
		g.renderer.clearPath()
	} else {
		g.world.Selected = hex
	}
}

func (g *Client) Order(mx, my int) {

	if g.world.Selected != nil {
		hex := g.HexAt(mx, my)
		if hex != nil && hex.Unit == nil {
			// is a hex and no other.Unit here
			// find a path
			path := make([]int, 0)

			path = gamelib.FindPath(g.world, g.world.Selected, hex)

			g.world.Selected.Unit.OrderQueue = append(g.world.Selected.Unit.OrderQueue, gamelib.Order{gamelib.OR_MOVE, path})
			g.world.Selected.Unit = nil
			g.world.Selected = nil
			g.renderer.clearPath()
		} else {
			// invalid order
			// do nothing
		}
	}
}

func (g *Client) Update(dt float64) {

	if kf.scrollUp {
		g.renderer.camera.z -= (g.renderer.camera.y * 0.02)
	} else if kf.scrollDown {
		g.renderer.camera.z += (g.renderer.camera.y * 0.02)
	}
	if kf.scrollLeft {
		g.renderer.camera.x -= (g.renderer.camera.y * 0.02)
	} else if kf.scrollRight {
		g.renderer.camera.x += (g.renderer.camera.y * 0.02)
	}
	if kf.zoomOut {
		g.renderer.camera.y += 10
	} else if kf.zoomIn {
		g.renderer.camera.y -= 10
	}
	if kf.tiltUp {
		g.renderer.camera.rx -= 1
	} else if kf.tiltDown {
		g.renderer.camera.rx += 1
	}

	for i := range g.world.Objects {

		obj := g.world.Objects[i]
		if len(obj.OrderQueue) > 0 {

			order := &obj.OrderQueue[0]
			if obj.Dest != nil {
				// look at order at front of queue
				a := g.world.Lookup(obj.X, obj.Y)
				b := obj.Dest

				// next := g.world.Index(order.Path[len(order.Path)-1])
				// if theres at least two more nodes in the path
				// and the node two turns from now has the same x
				// and the next node has the same terrain type
				// interpolate obj.fx,fy,fz
				// from prev to next
				obj.Facing = g.world.Direction(a, b)
				x0, y0, z0 := g.renderer.hexCenter(g.world, a)
				x1, y1, z1 := g.renderer.hexCenter(g.world, b)

				if obj.Type == gamelib.OBJ_AIRCRAFT {
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
				if obj.Type == gamelib.OBJ_AIRCRAFT {
					t = float32(math.Min(1.0, float64(obj.AnimCounter)/float64(obj.AnimTotal)))
				} else {
					t = float32(math.Min(1.0, float64(obj.AnimCounter)/TURN_TICKS)) // float64(obj.AnimTotal)		
				}

				ts := t * t
				tc := ts * t

				switch {
				case obj.Type == gamelib.OBJ_AIRCRAFT:
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

				g.renderer.buildObjects(g.world)
				obj.AnimCounter++
			}
			if obj.Dest == nil || obj.AnimCounter >= obj.AnimTotal {

				// if obj.Dest == nil {
				// 	order.Path = order.Path[0 : len(order.Path)-1] // pop				
				// }

				newHex := g.world.Index(order.Path[len(order.Path)-1])
				order.Path = order.Path[0 : len(order.Path)-1] // pop
				obj.X = newHex.Index % g.world.Width
				obj.Y = newHex.Index / g.world.Width

				if len(order.Path) == 0 {
					// remove order
					obj.Dest = nil
					obj.OrderQueue = obj.OrderQueue[0 : len(obj.OrderQueue)-1]
					newHex.Unit = g.world.Objects[i]
					continue
				} else {
					obj.Dest = g.world.Index(order.Path[len(order.Path)-1])
					cost := 1 - gamelib.TMOD[obj.Type][obj.Dest.TerrainType].MOV
					obj.AnimCounter = 0
					if obj.Type == gamelib.OBJ_AIRCRAFT {

						x := float64(obj.Dest.Index%g.world.Width - obj.X)
						y := float64(obj.Dest.Index/g.world.Width - obj.Y)
						h := int(TURN_TICKS*math.Hypot(y, x)) / 2

						obj.AnimTotal = h
					} else {
						obj.AnimTotal = TURN_TICKS * cost
					}
				}
			}

		}
	}
}

func (g *Client) Run() {

	t := 0.0
	const dt = 1.0 / 60
	currentTime := float64(time.Now().UnixNano()) / 1000000000
	accumulator := 0.0

	for g.running {

		newTime := float64(time.Now().UnixNano()) / 1000000000
		frameTime := newTime - currentTime
		currentTime = newTime
		accumulator += frameTime

		for accumulator >= dt {

			if !g.headless {
				kf.PollInput()
			}
			g.Update(dt)
			accumulator -= dt
			t += dt
		}
		if !g.headless {
			g.Render()
		}
	}

	closeWindow()
}

func (g *Client) Render() {
	g.renderer.Render(g.world)
}

func (g *Client) handleKeyDown(key, state int) {
	switch {
	case key == 'R' && state == 1:
		g.GenerateMap()
	case key == glfw.KeyEsc && state == 1:
		g.running = false
	}
}

func (g *Client) handleMousePos(mx, my int) {
	deltaX, deltaY := float64(mx-prevMouseX), float64(my-prevMouseY)
	prevMouseX = mx
	prevMouseY = my
	if glfw.MouseButton(glfw.MouseMiddle) == 1 {
		g.renderer.camera.x -= deltaX * (g.renderer.camera.y * 0.0015)
		g.renderer.camera.z -= deltaY * (g.renderer.camera.y * 0.0015)
	}
	g.Hover(mx, my)
}

func (g *Client) handleMouseButton(button, state int) {
	// fmt.Printf("button '%d' -> %d\n", button, state)
	mx, my := glfw.MousePos()
	switch {
	case button == glfw.MouseLeft && state == 1:
		g.Select(mx, my)
	case button == glfw.MouseRight && state == 1:
		g.Order(mx, my)
	}
}

func (g *Client) handleMouseWheel(pos int) {
	// fmt.Println("hello!", delta)
	// println(delta)
	g.renderer.camera.y = 1200 - float64(pos)*25
}
