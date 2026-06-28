//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func main() {
	size := 256
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Background: transparent
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Draw rounded square with blue gradient
	radius := 56.0
	accent := color.RGBA{R: 0, G: 113, B: 227, A: 255}
	accentLight := color.RGBA{R: 50, G: 153, B: 255, A: 255}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !inRoundedRect(float64(x), float64(y), 8, 8, float64(size-8), float64(size-8), radius) {
				continue
			}
			t := float64(y) / float64(size)
			c := lerpColor(accent, accentLight, t)
			img.Set(x, y, c)
		}
	}

	// Draw a white window/layered rectangle symbol in the center
	drawWindowSymbol(img, size)

	outPath := filepath.Join("winres", "icon.png")
	if err := os.MkdirAll("winres", 0755); err != nil {
		panic(err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
	println("generated:", outPath)
}

func inRoundedRect(x, y, x0, y0, x1, y1, r float64) bool {
	if x < x0 || x > x1 || y < y0 || y > y1 {
		return false
	}
	if x < (x0+r) && y < (y0+r) {
		return dist(x, y, x0+r, y0+r) <= r
	}
	if x > (x1-r) && y < (y0+r) {
		return dist(x, y, x1-r, y0+r) <= r
	}
	if x < (x0+r) && y > (y1-r) {
		return dist(x, y, x0+r, y1-r) <= r
	}
	if x > (x1-r) && y > (y1-r) {
		return dist(x, y, x1-r, y1-r) <= r
	}
	return true
}

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

func drawWindowSymbol(img *image.RGBA, size int) {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	whiteSemi := color.RGBA{R: 255, G: 255, B: 255, A: 180}

	// Back window (semi-transparent, offset)
	drawRect(img, 76, 96, 196, 176, whiteSemi)

	// Front window (opaque)
	drawRect(img, 60, 80, 180, 160, white)

	// Title bar line on front window
	drawRect(img, 60, 80, 180, 100, color.RGBA{R: 0, G: 113, B: 227, A: 255})
}

func drawRect(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}
