// go:build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// ── Config ────────────────────────────────────────────────────────────────────

const (
	width            = 960
	height           = 960
	fps              = 60
	fontPath         = "DTM-Mono.otf"
	fontSize         = 40
	smallSize        = 48
	segmentTimerSize = 64
	segmentCsSize    = 48
	timerSize        = 82
	timerCsSize      = 58
	timerPad         = 1 // tight vertical padding around timer glyphs
	rowHeight        = 46
	statsGap         = 2
	iconSize         = 32
	headerPad        = -5
)

var (
	colorBg         = color.RGBA{0, 0, 0, 255}
	colorText       = color.RGBA{255, 255, 255, 255}
	colorHighlight  = color.RGBA{50, 130, 255, 255}
	colorGreen      = color.RGBA{0, 255, 0, 255}
	colorGold       = color.RGBA{255, 215, 0, 255}
	colorRed        = color.RGBA{255, 50, 50, 255}
	colorTimerGreen = color.RGBA{0, 220, 0, 255}
	colorTimerRed   = color.RGBA{220, 0, 0, 255}
	colorSeparator  = color.RGBA{0, 0, 0, 255}
)

// ── Data types ────────────────────────────────────────────────────────────────

type Split struct {
	Name      string        `json:"name"`
	PBTime    time.Duration `json:"pb_time"`   // personal best time for this split
	BestTime  time.Duration `json:"best_time"` // best ever segment time
	SplitTime time.Duration // actual time when split was triggered
	Completed bool
}

type State struct {
	GameName      string  `json:"game_name"`
	CategoryName  string  `json:"category_name"`
	Splits        []Split `json:"splits"`
	SplitFrames   []int   `json:"split_frames"` // inputs list index (0-based) for each split
	PBAttempts    int     `json:"pb_attempts"`
	TotalAttempts int     `json:"total_attempts"`

	CurrentSplit    int
	StartTime       time.Time
	Running         bool
	Finished        bool
	CurrentTime     time.Duration
	PreviousSegment time.Duration
}

func loadState(path string) *State {
	f, err := os.Open(path)
	if err != nil {
		// Return a default state if no file found
		return &State{
			GameName:      "Undertale",
			CategoryName:  "Neutral Any%",
			PBAttempts:    38,
			TotalAttempts: 1419,
			Splits: []Split{
				{Name: "Ruins", PBTime: 6*time.Minute + 50*time.Second, BestTime: 6*time.Minute + 30*time.Second},
				{Name: "Snowdin", PBTime: 12*time.Minute + 40*time.Second, BestTime: 12*time.Minute + 10*time.Second},
				{Name: "Papyrus", PBTime: 15*time.Minute + 53*time.Second, BestTime: 15*time.Minute + 20*time.Second},
				{Name: "Spears 2", PBTime: 21*time.Minute + 43*time.Second, BestTime: 21*time.Minute + 0*time.Second},
				{Name: "Enter Lab", PBTime: 26*time.Minute + 31*time.Second, BestTime: 26*time.Minute + 0*time.Second},
				{Name: "Right Floor 2", PBTime: 28*time.Minute + 30*time.Second, BestTime: 28*time.Minute + 0*time.Second},
				{Name: "Left Floor 3", PBTime: 30*time.Minute + 3*time.Second, BestTime: 29*time.Minute + 30*time.Second},
				{Name: "Right Floor 3", PBTime: 31*time.Minute + 50*time.Second, BestTime: 31*time.Minute + 20*time.Second},
				{Name: "End", PBTime: 50*time.Minute + 59*time.Second, BestTime: 50*time.Minute + 0*time.Second},
			},
		}
	}
	defer f.Close()

	var s State
	json.NewDecoder(f).Decode(&s)
	if len(s.SplitFrames) != 0 && len(s.SplitFrames) != len(s.Splits) {
		panic(fmt.Sprintf("split_frames has %d entries but splits has %d", len(s.SplitFrames), len(s.Splits)))
	}
	return &s
}

// ── Fonts ─────────────────────────────────────────────────────────────────────

type Fonts struct {
	normal  *opentype.Font
	face    font.Face
	small   font.Face
	segment font.Face
	segCs   font.Face
	timer   font.Face
	timerCs font.Face
}

func loadFonts() *Fonts {
	data, err := os.ReadFile(fontPath)
	if err != nil {
		panic(err)
	}
	f, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}

	makeFace := func(size float64) font.Face {
		face, err := opentype.NewFace(f, &opentype.FaceOptions{
			Size: size,
			DPI:  72,
		})
		if err != nil {
			panic(err)
		}
		return face
	}

	return &Fonts{
		normal:  f,
		face:    makeFace(fontSize),
		small:   makeFace(smallSize),
		segment: makeFace(segmentTimerSize),
		segCs:   makeFace(segmentCsSize),
		timer:   makeFace(timerSize),
		timerCs: makeFace(timerCsSize),
	}
}

// ── Drawing helpers ───────────────────────────────────────────────────────────

func drawText(img *image.RGBA, f font.Face, x, y int, col color.RGBA, text string) {
	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: f,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

func drawTextRight(img *image.RGBA, f font.Face, x, y int, col color.RGBA, text string) {
	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: f,
	}
	w := d.MeasureString(text)
	d.Dot = fixed.P(x-int(w>>6), y)
	d.DrawString(text)
}

func drawTimerRight(img *image.RGBA, rightX, baseY int, col color.RGBA, mainFace, csFace font.Face, main, cs string) {
	csW := int((&font.Drawer{Face: csFace}).MeasureString(cs) >> 6)
	drawTextRight(img, csFace, rightX, baseY, col, cs)
	drawTextRight(img, mainFace, rightX-csW, baseY, col, main)
}

func fillRect(img *image.RGBA, x, y, w, h int, col color.RGBA) {
	rect := image.Rect(x, y, x+w, y+h)
	draw.Draw(img, rect, image.NewUniform(col), image.ZP, draw.Src)
}

func drawHLine(img *image.RGBA, y, x1, x2 int, col color.RGBA) {
	fillRect(img, x1, y, x2-x1, 1, col)
}

func faceAscent(f font.Face) int {
	return f.Metrics().Ascent.Ceil()
}

func faceDescent(f font.Face) int {
	return f.Metrics().Descent.Ceil()
}

// textBaselineCentered returns a baseline Y so text is vertically centered
// in [rowTop, rowTop+innerHeight), between separators above and below.
// layoutBottom returns the lowest Y coordinate used by the overlay for n splits.
func layoutBottom(n int, fonts *Fonts) int {
	y := headerPad
	line2 := y + fontSize*2 + 4
	y = line2 + faceDescent(fonts.face) + headerPad + 2
	y += n * rowHeight
	y++ // gap below split list
	y += faceAscent(fonts.timer) + faceDescent(fonts.timer)
	if n > 0 {
		y += 2 // gap below segment timer
	}
	statRow := faceAscent(fonts.face) + 1 + faceDescent(fonts.face) + statsGap
	y += 3 * statRow
	return y
}

func drawStatRow(img *image.RGBA, fonts *Fonts, y int, label, value string) int {
	baseY := y + faceAscent(fonts.face) + 1
	drawText(img, fonts.face, 8, baseY, colorText, label)
	drawTextRight(img, fonts.face, width-8, baseY, colorText, value)
	return y + faceAscent(fonts.face) + 1 + faceDescent(fonts.face) + statsGap
}

func textBaselineCentered(rowTop, innerHeight int, f font.Face) int {
	m := f.Metrics()
	a := m.Ascent.Ceil()
	d := m.Descent.Ceil()
	return rowTop + innerHeight/2 + (a-d)/2
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	cs := int(d.Milliseconds()/10) % 100
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%02d", h, m, s, cs)
	} else if m > 9 {
		return fmt.Sprintf("%02d:%02d.%02d", m, s, cs)
	} else if s > 9 {
		return fmt.Sprintf("%02d.%02d", s, cs)
	} else {
		return fmt.Sprintf("%2d.%02d", s, cs)
	}
}

func formatDurationCs(d time.Duration) (string, string) {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	cs := int(d.Milliseconds()/10) % 100
	var main string
	if h > 0 {
		main = fmt.Sprintf("%d:%02d:%02d", h, m, s)
	} else if m > 9 {
		main = fmt.Sprintf("%02d:%02d", m, s)
	} else if s > 9 {
		main = fmt.Sprintf("%02d", s)
	} else {
		main = fmt.Sprintf("%2d", s)
	}
	return main, fmt.Sprintf(".%02d", cs)
}

func currentSegmentTime(state *State) time.Duration {
	if state.CurrentSplit >= len(state.Splits) {
		return 0
	}
	if state.CurrentSplit == 0 {
		return state.CurrentTime
	}
	prev := &state.Splits[state.CurrentSplit-1]
	if prev.Completed {
		return state.CurrentTime - prev.SplitTime
	}
	return state.CurrentTime
}

func currentSegmentPB(state *State) time.Duration {
	if state.CurrentSplit >= len(state.Splits) {
		return 0
	}
	split := &state.Splits[state.CurrentSplit]
	if state.CurrentSplit == 0 {
		return split.PBTime
	}
	return split.PBTime - state.Splits[state.CurrentSplit-1].PBTime
}

func formatDiff(d time.Duration) (string, color.RGBA) {
	if d == 0 {
		return "-", colorText
	}
	sign := "+"
	col := colorRed
	if d < 0 {
		sign = "-"
		col = colorGreen
		d = -d
	}
	return sign + formatDuration(d), col
}

// ── Rendering ─────────────────────────────────────────────────────────────────

func render(img *image.RGBA, state *State, fonts *Fonts) {
	// Background
	draw.Draw(img, img.Rect, image.Black, image.ZP, draw.Src)

	y := headerPad

	// ── Header ────────────────────────────────────────────────────────────────
	line1 := y + fontSize + 2
	line2 := y + fontSize*2 + 4
	drawText(img, fonts.face, 8, line1, colorText, state.GameName)
	drawText(img, fonts.face, 8, line2, colorText, state.CategoryName)
	attempts := fmt.Sprintf("%d/%d", state.PBAttempts, state.TotalAttempts)
	drawTextRight(img, fonts.face, width-8, line1, colorText, attempts)
	y = line2 + faceDescent(fonts.face) + headerPad

	drawHLine(img, y, 0, width, colorSeparator)
	y += 2

	// ── Splits ────────────────────────────────────────────────────────────────
	splitsTop := y
	numSplits := len(state.Splits)

	for i := 0; i < numSplits; i++ {
		split := &state.Splits[i]
		rowY := splitsTop + i*rowHeight

		// Highlight current split
		if i == state.CurrentSplit && state.Running {
			fillRect(img, 0, rowY, width, rowHeight, colorHighlight)
		}

		// Text sits in the row above the bottom separator line.
		rowInner := rowHeight - 1
		nameY := textBaselineCentered(rowY, rowInner, fonts.face)
		timeY := textBaselineCentered(rowY, rowInner, fonts.small)
		diffY := textBaselineCentered(rowY, rowInner, fonts.small)

		// Split name
		drawText(img, fonts.face, 8+4, nameY, colorText, split.Name)

		// Split time / diff
		if split.Completed {
			diff := split.SplitTime - split.PBTime
			segTime := split.SplitTime
			if i > 0 {
				segTime -= state.Splits[i-1].SplitTime
			}
			diffStr, diffCol := formatDiff(diff)

			// Gold if best segment
			if segTime <= split.BestTime {
				diffCol = colorGold
				diffStr = formatDuration(segTime)
			}

			drawTextRight(img, fonts.face, width-260, diffY, diffCol, diffStr)
			drawTextRight(img, fonts.face, width-8, timeY, colorText, formatDuration(split.SplitTime))
		} else {
			drawTextRight(img, fonts.face, width-8, timeY, colorText, formatDuration(split.PBTime))
		}

		// drawHLine(img, rowY+rowHeight-1, 0, width, colorSeparator)
	}

	y = splitsTop + numSplits*rowHeight

	// drawHLine(img, y, 0, width, colorSeparator)
	y++ // gap above main timer (was 2px)

	// ── Main timer ────────────────────────────────────────────────────────────
	timerCol := colorTimerGreen
	if state.CurrentSplit < len(state.Splits) && state.Running {
		split := &state.Splits[state.CurrentSplit]
		if state.CurrentTime > split.PBTime {
			timerCol = colorTimerRed
		}
	}

	timerTop := y
	main, cs := formatDurationCs(state.CurrentTime)
	drawTimerRight(img, width-8, timerTop+faceAscent(fonts.timer)+timerPad, timerCol, fonts.timer, fonts.timerCs, main, cs)

	y = timerTop + faceAscent(fonts.timer) + faceDescent(fonts.timer)

	// ── Segment timer (directly below main timer, no separator) ───────────────
	segTime := currentSegmentTime(state)
	segPB := currentSegmentPB(state)
	segCol := colorText
	if segTime > segPB {
		segCol = colorTimerRed
	}

	segTop := timerTop + faceAscent(fonts.timer)
	segMain, segCs := formatDurationCs(segTime)
	drawTimerRight(img, width-8, segTop+faceAscent(fonts.segment)+timerPad, segCol, fonts.segment, fonts.segCs, segMain, segCs)

	y = segTop + faceAscent(fonts.segment) + faceDescent(fonts.segment)

	// Best possible time
	bpt := time.Duration(0)
	for i, s := range state.Splits {
		if i < state.CurrentSplit && s.Completed {
			seg := s.SplitTime
			if i > 0 {
				seg -= state.Splits[i-1].SplitTime
			}
			bpt += seg
		} else {
			bpt += s.BestTime
		}
	}

	// Previous segment
	prevSegStr := "-"
	if state.PreviousSegment > 0 {
		prevSegStr = formatDuration(state.PreviousSegment)
	}
	y = drawStatRow(img, fonts, y, "Previous Segment", prevSegStr)

	// Possible time save
	possibleSave := time.Duration(0)
	if state.CurrentSplit < len(state.Splits) {
		split := &state.Splits[state.CurrentSplit]
		seg := split.PBTime
		if state.CurrentSplit > 0 {
			seg -= state.Splits[state.CurrentSplit-1].PBTime
		}
		if seg > split.BestTime {
			possibleSave = seg - split.BestTime
		}
	}
	y = drawStatRow(img, fonts, y, "Possible Time Save", formatDuration(possibleSave))

	// Best possible time
	drawStatRow(img, fonts, y, "Best Possible Time", formatDuration(bpt))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkF(f func() error) {
	check(f())
}

type frameLine struct {
	bits uint32
	fps  int
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

func parseRaw(r io.Reader) []frameLine {
	s := bufio.NewScanner(r)
	defer checkF(s.Err)
	var lines []frameLine

	for s.Scan() {
		line := s.Text()
		segments := strings.Split(line, "|")

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
					// ignore unrecognised keys
				}
			}
		}

		lines = append(lines, frameLine{bits: b, fps: fps})
	}
	return lines
}

func expand(lines []frameLine) []uint32 {
	var bits []uint32
	fps120toggle := false
	for _, line := range lines {
		switch line.fps {
		case 30:
			bits = append(bits, line.bits, line.bits)
		case 60:
			bits = append(bits, line.bits)
		case 120:
			if !fps120toggle {
				bits = append(bits, line.bits)
			}
			fps120toggle = !fps120toggle
		}
	}
	return bits
}

func frameTimestamps(bits []uint32, rawLines []frameLine) []time.Duration {
	timestamps := make([]time.Duration, 0, len(bits))
	elapsed := time.Duration(0)

	for _, line := range rawLines {
		switch line.fps {
		case 30:
			frameDur := time.Second / 30
			timestamps = append(timestamps, elapsed, elapsed+frameDur/2)
			elapsed += frameDur
		case 60:
			frameDur := time.Second / 60
			timestamps = append(timestamps, elapsed)
			elapsed += frameDur
		case 120:
			frameDur := time.Second / 120
			if len(timestamps)%2 == 0 {
				timestamps = append(timestamps, elapsed)
			}
			elapsed += frameDur
		}
	}

	return timestamps
}

// ── Input handling ────────────────────────────────────────────────────────────

func handleInput(state *State, line string) {
	switch line {
	case "split":
		if !state.Running {
			state.Running = true
			state.StartTime = time.Now()
			state.CurrentSplit = 0
		} else if state.CurrentSplit < len(state.Splits) {
			state.Splits[state.CurrentSplit].Completed = true
			state.Splits[state.CurrentSplit].SplitTime = state.CurrentTime
			if state.CurrentSplit > 0 {
				seg := state.Splits[state.CurrentSplit].SplitTime - state.Splits[state.CurrentSplit-1].SplitTime
				state.PreviousSegment = seg
			}
			state.CurrentSplit++
			if state.CurrentSplit >= len(state.Splits) {
				state.Finished = true
				state.Running = false
			}
		}
	case "reset":
		state.Running = false
		state.Finished = false
		state.CurrentSplit = 0
		state.CurrentTime = 0
		state.PreviousSegment = 0
		for i := range state.Splits {
			state.Splits[i].Completed = false
			state.Splits[i].SplitTime = 0
		}
	case "undo":
		if state.CurrentSplit > 0 {
			state.CurrentSplit--
			state.Splits[state.CurrentSplit].Completed = false
			state.Splits[state.CurrentSplit].SplitTime = 0
			if state.CurrentSplit > 0 {
				state.PreviousSegment = state.Splits[state.CurrentSplit-1].SplitTime
			} else {
				state.PreviousSegment = 0
			}
		}
	}
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Rect, image.Black, image.ZP, draw.Src)
	os.Stdout.Write(img.Pix) // initial blank frame

	inputs, err := os.Open("tas/inputs")
	check(err)
	defer checkF(inputs.Close)

	rawLines := parseRaw(inputs)
	bits := expand(rawLines)
	timestamps := frameTimestamps(bits, rawLines)

	state := loadState("splits.json")
	fonts := loadFonts()
	if bottom := layoutBottom(len(state.Splits), fonts); bottom > height {
		panic(fmt.Sprintf("layout needs %dpx height but canvas is %dpx", bottom, height))
	}

	os.Stdin.Close()

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	for frame := range bits {
		<-ticker.C

		if frame < len(timestamps) {
			state.CurrentTime = timestamps[frame]
		}

		if frame == 0 && !state.Running {
			state.Running = true
		}

		if state.CurrentSplit < len(state.Splits) &&
			state.CurrentSplit < len(state.SplitFrames) &&
			state.SplitFrames[state.CurrentSplit] == frame {
			handleInput(state, "split")
		}

		render(img, state, fonts)
		os.Stdout.Write(img.Pix) // only written once per tick
	}
}
