package emf

import (
	"bytes"
	"image"
	"image/draw"

	"code.google.com/p/draw2d/draw2d"
)

type EmfFile struct {
	Header  *HeaderRecord
	Records []Recorder
	Eof     *EofRecord
}

func ReadFile(data []byte) (*EmfFile, error) {
	reader := bytes.NewReader(data)
	file := &EmfFile{}

	for reader.Len() > 0 {
		rec, err := readRecord(reader)

		if err != nil {
			return nil, err
		}

		switch rec := rec.(type) {
		case *HeaderRecord:
			file.Header = rec
		case *EofRecord:
			file.Eof = rec
		default:
			file.Records = append(file.Records, rec)
		}
	}

	return file, nil
}

type context struct {
	draw2d.GraphicContext
	img     draw.Image
	objects map[uint32]interface{}
}

func (f *EmfFile) initContext(w, h int) *context {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	gc := draw2d.NewGraphicContext(img)

	return &context{gc, img, make(map[uint32]interface{})}
}

func (f *EmfFile) Draw() image.Image {

	bounds := f.Header.Bounds

	width := int(bounds.Right - bounds.Left)
	height := int(bounds.Bottom - bounds.Top)

	ctx := f.initContext(width, height)

	// only vertical flip for now
	if bounds.Top < 0 && bounds.Top > bounds.Bottom {
		ctx.Translate(0, float64(height))
		ctx.Scale(1, -1)
	}

	for _, rec := range f.Records {
		rec.Draw(ctx)
	}

	return ctx.img
}
