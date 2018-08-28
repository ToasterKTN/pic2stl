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
	bwimg := effect.Grayscale(img)
	if settings.Invert {
		oldimg := img
		img = effect.Invert(oldimg)
	}
	output.WriteString("solid Converted\n")
	wg.Add(bwimg.Bounds().Size().X - 1)
	lines := make(chan string)
	for x := 1; x < bwimg.Bounds().Size().X; x++ {
		go func(x int) {
			defer wg.Done()
			for y := 1; y < bwimg.Bounds().Size().Y; y++ {
				var answer = ""
				answer += "facet normal 0 0 0\n"
				answer += "outer loop\n"
				answer += fmt.Sprintf("vertex %d %d %d \n", x, y, uint8(float64(bwimg.GrayAt(x, y).Y)*settings.ZDivider))
				answer += fmt.Sprintf("vertex %d %d %d \n", x-1, y, uint8(float64(bwimg.GrayAt(x-1, y).Y)*settings.ZDivider))
				answer += fmt.Sprintf("vertex %d %d %d \n", x, y-1, uint8(float64(bwimg.GrayAt(x, y-1).Y)*settings.ZDivider))
				answer += "endloop\n"
				answer += "endfacet\n"
				answer += "facet normal 0 0 0\n"
				answer += "outer loop\n"
				answer += fmt.Sprintf("vertex %d %d %d \n", x-1, y, uint8(float64(bwimg.GrayAt(x-1, y).Y)*settings.ZDivider))
				answer += fmt.Sprintf("vertex %d %d %d \n", x-1, y-1, uint8(float64(bwimg.GrayAt(x-1, y-1).Y)*settings.ZDivider))
				answer += fmt.Sprintf("vertex %d %d %d \n", x, y-1, uint8(float64(bwimg.GrayAt(x, y-1).Y)*settings.ZDivider))
				answer += "endloop\n"
				answer += "endfacet\n"
				lines <- answer
			}
		}(x)
	}

	go func() {
		for a := range lines {
			output.WriteString(a)
		}
	}()

	wg.Wait()
	output.WriteString("endsolid Converted\n")
}
