package gif

// WARNING: THIS ENTIRE FILE WAS VIBE CODED

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"math/rand"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Wheel colors for 4 spokes (solid, user-specified)
var wheelColors = []color.Color{
	color.RGBA{213, 15, 37, 255},  // Red
	color.RGBA{51, 105, 232, 255}, // Blue
	color.RGBA{0, 153, 37, 255},   // Green
	color.RGBA{238, 178, 17, 255}, // Yellow
}

// Draw a filled arc (pie slice)
func drawArc(img *image.Paletted, cx, cy, r int, start, end float64, col color.Color) {
	for angle := start; angle < end; angle += 0.001 {
		for rad := 0; rad < r; rad++ {
			x := cx + int(float64(rad)*math.Cos(angle))
			y := cy + int(float64(rad)*math.Sin(angle))
			img.Set(x, y, col)
		}
	}
}

// CreateSpinningWheelGIF creates a 200x200 spinning wheel GIF and saves it to filename
// nSpokes: number of wheel segments
// endText: text to display in a box after the wheel stops
func CreateSpinningWheelGIF(filename string, nSpokes int, endText string) (*os.File, error) {
	const size = 200
	const center = size / 2
	const radius = size / 2
	const pauseFrames = 20  // 2 seconds at 10 fps
	const textBoxFrames = 5 // frames for text box animation

	// Palette: white background + 6 colors + black outline
	palette := []color.Color{color.White, color.Black}
	palette = append(palette, wheelColors...)

	var images []*image.Paletted
	var delays []int

	// Spin: fast at first, then slow down exponentially
	angle := 0.0
	angleStep := 0.8 // radians per frame, will decrease
	expBase := 0.95  // exponential base for slowing down
	minStep := 0.01  // when to consider the wheel stopped

	// Calculate number of frames until angleStep < minStep
	nFrames := 0
	tempStep := angleStep
	for tempStep > minStep {
		tempStep *= expBase
		nFrames++
	}
	nFrames = int(float64(nFrames) * 1.2) // add 20% more frames for realism

	for frame := 0; frame < nFrames; frame++ {
		img := image.NewPaletted(image.Rect(0, 0, size, size), palette)
		draw.Draw(img, img.Rect, &image.Uniform{color.White}, image.Point{}, draw.Src)

		// Draw wheel
		spokeAngle := 2 * math.Pi / float64(nSpokes)
		for i := 0; i < nSpokes; i++ {
			start := angle + float64(i)*spokeAngle
			end := start + spokeAngle
			// Alternate equally between the 4 colors
			col := wheelColors[i%len(wheelColors)]
			drawArc(img, center, center, radius, start, end, col)
		}

		// Draw outline
		for t := 0.0; t < 2*math.Pi; t += 0.001 {
			x := center + int(float64(radius)*math.Cos(t))
			y := center + int(float64(radius)*math.Sin(t))
			img.Set(x, y, color.Black)
		}

		images = append(images, img)
		delays = append(delays, 3) // ~30 fps

		// Exponential slow down
		angle += angleStep
		angleStep *= expBase // exponential decrease
	}

	// --- Confetti burst BEFORE text bubble ---
	type confetto struct {
		x, y   float64
		vx, vy float64
		color  color.Color
		size   int
	}
	confettiColors := []color.Color{
		color.RGBA{213, 15, 37, 255},
		color.RGBA{51, 105, 232, 255},
		color.RGBA{0, 153, 37, 255},
		color.RGBA{238, 178, 17, 255},
		color.RGBA{255, 255, 255, 255},
	}
	confettiBundles := 40
	bundleSize := 12
	nConfetti := confettiBundles * bundleSize
	confetti := make([]confetto, 0, nConfetti)
	for b := 0; b < confettiBundles; b++ {
		// Randomize confetti start position anywhere in the image
		bx := rand.Float64() * float64(size)
		by := rand.Float64() * float64(size)
		baseAngle := rand.Float64() * 2 * math.Pi
		for j := 0; j < bundleSize; j++ {
			angle := baseAngle + (float64(j)/float64(bundleSize))*2*math.Pi + rand.Float64()*0.2
			speed := 5 + rand.Float64()*3
			confetti = append(confetti, confetto{
				x:     bx,
				y:     by,
				vx:    speed * math.Cos(angle),
				vy:    speed * math.Sin(angle),
				color: confettiColors[(b*bundleSize+j)%len(confettiColors)],
				size:  4 + rand.Intn(2),
			})
		}
	}
	confettiFrames := 100 // Double the duration of confetti and text box animation
	// Font scale based on size (base: 200px -> scale 5)
	scale := size / 80
	if scale < 1 {
		scale = 1
	}
	// Cache the last wheel frame to use as a clean background for each confetti frame
	lastWheelFrame := images[len(images)-1]
	bubbleAnimFrames := 20 // Number of frames for the bubble to animate in
	for f := 0; f < confettiFrames; f++ {
		// Instead of drawing on top of the previous confetti frame, always start from the last wheel frame
		img := image.NewPaletted(image.Rect(0, 0, size, size), palette)
		draw.Draw(img, img.Rect, lastWheelFrame, image.Point{}, draw.Over)
		for i := range confetti {
			confetti[i].vy += 0.5 + 0.1*math.Sin(float64(i)+float64(f))
			confetti[i].vx *= 0.98
			confetti[i].vy *= 0.98
			confetti[i].x += confetti[i].vx
			confetti[i].y += confetti[i].vy
			for dx := 0; dx < confetti[i].size; dx++ {
				for dy := 0; dy < confetti[i].size; dy++ {
					x := int(confetti[i].x) + dx
					y := int(confetti[i].y) + dy
					if x >= 0 && x < size && y >= 0 && y < size {
						img.Set(x, y, confetti[i].color)
					}
				}
			}
		}
		// Animate the text bubble flying in from the top, then keep it centered
		boxW := int(float64(size) * 0.8)
		boxH := int(float64(size) * 0.25)
		boxX := center - boxW/2
		startY := -boxH
		endY := center - boxH/2
		var bubbleY int
		if f < bubbleAnimFrames {
			progress := float64(f) / float64(bubbleAnimFrames-1)
			// Ease out cubic for smoothness
			bubbleY = int(float64(startY)*(1-progress)*(1-progress)*(1-progress) + float64(endY)*(1-(1-progress)*(1-progress)*(1-progress)))
		} else {
			bubbleY = endY
		}
		// Only clear and draw the text box area, not the whole column
		for x := boxX; x < boxX+boxW; x++ {
			for y := bubbleY; y < bubbleY+boxH; y++ {
				img.Set(x, y, color.White)
			}
		}
		// Now draw the text box border
		for x := boxX; x < boxX+boxW; x++ {
			img.Set(x, bubbleY, color.Black)
			img.Set(x, bubbleY+boxH-1, color.Black)
		}
		for y := bubbleY; y < bubbleY+boxH; y++ {
			img.Set(boxX, y, color.Black)
			img.Set(boxX+boxW-1, y, color.Black)
		}
		// Draw text (scaled up, as before)
		text := endText
		face := basicfont.Face7x13
		textImg := image.NewRGBA(image.Rect(0, 0, font.MeasureString(face, text).Round(), face.Metrics().Height.Ceil()))
		textDrawer := &font.Drawer{
			Dst:  textImg,
			Src:  image.NewUniform(color.Black),
			Face: face,
			Dot:  fixed.P(0, face.Metrics().Ascent.Ceil()),
		}
		textDrawer.DrawString(text)
		textWidth := textImg.Bounds().Dx() * scale
		textHeight := textImg.Bounds().Dy() * scale
		textX := center - textWidth/2
		textY := bubbleY + boxH/2 - textHeight/2
		for y := 0; y < textImg.Bounds().Dy(); y++ {
			for x := 0; x < textImg.Bounds().Dx(); x++ {
				r, g, b, a := textImg.At(x, y).RGBA()
				if a > 0x8000 && r < 0x4000 && g < 0x4000 && b < 0x4000 {
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							img.Set(textX+x*scale+dx, textY+y*scale+dy, color.Black)
						}
					}
				}
			}
		}
		images = append(images, img)
		delays = append(delays, 5)
	}

	// Save GIF
	var f *os.File
	var err error
	if filename != "" {
		f, err = os.Create(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
	} else {
		f, err = os.CreateTemp("", "spinning_wheel_*.gif")
		if err != nil {
			return nil, err
		}
		// Do not close here, caller will handle it
	}

	err = gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	if err != nil {
		if filename == "" && f != nil {
			f.Close()
			os.Remove(f.Name())
		}
		return nil, err
	}

	if filename != "" {
		return nil, nil
	}
	// Seek to beginning for reading
	_, err = f.Seek(0, 0)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	return f, nil
}

// func main() {
// 	err := CreateSpinningWheelGIF("spinning_wheel.gif", 12, "456")
// 	if err != nil {
// 		log.Fatalf("Failed to create GIF: %v", err)
// 	}
// 	log.Println("GIF created: spinning_wheel.gif")
// }
