package main

import "github.com/mjard/gl"
import "github.com/go-gl/glu"

type Camera struct {
	x, y, z    float64
	rx, ry, rz float64
}

func (c *Camera) Init(width, height, far, fov float64) {
	gl.Viewport(0, 0, int(width), int(height))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	glu.Perspective(fov, width/height, 1.0, far)
	gl.MatrixMode(gl.MODELVIEW)
}

// func (c *Camera) OrthoIn(width, height float64) {
// 	// 	gl.Viewport(0, 0, w, h)
// 	gl.PushMatrix()
// 	gl.PushAttrib(gl.CURRENT_BIT | gl.COLOR_BUFFER_BIT | gl.ENABLE_BIT | gl.LIGHTING_BIT | gl.POLYGON_BIT | gl.LINE_BIT)
// 	gl.MatrixMode(gl.PROJECTION)
// 	gl.LoadIdentity()
// 	gl.Ortho(0, float64(width), float64(height), 0, 0, 1)
// 	gl.MatrixMode(gl.MODELVIEW)
// }

// func (c *Camera) OrthoOut() {
// 	// gl.Viewport(0, 0, w, h)
// 	gl.PopAttrib()
// 	gl.PopMatrix()

// }

func (c *Camera) Enter() {
	gl.PushMatrix()
	gl.PushAttrib(gl.CURRENT_BIT | gl.ENABLE_BIT | gl.LIGHTING_BIT | gl.POLYGON_BIT | gl.LINE_BIT)

	// gl.Translatef(0, 0, -1800)
	gl.Rotated(c.rx, 1, 0, 0)
	gl.Rotated(c.ry, 0, 1, 0)
	gl.Rotated(c.rz, 0, 0, 1)
	gl.Translated(-c.x, -c.y, -c.z)
}

func (c *Camera) Exit() {
	gl.PopAttrib()
	gl.PopMatrix()
}

func (c *Camera) WorldCoords(mx, my int) (x, y, z float64) {

	mz := float32(0.0)
	model := [16]float64{}
	proj := [16]float64{}
	view := [4]int32{}

	c.Enter()
	defer c.Exit()

	gl.GetDoublev(gl.MODELVIEW_MATRIX, model[:])
	gl.GetDoublev(gl.PROJECTION_MATRIX, proj[:])
	gl.GetIntegerv(gl.VIEWPORT, view[:])

	// modify my so it reports correctly
	my = int(view[3]) - my

	gl.ReadPixels(mx, my, 1, 1, gl.DEPTH_COMPONENT, gl.FLOAT, &mz)
	// fmt.Println(mz)
	// s := 0.28
	// ex0, ey0, ez0 := glu.UnProject(float64(mx), float64(my), 0, &model, &proj, &view)
	// ex1, ey1, ez1 := glu.UnProject(float64(mx), float64(my), 1, &model, &proj, &view)

	// now intesect line with 
	//ex0 + (ex1-ex0)*s, ey0 + (ey1-ey0)*s, ez0 + (ez1-ez0)*s	
	return glu.UnProject(float64(mx), float64(my), float64(mz), &model, &proj, &view)

}
