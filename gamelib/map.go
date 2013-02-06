package gamelib

import (
	// "archive/zip"
	// "fmt"
	// "log"
	"encoding/json"
	"github.com/sixthgear/noise"
	"math"
	"math/rand"
	"time"
)

const (
	BOUNDARY   = 4 // number of hexes to mark as unplayable around the edges
	HEX_SIZE   = 120
	HEX_WIDTH  = HEX_SIZE * 0.866025403784439
	HEX_HEIGHT = HEX_SIZE * 0.75
	HEX_GAP    = 0
)

const (
	T_OUTDOOR = iota
	T_INDOOR
	T_ROAD
	T_FOREST
	T_FIELD
	T_RIVER
	T_BOG
	T_HILL
	T_BEACH
	T_BOUNDS
)

var TMOD = [...][10]ObjStats{
	OBJ_INFANTRY: [...]ObjStats{
		T_OUTDOOR: ObjStats{MOV: -1},
		T_INDOOR:  ObjStats{STH: +2, DEF: +2},
		T_ROAD:    ObjStats{MOV: +2},
		T_FOREST:  ObjStats{MOV: -2, VIS: -2, STH: +2, DEF: +2},
		T_FIELD:   ObjStats{STH: +2},
		T_RIVER:   ObjStats{MOV: -6, STH: -2},
		T_BEACH:   ObjStats{MOV: -2, STH: -2},
		T_BOG:     ObjStats{MOV: -2, STH: -2},
		T_HILL:    ObjStats{MOV: -3, RAN: +4, VIS: +4, STH: -2, DEF: +1},
		T_BOUNDS:  ObjStats{MOV: IMPOSSIBLE, ATT: IMPOSSIBLE},
	},
	OBJ_VEHICLE: [...]ObjStats{
		T_OUTDOOR: ObjStats{},
		T_INDOOR:  ObjStats{MOV: IMPOSSIBLE},
		T_ROAD:    ObjStats{MOV: +4},
		T_FOREST:  ObjStats{MOV: -3, VIS: -2, STH: +2, DEF: +2},
		T_FIELD:   ObjStats{STH: +2},
		T_RIVER:   ObjStats{MOV: IMPOSSIBLE},
		T_BEACH:   ObjStats{MOV: -1, STH: -2},
		T_BOG:     ObjStats{MOV: -4, STH: -2},
		T_HILL:    ObjStats{MOV: IMPOSSIBLE},
		T_BOUNDS:  ObjStats{MOV: IMPOSSIBLE, ATT: IMPOSSIBLE},
	},
	OBJ_BOAT: [...]ObjStats{
		T_OUTDOOR: ObjStats{MOV: IMPOSSIBLE},
		T_INDOOR:  ObjStats{STH: +2, DEF: +2},
		T_ROAD:    ObjStats{MOV: +2},
		T_FOREST:  ObjStats{MOV: IMPOSSIBLE, VIS: -2, STH: +2, DEF: +2},
		T_FIELD:   ObjStats{STH: +2},
		T_RIVER:   ObjStats{MOV: -1, STH: -2},
		T_BEACH:   ObjStats{MOV: IMPOSSIBLE, STH: -2},
		T_BOG:     ObjStats{MOV: -2, STH: -2},
		T_HILL:    ObjStats{MOV: IMPOSSIBLE, RAN: +4, VIS: +4, STH: -2, DEF: +1},
		T_BOUNDS:  ObjStats{MOV: IMPOSSIBLE, ATT: IMPOSSIBLE},
	},
	OBJ_AIRCRAFT: [...]ObjStats{
		T_OUTDOOR: ObjStats{},
		T_INDOOR:  ObjStats{},
		T_ROAD:    ObjStats{},
		T_FOREST:  ObjStats{},
		T_FIELD:   ObjStats{},
		T_RIVER:   ObjStats{},
		T_BEACH:   ObjStats{},
		T_BOG:     ObjStats{},
		T_HILL:    ObjStats{},
		T_BOUNDS:  ObjStats{},
	},
}

type Map struct {
	Width, Depth int
	Seed         int64
	Structures   []*Struct
	Objects      []*Obj
	Selected     *Hex

	Grid []Hex
}

type Hex struct {
	Index       int
	Height      float32
	Color       [3]float32
	TerrainType uint32
	Unit        *Obj `json:"-"`
}

func (m *Map) Init(width, depth int) {
	m.Width = width
	m.Depth = depth
	m.Grid = make([]Hex, width*depth)
	m.Objects = make([]*Obj, 0)
	m.Selected = nil
}

func (m *Map) Index(i int) *Hex {
	return &m.Grid[i]
}
func (m *Map) Lookup(x, y int) *Hex {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Depth {
		return nil
	}
	return &m.Grid[y*m.Width+x]
}

func (m *Map) Generate() {

	m.Init(m.Width, m.Depth)
	m.Seed = time.Now().UTC().UnixNano()
	// fmt.Println("seed:", m.Seed)
	rand.Seed(m.Seed)

	nx := rand.Float64() * 320000
	nz := rand.Float64() * 320000

	// generate height map and terrain types
	for i := 0; i < m.Width*m.Depth; i++ {
		hex := m.Index(i)
		hex.Index = i
		x := i % m.Width
		z := i / m.Width
		row := float64(z % 2)
		hex.Height = float32(noise.OctaveNoise2d(nx+float64(x)+row/2, nz+float64(z), 4, 0.25, 1.0/24))
		hex.Height = (hex.Height + 1.0) * 0.5

		switch {
		case x < BOUNDARY, x >= m.Width-BOUNDARY, z < BOUNDARY, z >= m.Depth-BOUNDARY:
			hex.TerrainType = T_BOUNDS
			v := float32(math.Sqrt(float64(hex.Height))/4) - 0.05
			hex.Color = [3]float32{v, v, v}
			hex.Height = 0.0 * 32
		case hex.Height < 0.375:
			hex.TerrainType = T_RIVER
			hex.Color = [3]float32{0.1, 0.1, hex.Height*1.5 + 0.2}
			hex.Height = 0.125 * 32
		case hex.Height < 0.5:
			hex.TerrainType = T_BEACH
			hex.Color = [3]float32{hex.Height * 1.6, hex.Height * 1.6, 0.6}
			hex.Height = 0.125 * 32
		case hex.Height < 0.625:
			hex.TerrainType = T_OUTDOOR
			hex.Color = [3]float32{0.1, hex.Height, 0.1}
			hex.Height = 0.375 * 32
		case hex.Height < 0.75:
			hex.TerrainType = T_FOREST
			hex.Color = [3]float32{0, hex.Height / 3, 0}
			hex.Height = 0.625 * 32
		default:
			hex.TerrainType = T_HILL
			v := float32(hex.Height*hex.Height) - 0.3
			hex.Color = [3]float32{v, v, v}
			hex.Height = 1.875 * 32
		}
	}
}

func (m *Map) Direction(a, b *Hex) (dist int) {
	dir := 0
	ax, az := a.Index%m.Width, a.Index/m.Width
	bx, bz := b.Index%m.Width, b.Index/m.Width
	switch {
	case ax < bx && az == bz:
		dir = 0
	case ax < bx && az < bz:
		dir = 1
	case ax == bx && az < bz && az%2 == 0:
		dir = 1
	case ax > bx && az < bz:
		dir = 2
	case ax == bx && az < bz && az%2 == 1:
		dir = 2
	case ax > bx && az == bz:
		dir = 3
	case ax > bx && az > bz:
		dir = 4
	case ax == bx && az > bz && az%2 == 1:
		dir = 4
	case ax < bx && az > bz:
		dir = 5
	case ax == bx && az > bz && az%2 == 0:
		dir = 5
	}
	return dir
}

func (m *Map) Distance(a, b *Hex) (dist int) {

	floor2 := func(x int) (f int) {
		if x >= 0 {
			f = x >> 1
		} else {
			f = (x - 1) / 2
		}
		return f
	}

	ceil2 := func(x int) (c int) {
		if x >= 0 {
			c = (x + 1) >> 1
		} else {
			c = x / 2
		}
		return c
	}

	ax, az := a.Index%m.Width, a.Index/m.Width
	bx, bz := b.Index%m.Width, b.Index/m.Width
	dx := float64((bx - floor2(bz)) - (ax - floor2(az)))
	dz := float64((bx + ceil2(bz)) - (ax + ceil2(az)))

	if (dx < 0 && dz < 0) || (dx >= 0 && dz >= 0) {
		dist = int(math.Max(math.Abs(dx), math.Abs(dz)))
	} else {
		dist = int(math.Abs(dx) + math.Abs(dz))
	}
	return dist

}

func (m *Map) HexCenter(hex *Hex) (fx, fy, fz float32) {
	x := hex.Index % m.Width
	z := hex.Index / m.Width
	fx = float32(x)*HEX_WIDTH + float32(z%2)*HEX_WIDTH/2
	fy = float32(hex.Height)
	fz = float32(z) * HEX_HEIGHT
	return fx, fy, fz
}

func (m *Map) Neighbors(hex *Hex) (neighbors []Hex) {
	neighbors = make([]Hex, 0)
	x := hex.Index % m.Width
	z := hex.Index / m.Depth
	neighbors = append(neighbors, m.Grid[(z)*m.Width+(x-1)]) // 3
	neighbors = append(neighbors, m.Grid[(z)*m.Width+(x+1)]) // 0
	neighbors = append(neighbors, m.Grid[(z-1)*m.Width+(x)]) // 4 or 5
	neighbors = append(neighbors, m.Grid[(z+1)*m.Width+(x)]) // 1 or 2
	if z%2 == 0 {
		neighbors = append(neighbors, m.Grid[(z-1)*m.Width+(x-1)]) // 4
		neighbors = append(neighbors, m.Grid[(z+1)*m.Width+(x-1)]) // 2
	} else {
		neighbors = append(neighbors, m.Grid[(z-1)*m.Width+(x+1)]) // 5
		neighbors = append(neighbors, m.Grid[(z+1)*m.Width+(x+1)]) // 1
	}
	return neighbors

}

func (m *Map) Encode() []byte {
	output, _ := json.Marshal(m)
	return output
}

func (m *Map) Decode(data []byte) (*Map, error) {
	err := json.Unmarshal(data, m)

	// reconect object references
	for i := range m.Objects {
		o := m.Objects[i]
		m.Lookup(o.X, o.Y).Unit = o
	}

	return m, err
}
