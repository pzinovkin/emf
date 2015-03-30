package emf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"os"

	"code.google.com/p/draw2d/draw2d"
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
		return nil, fmt.Errorf("unknown record %#v", rec.Type)
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
	Bounds, Frame           RectL
	RecordSignature         uint32
	Version, Bytes, Records uint32
	Handles                 uint16
	nDescription            uint32
	offDescription          uint32
	nPalEntries             uint32
	Device, Millimeters     SizeL
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

	if err := binary.Read(reader, binary.LittleEndian, &r.RecordSignature); err != nil {
		return nil, err
	}

	if r.RecordSignature != ENHMETA_SIGNATURE {
		return nil, fmt.Errorf("unknown signature %#v", r.RecordSignature)
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Version); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bytes); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Records); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Handles); err != nil {
		return nil, err
	}

	// Reserved
	reader.Seek(int64(2), os.SEEK_CUR)

	if err := binary.Read(reader, binary.LittleEndian, &r.nDescription); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.offDescription); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.nPalEntries); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Device); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Millimeters); err != nil {
		return nil, err
	}
	// skip the rest of structure
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
	ctx.we = &r.Extent
	ctx.applyTransformation()
}

type SetwindoworgexRecord struct {
	Record
	Origin PointL
}

func readSetwindoworgexRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetwindoworgexRecord{}
	r.Record = Record{Type: EMR_SETWINDOWORGEX, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Origin); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SetwindoworgexRecord) Draw(ctx *context) {
	ctx.wo = &r.Origin
	ctx.applyTransformation()
}

type SetviewportextexRecord struct {
	Record
	Extent SizeL
}

func readSetviewportextexRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetviewportextexRecord{}
	r.Record = Record{Type: EMR_SETVIEWPORTEXTEX, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Extent); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SetviewportextexRecord) Draw(ctx *context) {
	ctx.ve = &r.Extent
	ctx.applyTransformation()
}

type SetviewportorgexRecord struct {
	Record
	Origin PointL
}

func readSetviewportorgexRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetviewportorgexRecord{}
	r.Record = Record{Type: EMR_SETVIEWPORTORGEX, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Origin); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SetviewportorgexRecord) Draw(ctx *context) {
	ctx.vo = &r.Origin
	ctx.applyTransformation()
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

type SetmapmodeRecord struct {
	Record
	MapMode uint32
}

func readSetmapmodeRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetmapmodeRecord{}
	r.Record = Record{Type: EMR_SETMAPMODE, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.MapMode); err != nil {
		return nil, err
	}

	return r, nil
}

// https://www-user.tu-chemnitz.de/~heha/petzold/ch05f.htm
// http://msdn.microsoft.com/en-us/library/dd183475(v=vs.85).aspx
func (r *SetmapmodeRecord) Draw(ctx *context) {
	ctx.mm = r.MapMode
	switch r.MapMode {
	// rotate y axis
	case MM_LOMETRIC, MM_HIMETRIC, MM_LOENGLISH, MM_HIENGLISH, MM_TWIPS:
		ctx.Scale(1, -1)
		// can't use ctx.Translate here because it will be scaled
		// if scaling already applied before
		tr := ctx.GetMatrixTransform()
		tr[5] = float64(ctx.h)
		ctx.SetMatrixTransform(tr)
	}
}

type SetbkmodeRecord struct {
	Record
	BackgroundMode uint32
}

func readSetbkmodeRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetbkmodeRecord{}
	r.Record = Record{Type: EMR_SETBKMODE, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.BackgroundMode); err != nil {
		return nil, err
	}

	return r, nil
}

type SetpolyfillmodeRecord struct {
	Record
	PolygonFillMode uint32
}

func readSetpolyfillmodeRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetpolyfillmodeRecord{}
	r.Record = Record{Type: EMR_SETPOLYFILLMODE, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.PolygonFillMode); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SetpolyfillmodeRecord) Draw(ctx *context) {
	if r.PolygonFillMode == ALTERNATE {
		ctx.SetFillRule(draw2d.FillRuleEvenOdd)
	} else if r.PolygonFillMode == WINDING {
		ctx.SetFillRule(draw2d.FillRuleWinding)
	}
}

type SettextalignRecord struct {
	Record
	TextAlignmentMode uint32
}

func readSettextalignRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SettextalignRecord{}
	r.Record = Record{Type: EMR_SETTEXTALIGN, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.TextAlignmentMode); err != nil {
		return nil, err
	}

	return r, nil
}

type SetstretchbltmodeRecord struct {
	Record
	StretchMode uint32
}

func readSetstretchbltmodeRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetstretchbltmodeRecord{}
	r.Record = Record{Type: EMR_SETSTRETCHBLTMODE, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.StretchMode); err != nil {
		return nil, err
	}

	return r, nil
}

type SettextcolorRecord struct {
	Record
	Color ColorRef
}

func readSettextcolorRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SettextcolorRecord{}
	r.Record = Record{Type: EMR_SETTEXTCOLOR, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Color); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SettextcolorRecord) Draw(ctx *context) {
	ctx.SetFillColor(r.Color.GetColor())
}

type SetbkcolorRecord struct {
	Record
	Color ColorRef
}

func readSetbkcolorRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetbkcolorRecord{}
	r.Record = Record{Type: EMR_SETBKCOLOR, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Color); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SetbkcolorRecord) Draw(ctx *context) {
	ctx.SetFillColor(r.Color.GetColor())
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

type IntersectcliprectRecord struct {
	Record
	Clip RectL
}

func readIntersectcliprectRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &IntersectcliprectRecord{}
	r.Record = Record{Type: EMR_INTERSECTCLIPRECT, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Clip); err != nil {
		return nil, err
	}

	return r, nil
}

type SavedcRecord struct {
	Record
}

func readSavedcRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	return &SavedcRecord{Record: Record{Type: EMR_SAVEDC, Size: size}}, nil
}

func (r *SavedcRecord) Draw(ctx *context) {
	ctx.Save()
}

type RestoredcRecord struct {
	Record
	SavedDC int32
}

func readRestoredcRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &RestoredcRecord{}
	r.Record = Record{Type: EMR_RESTOREDC, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.SavedDC); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *RestoredcRecord) Draw(ctx *context) {
	ctx.Restore()
}

type SetworldtransformRecord struct {
	Record
	XForm XForm
}

func readSetworldtransformRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SetworldtransformRecord{}
	r.Record = Record{Type: EMR_SETWORLDTRANSFORM, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.XForm); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SetworldtransformRecord) Draw(ctx *context) {
	tr := ctx.GetMatrixTransform()
	tr[0] = float64(r.XForm.M11)
	tr[1] = float64(r.XForm.M12)
	tr[2] = float64(r.XForm.M21)
	tr[3] = float64(r.XForm.M22)
	tr[4] = float64(r.XForm.Dx)
	tr[5] = float64(r.XForm.Dy)
	ctx.SetMatrixTransform(tr)
}

type ModifyworldtransformRecord struct {
	Record
	XForm                    XForm
	ModifyWorldTransformMode uint32
}

func readModifyworldtransformRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &ModifyworldtransformRecord{}
	r.Record = Record{Type: EMR_MODIFYWORLDTRANSFORM, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.XForm); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.ModifyWorldTransformMode); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *ModifyworldtransformRecord) Draw(ctx *context) {
	ctx.Scale(float64(r.XForm.M11), float64(r.XForm.M22))
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

var StockObjects = map[uint32]interface{}{
	WHITE_BRUSH:         LogBrushEx{Color: ColorRef{Red: 255, Green: 255, Blue: 255}},
	LTGRAY_BRUSH:        LogBrushEx{Color: ColorRef{Red: 192, Green: 192, Blue: 192}},
	GRAY_BRUSH:          LogBrushEx{Color: ColorRef{Red: 128, Green: 128, Blue: 128}},
	DKGRAY_BRUSH:        LogBrushEx{Color: ColorRef{Red: 64, Green: 64, Blue: 64}},
	BLACK_BRUSH:         LogBrushEx{Color: ColorRef{Red: 0, Green: 0, Blue: 0}},
	NULL_BRUSH:          true,
	WHITE_PEN:           LogPen{ColorRef: ColorRef{Red: 255, Green: 255, Blue: 255}, Width: PointL{1, 0}},
	BLACK_PEN:           LogPen{ColorRef: ColorRef{Red: 0, Green: 0, Blue: 0}, Width: PointL{1, 0}},
	NULL_PEN:            true,
	SYSTEM_FONT:         LogFont{Height: 11},
	DEVICE_DEFAULT_FONT: LogFont{Height: 11},
}

func (r *SelectobjectRecord) Draw(ctx *context) {

	object, ok := StockObjects[r.ihObject]
	if !ok {
		object, ok = ctx.objects[r.ihObject]
		if !ok {
			fmt.Fprintf(os.Stderr, "emf: object 0x%x not found\n", r.ihObject)
			return
		}
	}

	switch o := object.(type) {
	case bool:
		if r.ihObject == NULL_PEN {
			ctx.SetStrokeColor(image.Transparent)
		} else if r.ihObject == NULL_BRUSH {
			ctx.SetFillColor(image.Transparent)
		}
	case LogPen:
		ctx.SetLineWidth(float64(o.Width.X))
		ctx.SetStrokeColor(o.ColorRef.GetColor())
	case LogPenEx:
		ctx.SetLineWidth(float64(o.Width))
		ctx.SetStrokeColor(o.ColorRef.GetColor())
	case LogBrushEx:
		ctx.SetFillColor(o.Color.GetColor())
	}
}

type CreatepenRecord struct {
	Record
	ihPen  uint32
	LogPen LogPen
}

func readCreatepenRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &CreatepenRecord{}
	r.Record = Record{Type: EMR_CREATEPEN, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.ihPen); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.LogPen); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *CreatepenRecord) Draw(ctx *context) {
	ctx.objects[r.ihPen] = r.LogPen
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

type DeleteobjectRecord struct {
	Record
	ihObject uint32
}

func readDeleteobjectRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &DeleteobjectRecord{}
	r.Record = Record{Type: EMR_DELETEOBJECT, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.ihObject); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *DeleteobjectRecord) Draw(ctx *context) {
	delete(ctx.objects, r.ihObject)
}

type RectangleRecord struct {
	Record
	Box RectL
}

func readRectangleRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &RectangleRecord{}
	r.Record = Record{Type: EMR_RECTANGLE, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Box); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RectangleRecord) Draw(ctx *context) {
	x1, y1, x2, y2 := float64(r.Box.Left), float64(r.Box.Top), float64(r.Box.Right), float64(r.Box.Bottom)
	ctx.MoveTo(x1, y1)
	ctx.LineTo(x2, y1)
	ctx.LineTo(x2, y2)
	ctx.LineTo(x1, y2)
	ctx.LineTo(x1, y1)
	ctx.FillStroke()
}

type ArcRecord struct {
	Record
	Box   RectL
	Start PointL
	End   PointL
}

func readArcRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &ArcRecord{}
	r.Record = Record{Type: EMR_ARC, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Box); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.Start); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &r.End); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *ArcRecord) Draw(ctx *context) {
	center := r.Box.Center()
	rx := (float64(r.Box.Right) - float64(r.Box.Left) - 1) / 2
	ry := (float64(r.Box.Bottom) - float64(r.Box.Top) - 1) / 2
	// angles are specified in radians
	sa := math.Atan2(float64(r.Start.Y-center.Y), float64(r.Start.X-center.X))
	ea := math.Atan2(float64(r.End.Y-center.Y), float64(r.End.X-center.X)) - sa

	ctx.ArcTo(float64(center.X), float64(center.Y), rx, ry, sa, ea)
	ctx.Stroke()
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

type ClosefigureRecord struct {
	Record
}

func readClosefigureRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	return &ClosefigureRecord{Record{Type: EMR_CLOSEFIGURE, Size: size}}, nil
}

func (r *ClosefigureRecord) Draw(ctx *context) {
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

type StrokeandfillpathRecord struct {
	Record
	Bounds RectL
}

func readStrokeandfillpathRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &StrokeandfillpathRecord{}
	r.Record = Record{Type: EMR_STROKEANDFILLPATH, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *StrokeandfillpathRecord) Draw(ctx *context) {
	ctx.Fill()
	ctx.Stroke()
}

type StrokepathRecord struct {
	Record
	Bounds RectL
}

func readStrokepathRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &StrokepathRecord{}
	r.Record = Record{Type: EMR_STROKEPATH, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *StrokepathRecord) Draw(ctx *context) {
	ctx.Stroke()
}

type SelectclippathRecord struct {
	Record
	RegionMode uint32
}

func readSelectclippathRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SelectclippathRecord{}
	r.Record = Record{Type: EMR_SELECTCLIPPATH, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.RegionMode); err != nil {
		return nil, err
	}
	return r, nil
}

type CommentRecord struct {
	Record
}

func readCommentRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &CommentRecord{}
	r.Record = Record{Type: EMR_COMMENT, Size: size}
	// skip record data
	reader.Seek(int64(r.Size-8), os.SEEK_CUR)
	return r, nil
}

type ExtcreatefontindirectwRecord struct {
	Record
	ihFonts uint32
	elw     LogFont
}

func readExtcreatefontindirectwRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &ExtcreatefontindirectwRecord{}
	r.Record = Record{Type: EMR_EXTCREATEFONTINDIRECTW, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.ihFonts); err != nil {
		return nil, err
	}

	var err error

	r.elw, err = readLogFont(reader)
	if err != nil {
		return nil, err
	}

	// skip the rest because we read only limited amount of data (LogFont) here
	reader.Seek(int64(r.Size-(12+92)), os.SEEK_CUR)

	return r, nil
}

func (r *ExtcreatefontindirectwRecord) Draw(ctx *context) {
	ctx.objects[r.ihFonts] = r.elw
}

type ExttextoutwRecord struct {
	Record
	Bounds           RectL
	iGraphicsMode    uint32
	exScale, eyScale float32
	wEmrText         EmrText
}

func readExttextoutwRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &ExttextoutwRecord{}
	r.Record = Record{Type: EMR_EXTTEXTOUTW, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.iGraphicsMode); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.exScale); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.eyScale); err != nil {
		return nil, err
	}

	var err error
	r.wEmrText, err = readEmrText(reader, reader.Len()+36)
	if err != nil {
		return nil, err
	}

	return r, nil
}

type Polybezier16Record struct {
	Record
	Bounds  RectL
	Count   uint32
	aPoints []PointS
}

func readPolybezier16Record(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &Polybezier16Record{}
	r.Record = Record{Type: EMR_POLYBEZIER16, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Count); err != nil {
		return nil, err
	}

	r.aPoints = make([]PointS, r.Count)
	if err := binary.Read(reader, binary.LittleEndian, &r.aPoints); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Polybezier16Record) Draw(ctx *context) {
	ctx.MoveTo(float64(r.aPoints[0].X), float64(r.aPoints[0].Y))
	for i := 1; i < int(r.Count); i = i + 3 {
		ctx.CubicCurveTo(
			float64(r.aPoints[i].X), float64(r.aPoints[i].Y),
			float64(r.aPoints[i+1].X), float64(r.aPoints[i+1].Y),
			float64(r.aPoints[i+2].X), float64(r.aPoints[i+2].Y),
		)
	}
}

type Polygon16Record struct {
	Record
	Bounds  RectL
	Count   uint32
	aPoints []PointS
}

func readPolygon16Record(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &Polygon16Record{}
	r.Record = Record{Type: EMR_POLYGON16, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Count); err != nil {
		return nil, err
	}

	r.aPoints = make([]PointS, r.Count)
	if err := binary.Read(reader, binary.LittleEndian, &r.aPoints); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Polygon16Record) Draw(ctx *context) {
	ctx.MoveTo(float64(r.aPoints[0].X), float64(r.aPoints[0].Y))
	for i := 1; i < int(r.Count); i++ {
		ctx.LineTo(float64(r.aPoints[i].X), float64(r.aPoints[i].Y))
	}
	ctx.Close()
	ctx.FillStroke()
}

type Polyline16Record struct {
	Record
	Bounds  RectL
	Count   uint32
	aPoints []PointS
}

func readPolyline16Record(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &Polyline16Record{}
	r.Record = Record{Type: EMR_POLYLINE16, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Count); err != nil {
		return nil, err
	}

	r.aPoints = make([]PointS, r.Count)
	if err := binary.Read(reader, binary.LittleEndian, &r.aPoints); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Polyline16Record) Draw(ctx *context) {
	ctx.MoveTo(float64(r.aPoints[0].X), float64(r.aPoints[0].Y))
	for i := 1; i < int(r.Count); i++ {
		ctx.LineTo(float64(r.aPoints[i].X), float64(r.aPoints[i].Y))
	}
	// ctx.Stroke()
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

	r.aPoints = make([]PointS, r.Count)
	if err := binary.Read(reader, binary.LittleEndian, &r.aPoints); err != nil {
		return nil, err
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

	r.aPoints = make([]PointS, r.Count)
	if err := binary.Read(reader, binary.LittleEndian, &r.aPoints); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Polylineto16Record) Draw(ctx *context) {
	for i := 0; i < int(r.Count); i++ {
		ctx.LineTo(float64(r.aPoints[i].X), float64(r.aPoints[i].Y))
	}
}

type Polypolygon16Record struct {
	Record
	Bounds            RectL
	NumberOfPolygons  uint32
	Count             uint32
	PolygonPointCount []uint32
	aPoints           []PointS
}

func readPolypolygon16Record(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &Polypolygon16Record{}
	r.Record = Record{Type: EMR_POLYPOLYGON16, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.NumberOfPolygons); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.Count); err != nil {
		return nil, err
	}

	r.PolygonPointCount = make([]uint32, r.NumberOfPolygons)
	if err := binary.Read(reader, binary.LittleEndian, &r.PolygonPointCount); err != nil {
		return nil, err
	}

	r.aPoints = make([]PointS, r.Count)
	if err := binary.Read(reader, binary.LittleEndian, &r.aPoints); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Polypolygon16Record) Draw(ctx *context) {
	idx := 0
	for p := 0; p < int(r.NumberOfPolygons); p++ {
		pCount := int(r.PolygonPointCount[p])
		ctx.MoveTo(float64(r.aPoints[idx].X), float64(r.aPoints[idx].Y))
		for i := 1; i < pCount; i++ {
			ctx.LineTo(float64(r.aPoints[idx+i].X), float64(r.aPoints[idx+i].Y))
		}
		idx += pCount
		ctx.Close()
	}
	ctx.FillStroke()
}

type ExtcreatepenRecord struct {
	Record
	ihPen           uint32
	offBmi, cbBmi   uint32
	offBits, cbBits uint32
	elp             LogPenEx
	BmiSrc          DibHeaderInfo
	BitsSrc         []byte
}

func readExtcreatepenRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &ExtcreatepenRecord{}
	r.Record = Record{Type: EMR_EXTCREATEPEN, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.ihPen); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.offBmi); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cbBmi); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.offBits); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cbBits); err != nil {
		return nil, err
	}

	var err error
	r.elp, err = readLogPenEx(reader)
	if err != nil {
		return nil, err
	}

	// offset for bitmap info less than possible minimum
	// assuming there is no bitmap
	if r.offBmi < 52 {
		return r, nil
	}

	// BitmapBuffer
	// skipping UndefinedSpace
	reader.Seek(int64(r.offBmi-52-(r.elp.NumStyleEntries*4)), os.SEEK_CUR)

	// record does not contain bitmap
	if r.cbBmi == 0 {
		return r, nil
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.BmiSrc); err != nil {
		return nil, err
	}

	r.BitsSrc = make([]byte, r.cbBits)
	if _, err := reader.Read(r.BitsSrc); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *ExtcreatepenRecord) Draw(ctx *context) {
	ctx.objects[r.ihPen] = r.elp
}

type SeticmmodeRecord struct {
	Record
	ICMMode uint32
}

func readSeticmmodeRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &SeticmmodeRecord{}
	r.Record = Record{Type: EMR_SETICMMODE, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.ICMMode); err != nil {
		return nil, err
	}

	return r, nil
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
	EMR_SETWINDOWORGEX:          readSetwindoworgexRecord,
	EMR_SETVIEWPORTEXTEX:        readSetviewportextexRecord,
	EMR_SETVIEWPORTORGEX:        readSetviewportorgexRecord,
	EMR_SETBRUSHORGEX:           nil,
	EMR_EOF:                     readEofRecord,
	EMR_SETPIXELV:               nil,
	EMR_SETMAPPERFLAGS:          nil,
	EMR_SETMAPMODE:              readSetmapmodeRecord,
	EMR_SETBKMODE:               readSetbkmodeRecord,
	EMR_SETPOLYFILLMODE:         readSetpolyfillmodeRecord,
	EMR_SETROP2:                 nil,
	EMR_SETSTRETCHBLTMODE:       readSetstretchbltmodeRecord,
	EMR_SETTEXTALIGN:            readSettextalignRecord,
	EMR_SETCOLORADJUSTMENT:      nil,
	EMR_SETTEXTCOLOR:            readSettextcolorRecord,
	EMR_SETBKCOLOR:              readSetbkcolorRecord,
	EMR_OFFSETCLIPRGN:           nil,
	EMR_MOVETOEX:                readMovetoexRecord,
	EMR_SETMETARGN:              nil,
	EMR_EXCLUDECLIPRECT:         nil,
	EMR_INTERSECTCLIPRECT:       readIntersectcliprectRecord,
	EMR_SCALEVIEWPORTEXTEX:      nil,
	EMR_SCALEWINDOWEXTEX:        nil,
	EMR_SAVEDC:                  readSavedcRecord,
	EMR_RESTOREDC:               readRestoredcRecord,
	EMR_SETWORLDTRANSFORM:       readSetworldtransformRecord,
	EMR_MODIFYWORLDTRANSFORM:    readModifyworldtransformRecord,
	EMR_SELECTOBJECT:            readSelectobjectRecord,
	EMR_CREATEPEN:               readCreatepenRecord,
	EMR_CREATEBRUSHINDIRECT:     readCreatebrushindirectRecord,
	EMR_DELETEOBJECT:            readDeleteobjectRecord,
	EMR_ANGLEARC:                nil,
	EMR_ELLIPSE:                 nil,
	EMR_RECTANGLE:               readRectangleRecord,
	EMR_ROUNDRECT:               nil,
	EMR_ARC:                     readArcRecord,
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
	EMR_CLOSEFIGURE:             readClosefigureRecord,
	EMR_FILLPATH:                readFillpathRecord,
	EMR_STROKEANDFILLPATH:       readStrokeandfillpathRecord,
	EMR_STROKEPATH:              readStrokepathRecord,
	EMR_FLATTENPATH:             nil,
	EMR_WIDENPATH:               nil,
	EMR_SELECTCLIPPATH:          readSelectclippathRecord,
	EMR_ABORTPATH:               nil,
	EMR_COMMENT:                 readCommentRecord,
	EMR_FILLRGN:                 nil,
	EMR_FRAMERGN:                nil,
	EMR_INVERTRGN:               nil,
	EMR_PAINTRGN:                nil,
	EMR_EXTSELECTCLIPRGN:        nil,
	EMR_BITBLT:                  readBitbltRecord,
	EMR_STRETCHBLT:              readStretchbltRecord,
	EMR_MASKBLT:                 nil,
	EMR_PLGBLT:                  nil,
	EMR_SETDIBITSTODEVICE:       nil,
	EMR_STRETCHDIBITS:           readStretchdibitsRecord,
	EMR_EXTCREATEFONTINDIRECTW:  readExtcreatefontindirectwRecord,
	EMR_EXTTEXTOUTA:             nil,
	EMR_EXTTEXTOUTW:             readExttextoutwRecord,
	EMR_POLYBEZIER16:            readPolybezier16Record,
	EMR_POLYGON16:               readPolygon16Record,
	EMR_POLYLINE16:              readPolyline16Record,
	EMR_POLYBEZIERTO16:          readPolybezierto16Record,
	EMR_POLYLINETO16:            readPolylineto16Record,
	EMR_POLYPOLYLINE16:          nil,
	EMR_POLYPOLYGON16:           readPolypolygon16Record,
	EMR_POLYDRAW16:              nil,
	EMR_CREATEMONOBRUSH:         nil,
	EMR_CREATEDIBPATTERNBRUSHPT: nil,
	EMR_EXTCREATEPEN:            readExtcreatepenRecord,
	EMR_POLYTEXTOUTA:            nil,
	EMR_POLYTEXTOUTW:            nil,
	EMR_SETICMMODE:              readSeticmmodeRecord,
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
