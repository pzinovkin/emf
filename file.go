package emf

import (
	"bytes"
	"image"
	"image/draw"

	"github.com/llgcode/draw2d/draw2dimg"
)

type EmfFile struct {
	Header  *HeaderRecord
	Records []Recorder
	EOF     *EOFRecord
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
		case *EOFRecord:
			file.EOF = rec
		default:
			file.Records = append(file.Records, rec)
		}
	}

	return file, nil
}

type context struct {
	draw2dimg.GraphicContext
	img     draw.Image
	objects map[uint32]interface{}

	w, h int

	wo, vo *PointL
	we, ve *SizeL
	mm     uint32
}

func (f *EmfFile) initContext(w, h int) *context {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	gc := draw2dimg.NewGraphicContext(img)

	return &context{
		GraphicContext: *gc,
		img:            img,
		w:              w,
		h:              h,
		mm:             MM_TEXT,
		objects:        make(map[uint32]interface{}),
	}
}

func (ctx context) applyTransformation() {
	if ctx.we == nil || ctx.ve == nil {
		return
	}

	switch ctx.mm {
	case MM_ISOTROPIC, MM_ANISOTROPIC:
		sx := float64(ctx.ve.Cx) / float64(ctx.we.Cx)
		sy := float64(ctx.ve.Cy) / float64(ctx.we.Cy)
		ctx.Scale(sx, sy)
	default:
		sx := float64(ctx.w) / float64(ctx.we.Cx)
		sy := float64(ctx.h) / float64(ctx.we.Cy)
		ctx.Scale(sx, sy)

	}
}

func (f *EmfFile) Draw() image.Image {

	bounds := f.Header.Bounds

	// inclusive-inclusive bounds
	width := int(bounds.Width()) + 1
	height := int(bounds.Height()) + 1

	ctx := f.initContext(width, height)

	if bounds.Left != 0 || bounds.Top != 0 {
		ctx.Translate(-float64(bounds.Left), -float64(bounds.Top))
	}

	for _, rec := range f.Records {
		rec.Draw(ctx)
	}

	return ctx.img
}
