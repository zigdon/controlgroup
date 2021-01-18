package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

func loadFiles(files []string) []string {
	words := []string{}
	for _, i := range files {
		data, err := ioutil.ReadFile(i)
		if err != nil {
			log.Fatalf("can't read %q: %v", i, err)
		}
		words = append(words, strings.Split(string(data), "\n")...)
	}
	return words
}

// Board represents the entire wordsearch collection
type Board struct {
	points        int
	words         []string
	letters       []rune
	width, height int
}

func newBoard(w, h int) *Board {
	if w != h {
		log.Fatalf("only supporting square boards")
	}
	return &Board{
		width:   w,
		height:  h,
		points:  0,
		words:   []string{},
		letters: make([]rune, w*h),
	}
}

// Set sets a letter in the board
func (b *Board) Set(x, y int, v rune) {
	b.letters[y*b.width+x] = v
}

// Get reads a letter from the board
func (b *Board) Get(x, y int) rune {
	return b.letters[y*b.width+x]
}

type Direction int

func (d Direction) String() string {
	return [...]string{"right", "left", "down", "up", "downRight", "downLeft", "upRight", "upLeft"}[d]
}

const (
	right Direction = iota
	left
	down
	up
	downRight
	downLeft
	upRight
	upLeft
)

type coord struct {
	x, y int
}

func (b *Board) FindBlanks() []coord {
	res := []coord{}
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			if b.Get(x, y) > 0 {
				continue
			}
			res = append(res, coord{x, y})
		}
	}
	return res
}

type fills struct {
	x, y int
	opts []rune
}

func (b *Board) Fill(allowed string) bool {
	tmpl := make(map[rune]bool)
	for _, c := range allowed {
		tmpl[c] = false
	}

	blanks := b.FindBlanks()
	fs := []fills{}
	for _, blank := range blanks {
		x, y := blank.x, blank.y
		seen := make(map[rune]bool)
		for k, v := range tmpl {
			seen[k] = v
		}
		for i := 0; i < b.width; i++ {
			c := b.Get(x, i)
			if c > 0 {
				seen[c] = true
			}
			c = b.Get(i, y)
			if c > 0 {
				seen[c] = true
			}
		}
		f := fills{x, y, []rune{}}
		for k, v := range seen {
			if v {
				continue
			}
			f.opts = append(f.opts, k)
		}
		fs = append(fs, f)
	}

	sort.Slice(fs, func(i, j int) bool {
		return len(fs[i].opts) < len(fs[j].opts)
	})
	log.Printf("Trying to fill %d blanks:", len(fs))
	for _, f := range fs {
		log.Printf("  %v", f)
	}
	addedX := make(map[int]map[rune]bool)
	addedY := make(map[int]map[rune]bool)
	for _, f := range fs {
		fixed := false
		for _, opt := range f.opts {
			if addedX[f.x] == nil {
				addedX[f.x] = make(map[rune]bool)
			}
			if addedY[f.y] == nil {
				addedY[f.y] = make(map[rune]bool)
			}
			if !addedX[f.x][opt] && !addedY[f.y][opt] {
				log.Printf("Filling (%d,%d) with %c", f.x, f.y, opt)
				b.Set(f.x, f.y, opt)
				addedX[f.x][opt] = true
				addedY[f.y][opt] = true
				fixed = true
				break
			}
		}
		if !fixed {
			b.Print()
			log.Fatalf("Giving up on filling (%d,%d)!", f.x, f.y)
		}
	}

	return true
}

func (b *Board) Print() {
	var out strings.Builder
	for i, c := range b.letters {
		if i%b.width == 0 {
			fmt.Fprintf(&out, "\n%d: ", i/b.width)
		}
		fmt.Fprintf(&out, " %c", c)
	}
	log.Print(out.String())
	log.Printf("Words: %v", b.words)
	log.Printf("Value: %d", b.points)
}

func (b *Board) Dump() {
	for i, c := range b.letters {
		if i%b.width == 0 {
			fmt.Println()
		}
		fmt.Printf("%c", c)
	}
	fmt.Print("\n\n")
	for _, k := range b.words {
		fmt.Println(k)
	}
	log.Printf("Value: %d", b.points)
}

func (b *Board) Place(word string, dir Direction) bool {
	var sx, ex, sy, ey, dx, dy int
	sx, sy, ex, ey = 0, 0, b.width-1, b.height-1
	switch dir {
	case right:
		dx = 1
		dy = 0
		ex = b.width - len(word)
	case left:
		dx = -1
		dy = 0
		sx = len(word)
	case down:
		dx = 0
		dy = 1
		ey = b.height - len(word)
	case up:
		dx = 0
		dy = -1
		sy = len(word)
	case downRight:
		dx = 1
		dy = 1
		ex = b.width - len(word)
		ey = b.height - len(word)
	case downLeft:
		dx = -1
		dy = 1
		sx = len(word)
		ey = b.height - len(word)
	case upRight:
		dx = 1
		dy = -1
		ex = b.width - len(word)
		sy = len(word)
	case upLeft:
		dx = -1
		dy = -1
		sx = len(word)
		sy = len(word)
	}

	var r rune
	for x := sx; x <= ex; x++ {
		for y := sy; y <= ey; y++ {
			good := true

			// Check if it fits on the board
			for i, c := range word {
				r = b.Get(x+i*dx, y+i*dy)

				// Already there? we're good
				if r == c {
					continue
				}

				// Something else there, no fit
				if r != rune(0) {
					good = false
					break
				}

				// Check if it would duplicate a letter in the column or row
				for dup := 0; dup < b.width; dup++ {
					if b.Get(dup, y+i*dy) == c || b.Get(x+i*dx, dup) == c {
						good = false
						break
					}
				}
				if !good {
					break
				}
			}

			if !good {
				continue
			}

			for i, c := range word {
				b.Set(x+i*dx, y+i*dy, c)
			}
			b.points = b.points + len(word) - 3
			b.words = append(b.words, word)
			log.Printf("Placed %q at (%d,%d) heading %s.", word, sx, sy, dir)
			b.Print()
			return true
		}
	}

	return false
}

func checkDouble(word string, dir Direction) bool {
	good := true
	needsCheck := false
	for _, d := range []Direction{up, down, left, right} {
		if dir == d {
			needsCheck = true
			break
		}
	}
	if !needsCheck {
		return false
	}
	counts := make(map[rune]int)
	for _, c := range word {
		counts[c]++
		if counts[c] > 1 {
			good = false
			break
		}
	}
	if !good {
		return true
	}
	return false
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	dirs := []Direction{right, left, up, down, downRight, downLeft, upLeft, upRight}

	words := loadFiles(os.Args[1:])
	log.Printf("loaded %d words", len(words))
	rand.Shuffle(len(words), func(i int, j int) {
		words[i], words[j] = words[j], words[i]
	})

	b := newBoard(8, 8)
	for _, w := range words {
		// check if this is a substring of any of the existing words, or any of the
		// words are substrings of this
		skip := false
		rev := Reverse(w)
		for _, existing := range b.words {
			if strings.Contains(w, existing) || strings.Contains(existing, w) {
				log.Printf("Skipping %q, overlapping with %q", w, existing)
				skip = true
				break
			}
			if strings.Contains(rev, existing) || strings.Contains(existing, rev) {
				log.Printf("Skipping %q (%q reversed), overlapping with %q", rev, w, existing)
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		rand.Shuffle(len(dirs), func(i int, j int) {
			dirs[i], dirs[j] = dirs[j], dirs[i]
		})
		for _, d := range dirs {
			fmt.Printf("%10s %s\r", w, d)
			if checkDouble(w, d) {
				continue
			}
			if b.Place(w, d) {
				if b.points >= 28 && len(b.FindBlanks()) == 0 {
					os.Exit(0)
				}
				break
			}
		}
	}

	log.Print("Out of words")
	b.Fill("ACDEGOPRSUV")
	b.Dump()

	log.Printf("Considered %d words, %q .. %q", len(words), words[0], words[len(words)-1])
}
