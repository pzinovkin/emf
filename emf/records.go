package emf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type Recorder interface {
	Draw(*context)
}

type Record struct {
	Type, Size uint32
}

func (r *Record) Draw(ctx *context) {}

func readRecord(reader *bytes.Reader) (Recorder, error) {
	var rec Record

	if err := binary.Read(reader, binary.LittleEndian, &rec); err != nil {
		return nil, err
	}

	fn, ok := records[rec.Type]
	if !ok {
		return nil, fmt.Errorf("emf: unknown record %#v", rec.Type)
	}

	if fn != nil {
		return fn(reader, rec.Size)
	}

	// default implementation skips record data
	_, err := reader.Seek(int64(rec.Size-8), os.SEEK_CUR)
	return &rec, err
}

type HeaderRecord struct {
	Record
	Bounds, Frame RectL
}

func readHeaderRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &HeaderRecord{}
	r.Record = Record{Type: EMR_HEADER, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Frame); err != nil {
		return nil, err
	}

	reader.Seek(int64(size), os.SEEK_SET)
	return r, nil
}

type SetwindowextexRecord struct {
	Record
	Extent SizeL
}

func readSetwindowextexRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetwindowextexRecord{}
	r.Record = Record{Type: EMR_SETWINDOWEXTEX, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Extent); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SetwindowextexRecord) Draw(ctx *context) {
	ctx.Scale(
		float64(ctx.img.Bounds().Dx())/float64(r.Extent.Cx),
		float64(ctx.img.Bounds().Dy())/float64(r.Extent.Cy),
	)
}

type EofRecord struct {
	Record
	nPalEntries, offPalEntries, SizeLast uint32
}

func readEofRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &EofRecord{}
	r.Record = Record{Type: EMR_EOF, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.nPalEntries); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.offPalEntries); err != nil {
		return nil, err
	}

	if r.nPalEntries > 0 {
		fmt.Fprintln(os.Stderr, "emf: nPalEntries found - ", r.nPalEntries)
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.SizeLast); err != nil {
		return nil, err
	}

	return r, nil
}

type MovetoexRecord struct {
	Record
	Offset PointL
}

func readMovetoexRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &MovetoexRecord{}
	r.Record = Record{Type: EMR_MOVETOEX, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Offset); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *MovetoexRecord) Draw(ctx *context) {
	ctx.MoveTo(float64(r.Offset.X), float64(r.Offset.Y))
}

type SelectobjectRecord struct {
	Record
	ihObject uint32
}

func readSelectobjectRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SelectobjectRecord{}
	r.Record = Record{Type: EMR_SELECTOBJECT, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.ihObject); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SelectobjectRecord) Draw(ctx *context) {

	// predefined colors
	colors := map[uint32]Colorer{
		WHITE_BRUSH: ColorRef{Red: 255, Green: 255, Blue: 255},
		BLACK_BRUSH: ColorRef{Red: 0, Green: 0, Blue: 0},
		WHITE_PEN:   ColorRef{Red: 255, Green: 255, Blue: 255},
		BLACK_PEN:   ColorRef{Red: 0, Green: 0, Blue: 0},
	}

	color, ok := colors[r.ihObject]

	if !ok {
		color, ok = ctx.objects[r.ihObject]
		if !ok {
			fmt.Fprintf(os.Stderr, "emf: object 0x%x not found\n", r.ihObject)
			return
		}
	}

	ctx.SetFillColor(color.GetColor())
}

type CreatebrushindirectRecord struct {
	Record
	ihBrush  uint32
	LogBrush LogBrushEx
}

func readCreatebrushindirectRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &CreatebrushindirectRecord{}
	r.Record = Record{Type: EMR_CREATEBRUSHINDIRECT, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.ihBrush); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.LogBrush); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *CreatebrushindirectRecord) Draw(ctx *context) {
	ctx.objects[r.ihBrush] = r.LogBrush
}

type LinetoRecord struct {
	Record
	Point PointL
}

func readLinetoRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &LinetoRecord{}
	r.Record = Record{Type: EMR_LINETO, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Point); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *LinetoRecord) Draw(ctx *context) {
	ctx.LineTo(float64(r.Point.X), float64(r.Point.Y))
}

type BeginpathRecord struct {
	Record
}

func readBeginpathRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	return &BeginpathRecord{Record{Type: EMR_BEGINPATH, Size: size}}, nil
}

func (r *BeginpathRecord) Draw(ctx *context) {
	ctx.BeginPath()
}

type EndpathRecord struct {
	Record
}

func readEndpathRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	return &EndpathRecord{Record{Type: EMR_ENDPATH, Size: size}}, nil
}

func (r *EndpathRecord) Draw(ctx *context) {
	ctx.Close()
}

type FillpathRecord struct {
	Record
	Bounds RectL
}

func readFillpathRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &FillpathRecord{}
	r.Record = Record{Type: EMR_FILLPATH, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *FillpathRecord) Draw(ctx *context) {
	ctx.Fill()
}

type Polybezierto16Record struct {
	Record
	Bounds  RectL
	Count   uint32
	aPoints []PointS
}

func readPolybezierto16Record(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &Polybezierto16Record{}
	r.Record = Record{Type: EMR_POLYBEZIERTO16, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Count); err != nil {
		return nil, err
	}

	for i := 0; i < int(r.Count); i++ {
		var p PointS

		if err := binary.Read(reader, binary.LittleEndian, &p); err != nil {
			return nil, err
		}

		r.aPoints = append(r.aPoints, p)
	}

	return r, nil
}

func (r *Polybezierto16Record) Draw(ctx *context) {
	for i := 0; i < int(r.Count); i = i + 3 {
		ctx.CubicCurveTo(
			float64(r.aPoints[i].X), float64(r.aPoints[i].Y),
			float64(r.aPoints[i+1].X), float64(r.aPoints[i+1].Y),
			float64(r.aPoints[i+2].X), float64(r.aPoints[i+2].Y),
		)
	}
}

type Polylineto16Record struct {
	Record
	Bounds  RectL
	Count   uint32
	aPoints []PointS
}

func readPolylineto16Record(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &Polylineto16Record{}
	r.Record = Record{Type: EMR_POLYLINETO16, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Count); err != nil {
		return nil, err
	}

	for i := 0; i < int(r.Count); i++ {
		var p PointS

		if err := binary.Read(reader, binary.LittleEndian, &p); err != nil {
			return nil, err
		}

		r.aPoints = append(r.aPoints, p)
	}

	return r, nil
}

func (r *Polylineto16Record) Draw(ctx *context) {
	for i := 0; i < int(r.Count); i++ {
		ctx.LineTo(float64(r.aPoints[i].X), float64(r.aPoints[i].Y))
	}
}

// map of readers for records
var records = map[uint32]func(*bytes.Reader, uint32) (Recorder, error){
	EMR_HEADER:                  readHeaderRecord,
	EMR_POLYBEZIER:              nil,
	EMR_POLYGON:                 nil,
	EMR_POLYLINE:                nil,
	EMR_POLYBEZIERTO:            nil,
	EMR_POLYLINETO:              nil,
	EMR_POLYPOLYLINE:            nil,
	EMR_POLYPOLYGON:             nil,
	EMR_SETWINDOWEXTEX:          readSetwindowextexRecord,
	EMR_SETWINDOWORGEX:          nil,
	EMR_SETVIEWPORTEXTEX:        nil,
	EMR_SETVIEWPORTORGEX:        nil,
	EMR_SETBRUSHORGEX:           nil,
	EMR_EOF:                     readEofRecord,
	EMR_SETPIXELV:               nil,
	EMR_SETMAPPERFLAGS:          nil,
	EMR_SETMAPMODE:              nil,
	EMR_SETBKMODE:               nil,
	EMR_SETPOLYFILLMODE:         nil,
	EMR_SETROP2:                 nil,
	EMR_SETSTRETCHBLTMODE:       nil,
	EMR_SETTEXTALIGN:            nil,
	EMR_SETCOLORADJUSTMENT:      nil,
	EMR_SETTEXTCOLOR:            nil,
	EMR_SETBKCOLOR:              nil,
	EMR_OFFSETCLIPRGN:           nil,
	EMR_MOVETOEX:                readMovetoexRecord,
	EMR_SETMETARGN:              nil,
	EMR_EXCLUDECLIPRECT:         nil,
	EMR_INTERSECTCLIPRECT:       nil,
	EMR_SCALEVIEWPORTEXTEX:      nil,
	EMR_SCALEWINDOWEXTEX:        nil,
	EMR_SAVEDC:                  nil,
	EMR_RESTOREDC:               nil,
	EMR_SETWORLDTRANSFORM:       nil,
	EMR_MODIFYWORLDTRANSFORM:    nil,
	EMR_SELECTOBJECT:            readSelectobjectRecord,
	EMR_CREATEPEN:               nil,
	EMR_CREATEBRUSHINDIRECT:     readCreatebrushindirectRecord,
	EMR_DELETEOBJECT:            nil,
	EMR_ANGLEARC:                nil,
	EMR_ELLIPSE:                 nil,
	EMR_RECTANGLE:               nil,
	EMR_ROUNDRECT:               nil,
	EMR_ARC:                     nil,
	EMR_CHORD:                   nil,
	EMR_PIE:                     nil,
	EMR_SELECTPALETTE:           nil,
	EMR_CREATEPALETTE:           nil,
	EMR_SETPALETTEENTRIES:       nil,
	EMR_RESIZEPALETTE:           nil,
	EMR_REALIZEPALETTE:          nil,
	EMR_EXTFLOODFILL:            nil,
	EMR_LINETO:                  readLinetoRecord,
	EMR_ARCTO:                   nil,
	EMR_POLYDRAW:                nil,
	EMR_SETARCDIRECTION:         nil,
	EMR_SETMITERLIMIT:           nil,
	EMR_BEGINPATH:               readBeginpathRecord,
	EMR_ENDPATH:                 readEndpathRecord,
	EMR_CLOSEFIGURE:             nil,
	EMR_FILLPATH:                readFillpathRecord,
	EMR_STROKEANDFILLPATH:       nil,
	EMR_STROKEPATH:              nil,
	EMR_FLATTENPATH:             nil,
	EMR_WIDENPATH:               nil,
	EMR_SELECTCLIPPATH:          nil,
	EMR_ABORTPATH:               nil,
	EMR_COMMENT:                 nil,
	EMR_FILLRGN:                 nil,
	EMR_FRAMERGN:                nil,
	EMR_INVERTRGN:               nil,
	EMR_PAINTRGN:                nil,
	EMR_EXTSELECTCLIPRGN:        nil,
	EMR_BITBLT:                  nil,
	EMR_STRETCHBLT:              nil,
	EMR_MASKBLT:                 nil,
	EMR_PLGBLT:                  nil,
	EMR_SETDIBITSTODEVICE:       nil,
	EMR_STRETCHDIBITS:           nil,
	EMR_EXTCREATEFONTINDIRECTW:  nil,
	EMR_EXTTEXTOUTA:             nil,
	EMR_EXTTEXTOUTW:             nil,
	EMR_POLYBEZIER16:            nil,
	EMR_POLYGON16:               nil,
	EMR_POLYLINE16:              nil,
	EMR_POLYBEZIERTO16:          readPolybezierto16Record,
	EMR_POLYLINETO16:            readPolylineto16Record,
	EMR_POLYPOLYLINE16:          nil,
	EMR_POLYPOLYGON16:           nil,
	EMR_POLYDRAW16:              nil,
	EMR_CREATEMONOBRUSH:         nil,
	EMR_CREATEDIBPATTERNBRUSHPT: nil,
	EMR_EXTCREATEPEN:            nil,
	EMR_POLYTEXTOUTA:            nil,
	EMR_POLYTEXTOUTW:            nil,
	EMR_SETICMMODE:              nil,
	EMR_CREATECOLORSPACE:        nil,
	EMR_SETCOLORSPACE:           nil,
	EMR_DELETECOLORSPACE:        nil,
	EMR_GLSRECORD:               nil,
	EMR_GLSBOUNDEDRECORD:        nil,
	EMR_PIXELFORMAT:             nil,
	EMR_DRAWESCAPE:              nil,
	EMR_EXTESCAPE:               nil,
	EMR_SMALLTEXTOUT:            nil,
	EMR_FORCEUFIMAPPING:         nil,
	EMR_NAMEDESCAPE:             nil,
	EMR_COLORCORRECTPALETTE:     nil,
	EMR_SETICMPROFILEA:          nil,
	EMR_SETICMPROFILEW:          nil,
	EMR_ALPHABLEND:              nil,
	EMR_SETLAYOUT:               nil,
	EMR_TRANSPARENTBLT:          nil,
	EMR_GRADIENTFILL:            nil,
	EMR_SETLINKEDUFIS:           nil,
	EMR_SETTEXTJUSTIFICATION:    nil,
	EMR_COLORMATCHTOTARGETW:     nil,
	EMR_CREATECOLORSPACEW:       nil,
}
