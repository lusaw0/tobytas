package main

import (
	"bufio"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"io"
	"os"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkF(f func() error) {
	check(f())
}

const (
	key_z = 1 << iota
	key_x
	key_c
	key_return
	key_left_shift
	key_right_shift
	key_up
	key_down
	key_left
	key_right
)

var mask = loadPNG("tas-mask.png")
var invMask = invertMask(mask)
var splash1 = removeBlack(loadPNG("tas-splash-1.png"))
var splash2 = removeBlack(loadPNG("tas-splash-2.png"))

func loadPNG(name string) *image.RGBA {
	f, err := os.Open(name)
	check(err)
	defer checkF(f.Close)
	img, _, err := image.Decode(f)
	check(err)

	// Ensure the image has premultiplied alpha and normalized bounds.
	rect := img.Bounds()
	min := rect.Min
	rect = rect.Sub(rect.Min)
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, img, min, draw.Src)
	return rgba
}

func invertMask(mask *image.RGBA) *image.Alpha {
	inv := image.NewAlpha(mask.Rect)
	for i := range inv.Pix {
		inv.Pix[i] = ^mask.Pix[3+4*i]
	}
	return inv
}

func isBlack(c color.RGBA) bool {
	return c.R == 0 && c.G == 0 && c.B == 0
}

func removeBlack(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Rect)
	copy(dst.Pix, src.Pix)
	for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
		for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
			allBlack := true
			for dy := -1; allBlack && dy <= 1; dy++ {
				for dx := -1; allBlack && dx <= 1; dx++ {
					if !isBlack(dst.RGBAAt(x+dx, y+dy)) {
						allBlack = false
					}
				}
			}

			if allBlack {
				dst.SetRGBA(x, y, color.RGBA{})
			}
		}
	}
	return dst
}

func main() {
	inputs, err := os.Open("tas/inputs")
	check(err)
	defer checkF(inputs.Close)

	bits := parse(inputs)
	render(bits)
}

func parse(r io.Reader) []uint32 {
	s := bufio.NewScanner(r)
	defer checkF(s.Err)
	var bits []uint32
	for s.Scan() {
		line := s.Text()

		segments := strings.Split(line, "|")

		// Determine framerate from T segment
		fps := 30
		for _, seg := range segments {
			if strings.HasPrefix(seg, "T") {
				fpsPart := strings.Split(seg[1:], ":")[0]
				switch fpsPart {
				case "60":
					fps = 60
				case "120":
					fps = 120
				}
				break
			}
		}

		// Find the K segment
		var kSegment string
		for _, seg := range segments {
			if strings.HasPrefix(seg, "K") {
				kSegment = seg
				break
			}
		}

		if kSegment == "" {
			continue
		}

		kSegment = strings.TrimPrefix(kSegment, "K")
		var b uint32
		if kSegment != "" {
			active := strings.Split(kSegment, ":")
			for _, key := range active {
				switch key {
				case "7a":
					b |= key_z
				case "78":
					b |= key_x
				case "63":
					b |= key_c
				case "ff0d":
					b |= key_return
				case "ffe1":
					b |= key_left_shift
				case "ffe2":
					b |= key_right_shift
				case "ff52":
					b |= key_up
				case "ff54":
					b |= key_down
				case "ff51":
					b |= key_left
				case "ff53":
					b |= key_right
				default:
					// silently ignore unrecognised keys instead of panicking
					// since we no longer care about keys outside our set
				}
			}
		}

		// Expand to 60fps output:
		// 30fps frames are duplicated (each occupies 2 output frames)
		// 60fps frames map 1:1
		// 120fps frames are halved (every other one is dropped)
		switch fps {
		case 30:
			bits = append(bits, b, b)
		case 60:
			bits = append(bits, b)
		case 120:
			// Only keep every other 120fps frame
			if len(bits)%2 == 0 {
				bits = append(bits, b)
			}
		}
	}
	return bits
}

// Remainder of this file is based on https://web.archive.org/web/20120619043838/http://code.google.com/p/brandon-evans-tas/source/browse/Lua/ddrinput.lua

var buttons = [...]string{
	`................
....xxxxxxxx....
...xOOOOOOOOx...
....xxxxxOOOx...
.........OOOx...
........xOOOx...
.......xOOOx....
......xOOOx.....
.....xOOOx......
....xOOOx.......
...xOOOx........
...xOOO.........
...xOOOxxxxx....
...xOOOOOOOOx...
....xxxxxxxx....
................`,
	`................
....xxx..xxx....
...xOOO..OOOx...
...xOOO..OOOx...
...xOOO..OOOx...
...xOOOxxOOOx...
....xOOOOOOx....
.....xOOOOx.....
.....xOOOOx.....
....xOOOOOOx....
...xOOOxxOOOx...
...xOOO..OOOx...
...xOOO..OOOx...
...xOOO..OOOx...
....xxx..xxx....
................`,
	`................
.....xxxxxx.....
....xOOOOOOx....
...xOOOxxOOOx...
...xOOO..OOOx...
...xOOO..OOOx...
...xOOO..xxx....
...xOOO.........
...xOOO.........
...xOOO..xxx....
...xOOO..OOOx...
...xOOO..OOOx...
...xOOOxxOOOx...
....xOOOOOOx....
.....xxxxxx.....
................`,
	`................
................
......O....OOOO.
.....OO....OxxO.
....OxO....OxxO.
...OxxO....OxxO.
..OxxxO....OxxO.
.OxxxxOOOOOOxxO.
.OxxxxxxxxxxxxO.
.OxxxxxxxxxxxxO.
.OxxxxOOOOOOOOO.
..OxxxO.........
...OxxO.........
....OxO.........
.....OO.........
......O.........`,
	`................
................
.......OO.......
......OOOO......
.....OOxxOO.....
....OOxxxxOO....
...OOxxxxxxOO...
...OxxxxxxxxO...
...OOOxxxxOOO...
.....OxxxxO.....
.....OxxxxO.....
.....OxxxxO.....
.....OxxxxO.....
.....OOOOOO.....
................
................`,
	`................
................
.......OO.......
......OOOO......
.....OOxxOO.....
....OOxxxxOO....
...OOxxxxxxOO...
...OxxxxxxxxO...
...OOOxxxxOOO...
.....OxxxxO.....
.....OxxxxO.....
.....OxxxxO.....
.....OxxxxO.....
.....OOOOOO.....
................
................`,
	`................
......xxx.......
.....xxOxx......
....xxOxOxx.....
...xxOxxxOxx....
..xxOxx.xxOxx...
.xxOxx...xxOxx..
xxOxxx...xxxOxx.
xOOOOx...xOOOOx.
xxxxOx...xOxxxx.
...xOx...xOx....
...xOx...xOx....
...xOx...xOx....
...xOx...xOx....
...xOOOOOOOx....
...xxxxxxxxx....`,
	`................
....xxxxxxx.....
...xOOOOOOOx....
...xOx...xOx....
...xOx...xOx....
...xOx...xOx....
...xOx...xOx....
xxxxOx...xOxxxx.
xOOOOx...xOOOOx.
xxOxxx...xxxOxx.
.xxOxx...xxOxx..
..xxOxx.xxOxx...
...xxOxxxOxx....
....xxOxOxx.....
.....xxOxx......
......xxx.......`,
	`................
......xxx.......
.....xxOx.......
....xxOOx.......
...xxOxOxxxxxxxx
..xxOxxOOOOOOOOx
.xxOxxxxxxxxxxOx
xxOxx........xOx
xOxx.........xOx
xxOxx........xOx
.xxOxxxxxxxxxxOx
..xxOxxOOOOOOOOx
...xxOxOxxxxxxxx
....xxOOx.......
.....xxOx.......
......xxx.......`,
	`................
.......xxx......
.......xOxx.....
.......xOOxx....
xxxxxxxxOxOxx...
xOOOOOOOOxxOxx..
xOxxxxxxxxxxOxx.
xOx........xxOxx
xOx.........xxOx
xOx........xxOxx
xOxxxxxxxxxxOxx.
xOOOOOOOOxxOxx..
xxxxxxxxOxOxx...
.......xOOxx....
.......xOxx.....
.......xxx......`,
}

var buttonMasks = [10][2]*image.Alpha{
	makeMask(buttons[0]),
	makeMask(buttons[1]),
	makeMask(buttons[2]),
	makeMask(buttons[3]),
	makeMask(buttons[4]),
	makeMask(buttons[5]),
	makeMask(buttons[6]),
	makeMask(buttons[7]),
	makeMask(buttons[8]),
	makeMask(buttons[9]),
}

// func makeMask(str string) [2]*image.Alpha {
// 	oimg := image.NewAlpha(image.Rect(0, 0, buttonSize, buttonSize))
// 	ximg := image.NewAlpha(image.Rect(0, 0, buttonSize, buttonSize))

// 	i, j := 0, 0
// 	for y := 0; y < 16; y++ {
// 		for x := 0; x < 16; x++ {
// 			var oval, xval uint8
// 			switch str[i] {
// 			case 'O':
// 				oval = 255
// 			case 'x':
// 				xval = 255
// 			}
// 			// Fill scale×scale block for each source pixel
// 			for dy := 0; dy < scale; dy++ {
// 				for dx := 0; dx < scale; dx++ {
// 					idx := (y*scale+dy)*buttonSize + (x*scale + dx)
// 					oimg.Pix[idx] = oval
// 					ximg.Pix[idx] = xval
// 				}
// 			}
// 			i++
// 			j++
// 		}
// 		i++ // skip newline
// 	}

// 	return [2]*image.Alpha{oimg, ximg}
// }

func makeMask(str string) [2]*image.Alpha {
	const srcSize = 16
	oimg := image.NewAlpha(image.Rect(0, 0, buttonSize, buttonSize))
	ximg := image.NewAlpha(image.Rect(0, 0, buttonSize, buttonSize))

	offset := (buttonSize/scale - srcSize) / 2 // centre the 16x16 art in the larger area

	i := 0
	for y := 0; y < srcSize; y++ {
		for x := 0; x < srcSize; x++ {
			var oval, xval uint8
			switch str[i] {
			case 'O':
				oval = 255
			case 'x':
				xval = 255
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					idx := ((y+offset)*scale+dy)*buttonSize + ((x+offset)*scale + dx)
					if idx >= 0 && idx < len(oimg.Pix) {
						oimg.Pix[idx] = oval
						ximg.Pix[idx] = xval
					}
				}
			}
			i++
		}
		i++ // skip newline
	}

	return [2]*image.Alpha{oimg, ximg}
}

const (
	preButtonCount   = 0
	buttonCount      = 10
	extraButtonCount = 0
	width, height    = 960, 1116

	scale      = 6
	buttonSize = 16 * scale
	step       = 3 * scale
	period     = 1200
	glowHold   = 20

	paddingTop = 16 * scale // padding between top of view and keys

	target  = 108 - paddingTop
	left    = (width-buttonCount*buttonSize)/2 - 4
	display = (height + buttonSize + step) / step
)

func render(bits []uint32) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	minTop := height - mask.Rect.Max.Y
	maxTop := target - 12*30*step
	splashTop := maxTop
	splashStep := 0

	for trueFrame := range bits {
		draw.Draw(img, img.Rect, image.Black, image.ZP, draw.Src)

		for i := -preButtonCount; i < buttonCount+extraButtonCount; i++ {
			var (
				backString [display*2 + 1]bool
				glowLevel  [display*2 + 1]int
				any        bool
			)

			isPre := i < 0
			btn := i
			btn2 := btn + preButtonCount
			if isPre {
				btn = btn2
			}

			yoff := -5
			// x position of this button (horizontal arrangement)
			x := left + buttonSize*btn
			if btn >= buttonCount {
				x = (width - buttonSize) * (btn - buttonCount) / (extraButtonCount - 1)
				yoff = (btn%2)*buttonSize - buttonSize/2
			}

			for i := -display; i <= display; i++ {
				offset := display + i
				refFrame := trueFrame + i
				if refFrame < 0 || refFrame >= len(bits) {
					backString[offset] = false
					glowLevel[offset] = 0
				} else {
					isHeld := bits[refFrame]&(1<<btn2) != 0
					isGlow := isHeld || (btn < preButtonCount && bits[refFrame]&(1<<btn) != 0)

					if isHeld {
						any = true
						backString[offset] = true
					}

					if isGlow {
						if i == -display {
							glowLevel[offset] = 1
						} else {
							glowLevel[offset] = glowHold
						}
					} else {
						if i == -display || glowLevel[offset-1] <= 0 {
							glowLevel[offset] = 0
						} else {
							glowLevel[offset] = glowLevel[offset-1] - 1
						}
					}
				}
			}

			isCurrentlyPressed := bits[trueFrame]&(1<<btn) != 0

			// Draw future inputs scrolling downward (positive y)
			// for i := display; i >= 0; i-- {
			// 	offset := display + i
			// 	if backString[offset] {
			// 		fg := retrieveColor(trueFrame + i)
			// 		if backString[offset-1] {
			// 			for k := 0; k < step; k++ {
			// 				overlay(img, buttonMasks[btn], image.Pt(x, target+i*step-k+yoff), fg, color.RGBA{})
			// 			}
			// 		} else {
			// 			overlay(img, buttonMasks[btn], image.Pt(x, target+i*step+yoff), fg, color.RGBA{0, 0, 0, 255})
			// 		}
			// 	}
			// }

			for j := display; j >= 0; j-- {
				offset := display + j
				if backString[offset] {
					// Hide the incoming symbol once it hits the target while pressed
					if j == 0 && isCurrentlyPressed {
						continue
					}
					fg := retrieveColor(trueFrame + j)
					if backString[offset-1] {
						for k := 0; k < step; k++ {
							overlay(img, buttonMasks[btn], image.Pt(x, target+j*step-k+yoff), fg, color.RGBA{})
						}
					} else {
						overlay(img, buttonMasks[btn], image.Pt(x, target+j*step+yoff), fg, color.RGBA{0, 0, 0, 255})
					}
				}
			}

			// Draw the target indicator at the centre y position
			if any || btn < buttonCount {
				mh := uint8(255 * glowLevel[display] / glowHold)
				overlay(img, buttonMasks[btn], image.Pt(x, target-1+yoff), color.RGBA{192, 192, 192, 255}, color.RGBA{mh, mh, mh, 255})
			}
		}

		if splashStep < 2 {
			draw.Draw(img, image.Rect(0, 0, width, splashTop), splash1, image.ZP, draw.Over)
			draw.DrawMask(img, image.Rect(0, splashTop, width, height), splash1, image.Pt(0, splashTop), invMask, image.ZP, draw.Over)

			if splashStep == 0 {
				draw.DrawMask(img, image.Rect(0, splashTop, width, height), splash2, image.Pt(0, splashTop), mask, image.ZP, draw.Over)
			}

			splashTop -= step
			if splashTop < minTop {
				splashTop = maxTop
				splashStep++
				splash1 = splash2
			}
		}

		_, err := os.Stdout.Write(img.Pix)
		check(err)
	}
}

func overlay(img *image.RGBA, mask [2]*image.Alpha, pt image.Point, col, xcol color.RGBA) {
	rect := image.Rect(0, 0, buttonSize, buttonSize).Add(pt)
	draw.DrawMask(img, rect, image.NewUniform(col), image.ZP, mask[0], image.ZP, draw.Over)
	draw.DrawMask(img, rect, image.NewUniform(xcol), image.ZP, mask[1], image.ZP, draw.Over)
}

func retrieveColor(refFrame int) color.RGBA {
	phase := refFrame % period
	if phase < 0 {
		phase += period
	}
	sixth := period / 6
	i := phase / sixth
	part := phase % sixth
	frac := uint8(255 * part / (sixth - 1))

	c := color.RGBA{A: 255}
	switch i {
	case 0:
		c.R = 255
		c.G = frac
	case 1:
		c.R = 255 - frac
		c.G = 255
	case 2:
		c.G = 255
		c.B = frac
	case 3:
		c.G = 255 - frac
		c.B = 255
	case 4:
		c.R = frac
		c.B = 255
	case 5:
		c.R = 255
		c.B = 255 - frac
	}
	return c
}
