package emf

import "image/color"

type Colorer interface {
	GetColor() color.NRGBA
}

type Header struct {
	Bounds, Frame RectL
}

type LogPaletteEntry struct {
	_, Blue, Green, Red uint8
}

type LogBrushEx struct {
	BrushStyle uint32
	Color      ColorRef
	BrushHatch uint32
}

func (l LogBrushEx) GetColor() color.NRGBA {
	return l.Color.GetColor()
}

// MS-WMF types
type ColorRef struct {
	Red, Green, Blue, _ uint8
}

func (c ColorRef) GetColor() color.NRGBA {
	return color.NRGBA{c.Red, c.Green, c.Blue, 255}
}

type SizeL struct {
	// MS-WMF says it's 32-bit unsigned integer
	// but there are files with negative values here
	Cx, Cy int32
}

type PointS struct {
	X, Y int16
}

type PointL struct {
	X, Y int32
}

type RectL struct {
	Left, Top, Right, Bottom int32
}
