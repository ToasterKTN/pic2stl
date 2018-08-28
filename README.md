# pic2stl
A Simple Go code, to convert a image into a stl mesh, to engrave it or use it otherwise

## Installation

### go get libraries
```bash 
go get github.com/namsral/flag
go get github.com/anthonynsimon/bild/effect
go get github.com/anthonynsimon/bild/imgio
go get github.com/anthonynsimon/bild/transform
```
### Build it
```bash
go build
```

## Usage

run it with --help to get the Options available

It takes an Image as input and creates a STL File as Output

You can set a Resize Factor for Z with -ZDivider

Also ResizeX and ResizeY is available. If only one is set we keep Ratio.

You can Invert black and white with -Invert


### Example

```bash 
pic2stl -ZDivider 0.1 -Input image.png -Output output.stl
```

Creates from

![input image](https://github.com/ToasterKTN/pic2stl/blob/master/image.png?raw=true "Input Image")

An STL like this  (Screenshot)

![output image](https://github.com/ToasterKTN/pic2stl/blob/master/screenshot.png?raw=true "Screenshot Image")

