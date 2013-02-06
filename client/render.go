package main

import (
	// "fmt"
	"github.com/go-gl/glfw"
	"github.com/mjard/gl"
	. "github.com/sixthgear/thewar/gamelib"
	"log"
	"math"
)

type Renderer interface {
	Init()
	Render()
}

const (
	W_FAR             = 8192
	W_FOV             = 60
	PRIMITIVE_RESTART = math.MaxUint32
)

type MapRenderer struct {
	hexes    RenderList
	objects  RenderList
	paths    RenderList
	camera   *Camera
	texAtlas gl.Texture
}

type RenderList struct {
	GLtype    gl.GLenum
	indices   []uint32
	outlines  []uint32
	vertices  []float32
	colors    []float32
	texcoords []float32
}

func (r *MapRenderer) Init() {

	// create camera
	r.camera = &Camera{x: M_WIDTH * HEX_WIDTH * 0.5, y: 1200, z: M_DEPTH*HEX_HEIGHT*0.5 + 900, rx: 80, ry: 0, rz: 0}
	r.camera.Init(W_WIDTH, W_HEIGHT, W_FAR, W_FOV)

	// load resources
	gl.Enable(gl.TEXTURE_2D)
	r.texAtlas = gl.GenTexture()
	r.texAtlas.Bind(gl.TEXTURE_2D)
	if !glfw.LoadTexture2D("data/sprites.tga", 0) {
		log.Fatal("Failed to load texture atlas!")
	}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	r.texAtlas.Unbind(gl.TEXTURE_2D)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Color4f(1, 1, 1, 1)
	gl.ClearColor(0.1, 0.05, 0.0, 1.0)

}

func (r *MapRenderer) buildVertices(m *Map) {

	r.hexes.GLtype = gl.TRIANGLES
	r.hexes.indices = make([]uint32, 0)
	r.hexes.outlines = make([]uint32, 0)
	r.hexes.vertices = make([]float32, 0)
	r.hexes.colors = make([]float32, 0)

	count := uint32(0)

	for i := 0; i < m.Width*m.Depth; i++ {

		hex := m.Index(i)
		color := hex.Color
		fx, fy, fz := m.HexCenter(hex)

		// build hex points around center
		r.hexes.vertices = append(r.hexes.vertices,
			fx, fy, fz,
			fx-HEX_WIDTH/2+HEX_GAP, fy, fz-HEX_HEIGHT/3+HEX_GAP,
			fx, fy, fz-HEX_SIZE/2+HEX_GAP,
			fx+HEX_WIDTH/2-HEX_GAP, fy, fz-HEX_HEIGHT/3+HEX_GAP,
			fx+HEX_WIDTH/2-HEX_GAP, fy, fz+HEX_HEIGHT/3-HEX_GAP,
			fx, fy, fz+HEX_SIZE/2-HEX_GAP,
			fx-HEX_WIDTH/2+HEX_GAP, fy, fz+HEX_HEIGHT/3-HEX_GAP,
		)

		// add colors
		for j := 0; j < 7; j++ {
			r.hexes.colors = append(r.hexes.colors, color[0], color[1], color[2])
		}

		// create hex from 6 triangles		
		for j := uint32(0); j < 6; j++ {
			k := count + j + 1
			kk := count + (j+1)%6 + 1
			r.hexes.indices = append(r.hexes.indices, count, k, kk)
		}

		// add 3 line segments for grid display
		r.hexes.outlines = append(r.hexes.outlines, count+1, count+2, count+2, count+3, count+3, count+4)

		count += 7
	}
}

func (r *MapRenderer) buildObjects(m *Map) {

	r.objects.GLtype = gl.QUADS
	r.objects.indices = make([]uint32, 0)
	r.objects.vertices = make([]float32, 0)
	r.objects.texcoords = make([]float32, 0)
	count := uint32(0)

	for _, o := range m.Objects {
		// hex := m.Lookup(o.x, o.y)
		x, y, z := o.Fx, o.Fy, o.Fz //hexCenter(m, hex)

		r.objects.indices = append(r.objects.indices,
			count+0, count+1, count+2, count+3, // top
			count+1, count+2, count+6, count+5, // front
			count+0, count+3, count+7, count+4, // back
			count+0, count+1, count+5, count+4, // left
			count+2, count+3, count+7, count+6, // right
		)

		r.objects.outlines = append(r.objects.outlines,
			count+0, count+1, count+1, count+2,
			count+2, count+3, count+3, count+0,
		)

		angle := float64(o.Facing)/6*(math.Pi*2) + 0.785398163

		c := float32(math.Cos(angle)) * 52
		s := float32(math.Sin(angle)) * 52

		r.objects.vertices = append(r.objects.vertices,
			x+c, y+16, z+s, // 0
			x-s, y+16, z+c, // 1
			x-c, y+16, z-s, // 2
			x+s, y+16, z-c, // 3
			x+c, y+0, z+s, // 4
			x-s, y+0, z+c, // 5
			x-c, y+0, z-s, // 6
			x+s, y+0, z-c, // 7
		)

		tx0 := float32(o.Team%4) / 4.0
		ty0 := float32(o.Type%4) / 4.0
		tx1 := tx0 + 1.0/4.0
		ty1 := ty0 + 1.0/4.0

		r.objects.texcoords = append(r.objects.texcoords, tx0, ty1, tx0, ty0, tx1, ty0, tx1, ty1)
		r.objects.texcoords = append(r.objects.texcoords, tx0, ty1, tx0, ty0, tx1, ty0, tx1, ty1)

		count += 8
	}
}

func (r *MapRenderer) buildPath(m *Map, path []int) {
	r.paths.GLtype = gl.LINES
	r.paths.indices = make([]uint32, 0)
	r.paths.vertices = make([]float32, 0)

	for i := 0; i < len(path); i++ {
		ax, ay, az := m.HexCenter(m.Index(path[i]))
		if i == 0 {
			bx, _, bz := m.HexCenter(m.Index(path[i+1]))
			a := math.Atan2(float64(bz-az), float64(bx-ax))
			x0 := ax + float32(math.Cos(a+0.523598776)*30)
			z0 := az + float32(math.Sin(a+0.523598776)*30)
			x1 := ax + float32(math.Cos(a-0.523598776)*30)
			z1 := az + float32(math.Sin(a-0.523598776)*30)
			r.paths.vertices = append(r.paths.vertices, x0, ay, z0)
			r.paths.vertices = append(r.paths.vertices, ax, ay, az)
			r.paths.vertices = append(r.paths.vertices, x1, ay, z1)
		}

		r.paths.vertices = append(r.paths.vertices, ax, ay, az)
	}

	for i := 0; i < len(path)-1+3; i++ {
		r.paths.indices = append(r.paths.indices, uint32(i), uint32(i+1))
	}

}

func (r *MapRenderer) clearPath() {
	r.paths.indices = make([]uint32, 0)
	r.paths.vertices = make([]float32, 0)
}

func (r *MapRenderer) Render(m *Map) {

	// gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	// gl.LoadIdentity()

	r.camera.Enter()

	// gl.Enable(gl.PRIMITIVE_RESTART)
	// gl.PrimitiveRestartIndex(PRIMITIVE_RESTART)
	gl.PushClientAttrib(gl.CLIENT_VERTEX_ARRAY_BIT)
	gl.PushAttrib(gl.CURRENT_BIT | gl.ENABLE_BIT | gl.LINE_BIT | gl.DEPTH_BUFFER_BIT)

	// draw r.hexes and terrain
	gl.EnableClientState(gl.COLOR_ARRAY)
	gl.ColorPointer(3, gl.FLOAT, 0, r.hexes.colors)
	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.VertexPointer(3, gl.FLOAT, 0, r.hexes.vertices)
	gl.DrawElements(r.hexes.GLtype, len(r.hexes.indices), gl.UNSIGNED_INT, r.hexes.indices)
	gl.DisableClientState(gl.COLOR_ARRAY)

	// draw selected hex
	if m.Selected != nil {
		start := m.Selected.Index * 18
		gl.Color4f(1, 0, 0, 0.75)
		gl.DrawElements(r.hexes.GLtype, 18, gl.UNSIGNED_INT, &r.hexes.indices[start])
	}

	// draw grid
	gl.Color4f(0.15, 0.15, 0.15, 1)
	gl.Enable(gl.LINE_SMOOTH)
	gl.Enable(gl.POLYGON_SMOOTH)
	gl.Hint(gl.LINE_SMOOTH_HINT, gl.NICEST)
	gl.Hint(gl.POLYGON_SMOOTH_HINT, gl.NICEST)
	// gl.Translatef(0, 10, 0)
	gl.LineWidth(1)
	gl.DrawElements(gl.LINES, len(r.hexes.outlines), gl.UNSIGNED_INT, r.hexes.outlines)

	// draw path
	if len(r.paths.indices) > 0 {

		gl.Disable(gl.DEPTH_TEST)
		gl.DisableClientState(gl.COLOR_ARRAY)
		gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
		// gl.Enable(gl.BLEND)
		// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.VertexPointer(3, gl.FLOAT, 0, r.paths.vertices)
		gl.LineWidth(5.0)
		gl.Color4f(1, 1, 0, 1)
		gl.DrawElements(r.paths.GLtype, len(r.paths.indices), gl.UNSIGNED_INT, r.paths.indices)
		gl.Enable(gl.DEPTH_TEST)
	}

	// draw r.objects
	gl.Color4f(1, 1, 1, 1)
	gl.Enable(gl.TEXTURE_2D)
	gl.VertexPointer(3, gl.FLOAT, 0, r.objects.vertices)
	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	gl.TexCoordPointer(2, gl.FLOAT, 0, r.objects.texcoords)
	r.texAtlas.Bind(gl.TEXTURE_2D)
	gl.DrawElements(r.objects.GLtype, len(r.objects.indices), gl.UNSIGNED_INT, r.objects.indices)
	r.texAtlas.Unbind(gl.TEXTURE_2D)

	gl.PopAttrib()
	gl.PopClientAttrib()
	r.camera.Exit()

}
