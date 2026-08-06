package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"sync"
	"time"
)

const (
	width  = 120
	height = 40
	ttl    = 5 * time.Minute
)

var (
	store sync.Map
	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu    sync.Mutex
)

type captchaEntry struct {
	answer string
	expire time.Time
}

func Generate() (string, string) {
	mu.Lock()
	op := rng.Intn(2)
	var a, b, result int
	var opStr string

	switch op {
	case 0:
		a = rng.Intn(90) + 10
		b = rng.Intn(90) + 10
		result = a + b
		opStr = "+"
	default:
		a = rng.Intn(90) + 10
		b = rng.Intn(90) + 10
		if a < b {
			a, b = b, a
		}
		result = a - b
		opStr = "-"
	}

	id := fmt.Sprintf("%d", rng.Int63n(9999999999))
	answer := fmt.Sprintf("%d", result)

	entry := captchaEntry{
		answer: answer,
		expire: time.Now().Add(ttl),
	}
	store.Store(id, entry)
	mu.Unlock()

	go cleanup(id)

	question := fmt.Sprintf("%d %s %d = ?", a, opStr, b)
	img := renderImage(question)

	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	return id, imgBase64
}

func Verify(id string, answer string) bool {
	mu.Lock()
	defer mu.Unlock()

	val, ok := store.Load(id)
	if !ok {
		return false
	}
	entry := val.(captchaEntry)

	if time.Now().After(entry.expire) {
		store.Delete(id)
		return false
	}

	store.Delete(id)
	return entry.answer == answer
}

func Delete(id string) {
	store.Delete(id)
}

func cleanup(id string) {
	time.Sleep(ttl)
	store.Delete(id)
}

func renderImage(text string) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	bgColor := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, bgColor)
		}
	}

	for i := 0; i < 30; i++ {
		x1 := rng.Intn(width)
		y1 := rng.Intn(height)
		x2 := x1 + rng.Intn(15) + 5
		y2 := y1 + rng.Intn(3) + 1
		if x2 > width {
			x2 = width
		}
		if y2 > height {
			y2 = height
		}
		c := randomColor(180, 220)
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				img.SetRGBA(x, y, c)
			}
		}
	}

	for i := 0; i < 40; i++ {
		x := rng.Intn(width)
		y := rng.Intn(height)
		img.SetRGBA(x, y, randomColor(150, 200))
	}

	drawText(img, text)

	return img
}

func drawText(img *image.RGBA, text string) {
	fontColor := color.RGBA{R: 50, G: 50, B: 50, A: 255}
	x := 10
	y := 26

	for _, ch := range text {
		offsetX := rng.Intn(4) - 2
		offsetY := rng.Intn(4) - 2
		drawChar(img, ch, x+offsetX, y+offsetY, fontColor)
		x += 14
	}
}

func drawChar(img *image.RGBA, ch rune, x, y int, c color.RGBA) {
	patterns := map[rune][][]int{
		'0': {
			{0, 1, 1, 1, 0},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{0, 1, 1, 1, 0},
		},
		'1': {
			{0, 0, 1, 0, 0},
			{0, 1, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 1, 1, 1, 0},
		},
		'2': {
			{0, 1, 1, 1, 0},
			{1, 0, 0, 0, 1},
			{0, 0, 0, 0, 1},
			{0, 0, 0, 1, 0},
			{0, 0, 1, 0, 0},
			{0, 1, 0, 0, 0},
			{1, 1, 1, 1, 1},
		},
		'3': {
			{1, 1, 1, 1, 0},
			{0, 0, 0, 0, 1},
			{0, 0, 0, 0, 1},
			{0, 1, 1, 1, 0},
			{0, 0, 0, 0, 1},
			{0, 0, 0, 0, 1},
			{1, 1, 1, 1, 0},
		},
		'4': {
			{0, 0, 0, 1, 0},
			{0, 0, 1, 1, 0},
			{0, 1, 0, 1, 0},
			{1, 0, 0, 1, 0},
			{1, 1, 1, 1, 1},
			{0, 0, 0, 1, 0},
			{0, 0, 0, 1, 0},
		},
		'5': {
			{1, 1, 1, 1, 1},
			{1, 0, 0, 0, 0},
			{1, 1, 1, 1, 0},
			{0, 0, 0, 0, 1},
			{0, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{0, 1, 1, 1, 0},
		},
		'6': {
			{0, 1, 1, 1, 0},
			{1, 0, 0, 0, 0},
			{1, 0, 0, 0, 0},
			{1, 1, 1, 1, 0},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{0, 1, 1, 1, 0},
		},
		'7': {
			{1, 1, 1, 1, 1},
			{0, 0, 0, 0, 1},
			{0, 0, 0, 1, 0},
			{0, 0, 1, 0, 0},
			{0, 1, 0, 0, 0},
			{0, 1, 0, 0, 0},
			{0, 1, 0, 0, 0},
		},
		'8': {
			{0, 1, 1, 1, 0},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{0, 1, 1, 1, 0},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{0, 1, 1, 1, 0},
		},
		'9': {
			{0, 1, 1, 1, 0},
			{1, 0, 0, 0, 1},
			{1, 0, 0, 0, 1},
			{0, 1, 1, 1, 1},
			{0, 0, 0, 0, 1},
			{0, 0, 0, 0, 1},
			{0, 1, 1, 1, 0},
		},
		'+': {
			{0, 0, 0, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{1, 1, 1, 1, 1},
			{0, 0, 1, 0, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 0, 0, 0},
		},
		'-': {
			{0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0},
			{1, 1, 1, 1, 1},
			{0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0},
		},
		'=': {
			{0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0},
			{1, 1, 1, 1, 1},
			{0, 0, 0, 0, 0},
			{1, 1, 1, 1, 1},
			{0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0},
		},
		'?': {
			{0, 1, 1, 1, 0},
			{1, 0, 0, 0, 1},
			{0, 0, 0, 0, 1},
			{0, 0, 0, 1, 0},
			{0, 0, 1, 0, 0},
			{0, 0, 0, 0, 0},
			{0, 0, 1, 0, 0},
		},
	}

	pattern, ok := patterns[ch]
	if !ok {
		return
	}

	for row, line := range pattern {
		for col, v := range line {
			if v == 1 {
				px := x + col
				py := y + row
				if px >= 0 && px < width && py >= 0 && py < height {
					img.SetRGBA(px, py, c)
				}
			}
		}
	}
}

func randomColor(min, max uint8) color.RGBA {
	r := uint8(int(min) + rng.Intn(int(max-min)))
	g := uint8(int(min) + rng.Intn(int(max-min)))
	b := uint8(int(min) + rng.Intn(int(max-min)))
	return color.RGBA{R: r, G: g, B: b, A: 255}
}