package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"github.com/mjard/gl"
	"image"
	"image/draw"
	"image/png"
	// "log"
	"os"
)

type TextLabel struct {
	X       float32
	Y       float32
	Size    int
	Text    string
	Font    *Font
	letters RenderList
}

type Char struct {
	XMLName  xml.Name `xml:"char"`
	Id       int      `xml:"id,attr"`
	Letter   string   `xml:"letter,attr"`
	X        float32  `xml:"x,attr"`
	Y        float32  `xml:"y,attr"`
	Width    float32  `xml:"width,attr"`
	Height   float32  `xml:"height,attr"`
	XOffset  float32  `xml:"xoffset,attr"`
	YOffset  float32  `xml:"yoffset,attr"`
	XAdvance float32  `xml:"xadvance,attr"`
}

type Metric struct {
	XMLName xml.Name `xml:"font"`
	Chars   []Char   `xml:"chars>char"`
	// FileName string   `xml:"pages>page>file,attr"`
	// Kerns    []Kern   `xml:"kernings>kerning"`
}

// type Kern struct {
// 	XMLName xml.Name `xml:"kerning"`
// 	First   int      `xml:"first,attr"`
// 	Second  int      `xml:"second,attr"`
// 	Smount  int      `xml:"amount,attr"`
// }

type Font struct {
	name    string
	glyphs  gl.Texture
	mapping map[rune]*Char
}

func (f *Font) Load(name string) (lf *Font, err error) {

	var imgF, datF *os.File
	var p image.Image

	f.mapping = make(map[rune]*Char, 52)

	if imgF, err = os.Open(fmt.Sprintf("data/%s.png", name)); err != nil {
		return nil, err
	} else {
		defer imgF.Close()
	}
	if p, err = png.Decode(bufio.NewReader(imgF)); err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(image.Rect(0, 0, p.Bounds().Dx(), p.Bounds().Dy()))
	draw.Draw(rgba, rgba.Bounds(), p, p.Bounds().Min, draw.Src)

	f.glyphs = gl.GenTexture()
	f.glyphs.Bind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, rgba.Bounds().Dx(), rgba.Bounds().Dy(), 0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	f.glyphs.Unbind(gl.TEXTURE_2D)

	metric := new(Metric)
	data := make([]byte, 1000000)

	if datF, err = os.Open(fmt.Sprintf("data/%s.fnt", name)); err != nil {
		return nil, err
	} else {
		defer datF.Close()
	}
	if _, err = bufio.NewReader(datF).Read(data); err != nil {
		return nil, err
	}
	if err = xml.Unmarshal(data, metric); err != nil {
		return nil, err
	}

	for i, c := range metric.Chars {
		switch c.Letter {
		case "space":
			f.mapping[' '] = &metric.Chars[i]
		case "&amp;":
			f.mapping['&'] = &metric.Chars[i]
		case "&lt;":
			f.mapping['<'] = &metric.Chars[i]
		case "&gt;":
			f.mapping['>'] = &metric.Chars[i]
		case "&quot;":
			f.mapping['"'] = &metric.Chars[i]
		default:
			b := rune(c.Letter[0])
			f.mapping[b] = &metric.Chars[i]
		}
	}

	return f, err
}

func (t *TextLabel) Init(text string, f *Font, x, y float32) {

	t.X = x
	t.Y = y
	t.Font = f
	t.SetText(text)
}

func (t *TextLabel) SetText(text string) {

	if text == t.Text {
		return
	}

	t.Text = text
	t.letters.GLtype = gl.QUADS
	t.letters.vertices = make([]float32, 0)
	t.letters.texcoords = make([]float32, 0)

	kern := float32(0.0)
	for _, l := range t.Text {
		c, ok := t.Font.mapping[l]
		if ok {
			t.letters.vertices = append(t.letters.vertices,
				kern+c.XOffset, 0+c.YOffset,
				kern+c.XOffset+c.Width, 0+c.YOffset,
				kern+c.XOffset+c.Width, c.YOffset+c.Height,
				kern+c.XOffset, c.YOffset+c.Height,
			)
			kern += c.XAdvance
			texL := c.X / 256
			texT := c.Y / 128
			texR := texL + c.Width/256
			texB := texT + c.Height/128
			t.letters.texcoords = append(t.letters.texcoords,
				texL, texT,
				texR, texT,
				texR, texB,
				texL, texB,
			)
		}
	}

}

func (t *TextLabel) Render(c *Camera) {

	c.OrthoIn(W_WIDTH, W_HEIGHT)

	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.VertexPointer(2, gl.FLOAT, 0, t.letters.vertices)
	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	gl.TexCoordPointer(2, gl.FLOAT, 0, t.letters.texcoords)

	t.Font.glyphs.Bind(gl.TEXTURE_2D)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	// gl.BlendFunc(gl.SRC_ALPHA, gl.DST_ALPHA)
	gl.TexEnvi(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
	gl.Translatef(t.X, t.Y, 0)
	gl.Color4f(1, 1, 1, 1)
	gl.DrawArrays(t.letters.GLtype, 0, len(t.letters.vertices)/2)
	t.Font.glyphs.Unbind(gl.TEXTURE_2D)

	gl.Disable(gl.BLEND)
	c.OrthoOut()
}
