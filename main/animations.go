package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type animatedButton struct {
	key       string
	x         float64
	y         float64
	baseScale float64
	scale     float64
	pressed   bool
}

func newAnimatedButton(key string, x float64, y float64, baseScale float64, scale float64) *animatedButton {
	return &animatedButton{
		key:       key,
		x:         x,
		y:         y,
		baseScale: baseScale,
		scale:     scale,
	}
}

func (b *animatedButton) baseSize() (float64, float64) {
	imgBounds := images[b.key].Bounds()
	return float64(imgBounds.Dx()) * b.baseScale, float64(imgBounds.Dy()) * b.baseScale
}

func (b *animatedButton) contains(px, py int) bool {
	width, height := b.baseSize()
	mx := float64(px)
	my := float64(py)
	return mx >= b.x && mx <= b.x+width && my >= b.y && my <= b.y+height
}

func (b *animatedButton) update(hovered, mousePressed bool) {
	targetScale := b.baseScale
	if hovered {
		targetScale = b.baseScale * 1.08
	}
	if hovered && mousePressed {
		targetScale = b.baseScale * 0.94
		b.pressed = true
	} else {
		b.pressed = false
	}

	b.scale += (targetScale - b.scale) * 0.25
}

func (b *animatedButton) draw(screen *ebiten.Image) {
	scale := b.scale
	imgBounds := images[b.key].Bounds()
	imgW := float64(imgBounds.Dx())
	imgH := float64(imgBounds.Dy())
	baseW, baseH := b.baseSize()
	cx := b.x + baseW/2
	cy := b.y + baseH/2

	yOffset := 0.0
	if b.pressed {
		yOffset = 2.0
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-imgW/2, -imgH/2)
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(cx, cy+yOffset)
	screen.DrawImage(images[b.key], opts)
}

// NOTE : controlHotspot is a click-region animator for +/- buttons that are part of
// a single source image. It does not draw the button art itself; it overlays
// a subtle tint/glow that scales in ("press in") or out ("pop out").
type controlHotspot struct {
	x, y, w, h float64
	scale      float64
	pressed    bool
}

func newControlHotspot(x, y, w, h float64) *controlHotspot {
	return &controlHotspot{x: x, y: y, w: w, h: h, scale: 1.0}
}

func (c *controlHotspot) contains(px, py int) bool {
	return pointInRect(float64(px), float64(py), c.x, c.y, c.w, c.h)
}

func (c *controlHotspot) update(hovered, mousePressed bool) {
	target := 1.0
	if hovered {
		target = 1.06
	}
	if hovered && mousePressed {
		target = 0.9
		c.pressed = true
	} else {
		c.pressed = false
	}
	c.scale += (target - c.scale) * 0.28
}

func (c *controlHotspot) draw(screen *ebiten.Image) {
	if math.Abs(c.scale-1.0) < 0.005 {
		return
	}

	cx := c.x + c.w/2
	cy := c.y + c.h/2
	sw := c.w * c.scale
	sh := c.h * c.scale

	var clr color.RGBA
	if c.scale < 1.0 {
		intensity := uint8(math.Min(180, (1.0-c.scale)*1400))
		clr = color.RGBA{R: 25, G: 12, B: 4, A: intensity} // Deep Espresso
	} else {
		intensity := uint8(math.Min(110, (c.scale-1.0)*1100))
		clr = color.RGBA{R: 90, G: 45, B: 20, A: intensity} // Dark Russet / Mahogany
	}

	vector.FillRect(
		screen,
		float32(cx-sw/2),
		float32(cy-sh/2),
		float32(sw),
		float32(sh),
		clr,
		true,
	)
}
