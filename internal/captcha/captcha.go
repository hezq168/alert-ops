package captcha

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/big"
	"strings"
	"sync"
	"time"
)

const (
	width       = 120
	height      = 40
	codeLength  = 4
	fontSize    = 22
	expireTime  = 30 * time.Second
	cleanupTime = 10 * time.Minute
)

// Store 验证码存储
type Store struct {
	mu    sync.RWMutex
	codes map[string]captchaItem
}

type captchaItem struct {
	code      string
	expiresAt time.Time
}

var defaultStore = &Store{codes: make(map[string]captchaItem)}

func init() {
	go func() {
		for {
			time.Sleep(cleanupTime)
			defaultStore.CleanExpired()
		}
	}()
}

// CleanExpired 清理过期验证码
func (s *Store) CleanExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, item := range s.codes {
		if now.After(item.expiresAt) {
			delete(s.codes, id)
		}
	}
}

// Generate 生成验证码，返回 id、base64 图片、验证码字符串
func Generate() (id, b64Img, code string, err error) {
	code = randomCode(codeLength)
	id = randomID()

	defaultStore.mu.Lock()
	defaultStore.codes[id] = captchaItem{code: code, expiresAt: time.Now().Add(expireTime)}
	defaultStore.mu.Unlock()

	img := drawImage(code)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return "", "", "", err
	}
	b64Img = "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	return id, b64Img, code, nil
}

// Verify 校验验证码（一次性消费）
func Verify(id, code string) bool {
	defaultStore.mu.Lock()
	defer defaultStore.mu.Unlock()

	item, ok := defaultStore.codes[id]
	if !ok {
		return false
	}
	delete(defaultStore.codes, id) // 一次性消费
	if time.Now().After(item.expiresAt) {
		return false
	}
	return strings.EqualFold(item.code, code)
}

func randomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func randomCode(length int) string {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

func drawImage(code string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 背景
	bgColor := color.RGBA{240, 240, 240, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// 干扰线
	for i := 0; i < 5; i++ {
		x1 := randInt(width)
		y1 := randInt(height)
		x2 := randInt(width)
		y2 := randInt(height)
		drawLine(img, x1, y1, x2, y2, color.RGBA{180, 180, 180, 255})
	}

	// 干扰点
	for i := 0; i < 60; i++ {
		img.Set(randInt(width), randInt(height), color.RGBA{160, 160, 160, 255})
	}

	// 绘制字符（简单位图）
	charColors := []color.RGBA{
		{200, 50, 50, 255},
		{50, 50, 200, 255},
		{50, 150, 50, 255},
		{180, 100, 30, 255},
	}
	for i, ch := range code {
		col := charColors[i%len(charColors)]
		drawChar(img, ch, 15+i*fontSize, 8, col)
	}

	return img
}

// 简单的 5x7 点阵字符绘制
var fontMap = map[rune][]string{
	'0': {"01110", "10001", "10011", "10101", "11001", "10001", "01110"},
	'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
	'2': {"01110", "10001", "00001", "00110", "01000", "10000", "11111"},
	'3': {"01110", "10001", "00001", "00110", "00001", "10001", "01110"},
	'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
	'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
	'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
	'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
	'C': {"01110", "10001", "10000", "10000", "10000", "10001", "01110"},
	'D': {"11100", "10010", "10001", "10001", "10001", "10010", "11100"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
	'G': {"01110", "10001", "10000", "10111", "10001", "10001", "01110"},
	'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'J': {"11111", "00010", "00010", "00010", "00010", "10010", "01100"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'M': {"10001", "11011", "10101", "10001", "10001", "10001", "10001"},
	'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
	'Q': {"01110", "10001", "10001", "10001", "10101", "10010", "01101"},
	'R': {"11110", "10001", "10001", "11110", "10010", "10001", "10001"},
	'S': {"01110", "10001", "10000", "01110", "00001", "10001", "01110"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
	'V': {"10001", "10001", "10001", "10001", "01010", "01010", "00100"},
	'W': {"10001", "10001", "10001", "10101", "10101", "11011", "10001"},
	'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
	'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
}

func drawChar(img *image.RGBA, ch rune, x0, y0 int, col color.RGBA) {
	rows, ok := fontMap[ch]
	if !ok {
		return
	}
	scale := 3 // 像素缩放
	for rowIdx, row := range rows {
		for colIdx, bit := range row {
			if bit == '1' {
				for dx := 0; dx < scale; dx++ {
					for dy := 0; dy < scale; dy++ {
						px := x0 + colIdx*scale + dx
						py := y0 + rowIdx*scale + dy
						if px >= 0 && px < width && py >= 0 && py < height {
							img.Set(px, py, col)
						}
					}
				}
			}
		}
	}
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		if x1 >= 0 && x1 < width && y1 >= 0 && y1 < height {
			img.Set(x1, y1, col)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func randInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}
