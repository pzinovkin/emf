package emf

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"os"
	"strings"
	"unicode/utf16"
)

type LogPaletteEntry struct {
	_, Blue, Green, Red uint8
}

type LogPen struct {
	PenStyle uint32
	Width    PointL
	ColorRef ColorRef
}

type LogPenEx struct {
	PenStyle        uint32
	Width           uint32
	BrushStyle      uint32
	ColorRef        ColorRef
	BrushHatch      uint32
	NumStyleEntries uint32
	StyleEntry      []uint32
}

func readLogPenEx(reader *bytes.Reader) (LogPenEx, error) {
	r := LogPenEx{}
	if err := binary.Read(reader, binary.LittleEndian, &r.PenStyle); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Width); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.BrushStyle); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.ColorRef); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.BrushHatch); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.NumStyleEntries); err != nil {
		return r, err
	}

	if r.PenStyle == PS_USERSTYLE && r.NumStyleEntries > 0 {
		r.StyleEntry = make([]uint32, r.NumStyleEntries)
		if err := binary.Read(reader, binary.LittleEndian, &r.StyleEntry); err != nil {
			return r, err
		}
	}

	return r, nil
}

type LogBrushEx struct {
	BrushStyle uint32
	Color      ColorRef
	BrushHatch uint32
}

type XForm struct {
	M11, M12, M21, M22, Dx, Dy float32
}

type EmrText struct {
	Reference    PointL
	Chars        uint32
	offString    uint32
	Options      uint32
	Rectangle    RectL
	offDx        uint32
	OutputString string
	OutputDx     []uint32
}

func readEmrText(reader *bytes.Reader, offset int) (EmrText, error) {
	r := EmrText{}
	if err := binary.Read(reader, binary.LittleEndian, &r.Reference); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Chars); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.offString); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Options); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Rectangle); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.offDx); err != nil {
		return r, err
	}

	// UndefinedSpace1
	reader.Seek(int64(int(r.offString)-(offset-reader.Len())), os.SEEK_CUR)
	b := make([]uint16, r.Chars)
	if err := binary.Read(reader, binary.LittleEndian, &b); err != nil {
		return r, err
	}
	r.OutputString = string(utf16.Decode(b))

	// UndefinedSpace2
	reader.Seek(int64(int(r.offDx)-(offset-reader.Len())), os.SEEK_CUR)
	r.OutputDx = make([]uint32, r.Chars)
	if err := binary.Read(reader, binary.LittleEndian, &r.OutputDx); err != nil {
		return r, err
	}

	return r, nil
}

type LogFont struct {
	Height, Width                        int32
	Escapement, Orientation, Weight      int32
	Italic, Underline, StrikeOut         uint8
	CharSet, OutPrecision, ClipPrecision uint8
	Quality                              uint8
	PitchAndFamily                       int8
	Facename                             string
}

func readLogFont(reader *bytes.Reader) (LogFont, error) {
	r := LogFont{}
	if err := binary.Read(reader, binary.LittleEndian, &r.Height); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Width); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Escapement); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Orientation); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Weight); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Italic); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Underline); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.StrikeOut); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.CharSet); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.OutPrecision); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.ClipPrecision); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Quality); err != nil {
		return r, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.PitchAndFamily); err != nil {
		return r, err
	}

	b := make([]uint16, 32)
	if err := binary.Read(reader, binary.LittleEndian, &b); err != nil {
		return r, err
	}
	r.Facename = strings.TrimRight(string(utf16.Decode(b)), "\x00")

	return r, nil
}

// MS-WMF types
type ColorRef struct {
	Red, Green, Blue, _ uint8
}

func (c ColorRef) GetColor() color.RGBA {
	return color.RGBA{c.Red, c.Green, c.Blue, 0xff}
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

func (r RectL) Width() int32  { return r.Right - r.Left }
func (r RectL) Height() int32 { return r.Bottom - r.Top }

func (r RectL) Center() PointL {
	return PointL{
		X: r.Left + r.Width()/2,
		Y: r.Top + r.Height()/2,
	}
}

type BitmapInfoHeader struct {
	HeaderSize                   uint32
	Width, Height                int32
	Planes, BitCount             uint16
	Compression, ImageSize       uint32
	XPelsPerMeter, YPelsPerMeter int32
	ColorUsed, ColorImportant    uint32
}

type DibHeaderInfo struct{}
