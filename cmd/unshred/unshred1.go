// A solution to the Instagram Engineering Challenge (http://goo.gl/B92t0)
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
	"sort"
)

const shredWidth = 32

func fatal(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s, args...)
	os.Exit(2)
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

// shuffleImage returns an image with shreds in m reordered by shuffle. 
func shuffleImage(m image.Image, shuffle []int) image.Image {
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

// edge represents a possible connection between two shreds.
type edge struct {
	// Index of shreds to left and right of this edge.
	left, right int
	// Measure of difference between shreds at this edge.
	diff float64
}

// shred represents a shred.
type shred struct {
	// Index of shred joined to the left and right of this shred or -1 to
	// indicate that no shred is joined.
	left, right int
}

// byDiff sorts shreds by the diff field.
type byDiff []edge

func (edges byDiff) Len() int           { return len(edges) }
func (edges byDiff) Less(i, j int) bool { return edges[i].diff < edges[j].diff }
func (edges byDiff) Swap(i, j int)      { edges[i], edges[j] = edges[j], edges[i] }

func chanDiffSqr(c1, c2 uint32) float64 {
	d := math.Log10(float64(c1+1)/65535.0) - math.Log10(float64(c2+1)/65535.0)
	return d * d
}

// computeShuffle returns a slice that maps source shred index to destination
// shred index. 
func computeShuffle(m image.Image) []int {
	b := m.Bounds()
	n := (b.Max.X - b.Min.X) / shredWidth

	// Create an edge for all shred combinations.
	edges := make([]edge, 0, n*n-n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			x0 := b.Min.X + i*shredWidth + shredWidth - 1
			x1 := b.Min.X + j*shredWidth
			// Compute difference between edges as sum of euclidian distance
			// between adjacent pixels in RGB space.
			var diff float64
			for y := b.Min.Y; y < b.Max.X; y++ {
				r0, g0, b0, _ := m.At(x0, y).RGBA()
				r1, g1, b1, _ := m.At(x1, y).RGBA()
				diff += math.Sqrt(chanDiffSqr(r0, r1) + chanDiffSqr(g0, g1) + chanDiffSqr(b0, b1))
			}
			edges = append(edges, edge{i, j, diff})
		}
	}

	sort.Sort(byDiff(edges))

	// Initialize shreds with left and right set to -1
	shreds := make([]shred, n)
	for i := range shreds {
		shreds[i] = shred{-1, -1}
	}

	// Join shreds in edge quality order.
	joins := 0
	for _, e := range edges {
		if shreds[e.left].right >= 0 || shreds[e.right].left >= 0 {
			// One or both shreds for this edge are connected.
			continue
		}
		shreds[e.left].right = e.right
		shreds[e.right].left = e.left
		joins += 1
		if joins == n-1 {
			// Exit to avoid joining shreds in a loop.
			break
		}
	}

	// Find shred on the left edge of output.
	i := 0
	for ; i < n; i++ {
		if shreds[i].left < 0 {
			break
		}
	}

	// Compute shuffle by iterating from left to right through shreds.
	shuffle := make([]int, 0, n)
	for ; i >= 0; i = shreds[i].right {
		shuffle = append(shuffle, i)
	}

	return shuffle
}

func main() {
	if len(os.Args) != 3 {
		fatal("usage: unshred input output")
	}
	m := readImage(os.Args[1])
	shuffle := computeShuffle(m)
	m = shuffleImage(m, shuffle)
	writeImage(os.Args[2], m)
}
