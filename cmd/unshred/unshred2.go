package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path"
)

func fatal(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s, args...)
	os.Exit(2)
}

// chanDiff returns a measure of the difference between two color channel
// values.
func chanDiff(c1, c2 uint32) float64 {
	d := float64(c1) - float64(c2)
	return d * d
}

// computeDiff computes a measurment of the difference between shred edges.
// The (i + shredCount * j) element of the result is the difference between the
// right edge of shred i and the left edge of shred j.
func computeDiff(m image.Image, shredWidth, shredCount int) []float64 {
	diff := make([]float64, shredCount*shredCount)
	b := m.Bounds()
	for i := 0; i < shredCount; i++ {
		for j := 0; j < shredCount; j++ {
			if i == j {
				continue
			}
			x0 := b.Min.X + i*shredWidth + shredWidth - 1
			x1 := b.Min.X + j*shredWidth
			var d float64
			for y := b.Min.Y; y < b.Max.X; y++ {
				r0, g0, b0, _ := m.At(x0, y).RGBA()
				r1, g1, b1, _ := m.At(x1, y).RGBA()
				//d += chanDiff(r0, r1)
				//d += chanDiff(g0, g1)
				//d += chanDiff(b0, b1)
				d += math.Sqrt(chanDiff(r0, r1) + chanDiff(g0, g1) + chanDiff(b0, b1))
			}
			diff[i+shredCount*j] = d
		}
	}
	return diff
}

// computeShuffle returns a slice that maps source shred index to destination
// shred index. 
func computeShuffle(diff []float64, shredCount int) []int {
	shuffle := make([]int, shredCount)

	// Add best match to result.
	d := float64(math.MaxFloat64)
	for i := 0; i < shredCount; i++ {
		for j := 0; j < shredCount; j++ {
			if i != j && d > diff[i+shredCount*j] {
				d = diff[i+shredCount*j]
				shuffle[0] = i
				shuffle[1] = j
			}
		}
	}

	// inShuffle[i] is true if shred i is in result.
	inShuffle := make([]bool, shredCount)
	inShuffle[shuffle[0]] = true
	inShuffle[shuffle[1]] = true

	var order []int
	var diffs []float64
	order = append(order, -shuffle[0])
	order = append(order, shuffle[1])
	diffs = append(diffs, d)
	diffs = append(diffs, d)

	// Candidate shreds for left and right edges of result. 
	left, right := -1, -1

	// Difference between candidate shreds and the shuffled edges.
	var diffLeft, diffRight float64

	for k := 2; k < shredCount; k++ {
		if left < 0 {
			// Find new candidate for left side.
			diffLeft = math.MaxFloat64
			j := shuffle[0]
			for i := 0; i < shredCount; i++ {
				if !inShuffle[i] && diffLeft > diff[i+shredCount*j] {
					left = i
					diffLeft = diff[i+shredCount*j]
				}
			}
		}
		if right < 0 {
			// Find new candidate for right side.
			diffRight = math.MaxFloat64
			i := shuffle[k-1]
			for j := 0; j < shredCount; j++ {
				if !inShuffle[j] && diffRight > diff[i+shredCount*j] {
					right = j
					diffRight = diff[i+shredCount*j]
				}
			}
		}

		// Add best candidate to result.
		if diffLeft < diffRight {
			copy(shuffle[1:], shuffle[0:k])
			shuffle[0] = left
			inShuffle[left] = true
			order = append(order, -left)
			diffs = append(diffs, diffLeft)
			left = -1
		} else {
			shuffle[k] = right
			inShuffle[right] = true
			order = append(order, right)
			diffs = append(diffs, diffRight)
			right = -1
		}
	}
	for i := range order {
		fmt.Println(order[i], diffs[i])
	}
	fmt.Println(shuffle)
	return shuffle
}

// shuffleImage returns new image with m shreds reordered according to shuffle.
func shuffleImage(m image.Image, shredWidth int, shuffle []int) image.Image {
	b := m.Bounds()
	result := image.NewRGBA(image.Rect(0, 0, b.Max.X-b.Min.X, b.Max.Y-b.Min.Y))
	for dst, src := range shuffle {
		draw.Draw(result,
			image.Rect(dst*shredWidth, 0, dst*shredWidth+shredWidth, b.Max.Y-b.Min.Y),
			m,
			image.Pt(src*shredWidth, 0),
			draw.Src)
	}
	return result
}

func readImage(fname string) image.Image {
	f, err := os.Open(fname)
	if err != nil {
		fatal("Error opening input: %v\n", err)
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	if err != nil {
		fatal("Error reading input: %v\n", err)
	}
	return m
}

func writeImage(fname string, m image.Image) {
	var encode func(io.Writer, image.Image) error
	switch path.Ext(fname) {
	case ".jpg":
		encode = func(w io.Writer, m image.Image) error {
			return jpeg.Encode(w, m, nil)
		}
	case ".png":
		encode = png.Encode
	default:
		fatal("Unknown output type")
	}

	f, err := os.Create(fname)
	if err != nil {
		fatal("Error creating output: %v\n", err)
	}
	defer f.Close()

	if err := encode(f, m); err != nil {
		fatal("Error encoding output: %v\n", err)
	}
}

func main() {
	if len(os.Args) != 3 {
		fatal("usage: unshred input output")
	}

	m := readImage(os.Args[1])

	const shredWidth = 32
	b := m.Bounds()
	shredCount := (b.Max.X - b.Min.X) / shredWidth

	diff := computeDiff(m, shredWidth, shredCount)
	shuffle := computeShuffle(diff, shredCount)
	m = shuffleImage(m, shredWidth, shuffle)

	writeImage(os.Args[2], m)
}
