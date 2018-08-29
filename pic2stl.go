package main

import (
	"fmt"
	"os"
	"sync"

	flag "github.com/namsral/flag"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
)

type Settings struct {
	Input    string
	Output   string
	ZDivider float64
	ResizeX  int
	ResizeY  int
	Invert   bool
}

type Point struct {
	X int
	Y int
	Z float64
}

var settings Settings

var wg sync.WaitGroup

func initFlags(fs *flag.FlagSet) {
	fs.StringVar(&settings.Input, "Input", "image.png", "Image to be procesed as Input")
	fs.StringVar(&settings.Output, "Output", "output.stl", "STL Filename to be used as Output")
	fs.Float64Var(&settings.ZDivider, "ZDivider", 1, "Factor to Resize Z.")
	fs.IntVar(&settings.ResizeX, "ResizeX", 0, "Resize X to # pixels")
	fs.IntVar(&settings.ResizeY, "ResizeY", 0, "Resize Y to # pixels")
	fs.BoolVar(&settings.Invert, "Invert", false, "Invert the Image")
	fs.Parse(os.Args[1:])
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fs := flag.FlagSet{}
	initFlags(&fs)
	img, err := imgio.Open(settings.Input)
	check(err)
	output, err := os.Create(settings.Output)
	check(err)
	defer output.Close()
	if settings.ResizeX != 0 || settings.ResizeY != 0 {
		if settings.ResizeX == 0 {
			settings.ResizeX = int(float64(settings.ResizeY) * float64(img.Bounds().Size().Y) / float64(img.Bounds().Size().X))
		}
		if settings.ResizeY == 0 {
			settings.ResizeY = int(float64(settings.ResizeX) * float64(img.Bounds().Size().X) / float64(img.Bounds().Size().Y))
		}
		oldimg := img
		img = transform.Resize(oldimg, settings.ResizeX, settings.ResizeY, transform.Linear)
	}
	if settings.Invert {
		oldimg := img
		img = effect.Invert(oldimg)
	}
	bwimg := effect.Grayscale(img)
	output.WriteString("solid Converted\n")
	// Image
	wg.Add(bwimg.Bounds().Size().X - 1)
	lines := make(chan string)
	for x := 1; x < bwimg.Bounds().Size().X; x++ {
		go func(x int) {
			defer wg.Done()
			for y := 1; y < bwimg.Bounds().Size().Y; y++ {
				lines <- createFacet(
					Point{x, y, 1 + float64(bwimg.GrayAt(x, y).Y)*settings.ZDivider},
					Point{x, y - 1, 1 + float64(bwimg.GrayAt(x, y-1).Y)*settings.ZDivider},
					Point{x - 1, y, 1 + float64(bwimg.GrayAt(x-1, y).Y)*settings.ZDivider})
				lines <- createFacet(
					Point{x - 1, y, 1 + float64(bwimg.GrayAt(x-1, y).Y)*settings.ZDivider},
					Point{x, y - 1, 1 + float64(bwimg.GrayAt(x, y-1).Y)*settings.ZDivider},
					Point{x - 1, y - 1, 1 + float64(bwimg.GrayAt(x-1, y-1).Y)*settings.ZDivider})
			}
		}(x)
	}

	// Add our Lines to the Output File
	go func() {
		for {
			a := <-lines
			if a == "quit" {
				wg.Done()
				return
			}
			output.WriteString(a)
		}
	}()

	wg.Wait()
	// Sides
	for x := 1; x < bwimg.Bounds().Size().X; x++ {
		lines <- createFacet(
			Point{x, 0, 0},
			Point{x - 1, 0, 1 + float64(bwimg.GrayAt(x-1, 0).Y)*settings.ZDivider},
			Point{x, 0, 1 + float64(bwimg.GrayAt(x, 0).Y)*settings.ZDivider})
		lines <- createFacet(
			Point{x, 0, 0},
			Point{x - 1, 0, 0},
			Point{x - 1, 0, 1 + float64(bwimg.GrayAt(x-1, 0).Y)*settings.ZDivider})
		lines <- createFacet(
			Point{x, bwimg.Bounds().Size().Y - 1, 0},
			Point{x, bwimg.Bounds().Size().Y - 1, 1 + float64(bwimg.GrayAt(x, bwimg.Bounds().Size().Y-1).Y)*settings.ZDivider},
			Point{x - 1, bwimg.Bounds().Size().Y - 1, 1 + float64(bwimg.GrayAt(x-1, bwimg.Bounds().Size().Y-1).Y)*settings.ZDivider})
		lines <- createFacet(
			Point{x, bwimg.Bounds().Size().Y - 1, 0},
			Point{x - 1, bwimg.Bounds().Size().Y - 1, 1 + float64(bwimg.GrayAt(x-1, bwimg.Bounds().Size().Y-1).Y)*settings.ZDivider},
			Point{x - 1, bwimg.Bounds().Size().Y - 1, 0})
	}
	for y := 1; y < bwimg.Bounds().Size().Y; y++ {
		lines <- createFacet(
			Point{0, y, 0},
			Point{0, y, 1 + float64(bwimg.GrayAt(0, y).Y)*settings.ZDivider},
			Point{0, y - 1, 1 + float64(bwimg.GrayAt(0, y-1).Y)*settings.ZDivider})
		lines <- createFacet(
			Point{0, y, 0},
			Point{0, y - 1, 1 + float64(bwimg.GrayAt(0, y-1).Y)*settings.ZDivider},
			Point{0, y - 1, 0})
		lines <- createFacet(
			Point{bwimg.Bounds().Size().X - 1, y, 0},
			Point{bwimg.Bounds().Size().X - 1, y - 1, 1 + float64(bwimg.GrayAt(0, y-1).Y)*settings.ZDivider},
			Point{bwimg.Bounds().Size().X - 1, y, 1 + float64(bwimg.GrayAt(0, y).Y)*settings.ZDivider})
		lines <- createFacet(
			Point{bwimg.Bounds().Size().X - 1, y, 0},
			Point{bwimg.Bounds().Size().X - 1, y - 1, 0},
			Point{bwimg.Bounds().Size().X - 1, y - 1, 1 + float64(bwimg.GrayAt(0, y-1).Y)*settings.ZDivider})
	}
	// Floor
	centerx := bwimg.Bounds().Size().X / 2
	centery := bwimg.Bounds().Size().Y / 2
	for x := 1; x < bwimg.Bounds().Size().X; x++ {
		lines <- createFacet(
			Point{x, 0, 0},
			Point{centerx, centery, 0},
			Point{x - 1, 0, 0})
		lines <- createFacet(
			Point{x, bwimg.Bounds().Size().Y - 1, 0},
			Point{x - 1, bwimg.Bounds().Size().Y - 1, 0},
			Point{centerx, centery, 0})
	}
	for y := 1; y < bwimg.Bounds().Size().Y; y++ {
		lines <- createFacet(
			Point{0, y, 0},
			Point{0, y - 1, 0},
			Point{centerx, centery, 0})
		lines <- createFacet(
			Point{bwimg.Bounds().Size().X - 1, y, 0},
			Point{centerx, centery, 0},
			Point{bwimg.Bounds().Size().X - 1, y - 1, 0})
	}
	lines <- "endsolid Converted\n"

	// Wait for all the Lines to be written.. Maybe unneeded but i dont know.
	wg.Add(1)
	lines <- "quit"
	wg.Wait()
}

func createFacet(p1, p2, p3 Point) string {
	answer := "facet normal 0 0 0\n"
	answer += "outer loop\n"
	answer += fmt.Sprintf("vertex %d %d %f.4 \n", p1.X, p1.Y, p1.Z)
	answer += fmt.Sprintf("vertex %d %d  %f.4\n", p2.X, p2.Y, p2.Z)
	answer += fmt.Sprintf("vertex %d %d %f.4 \n", p3.X, p3.Y, p3.Z)
	answer += "endloop\n"
	answer += "endfacet\n"
	return answer
}
