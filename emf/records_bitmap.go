package emf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"os"

	"github.com/disintegration/imaging"
)

type bitmapRecord struct {
	Record
	Bounds                       RectL
	xDest, yDest, cxDest, cyDest int32
	BitBltRasterOperation        uint32
	xSrc, ySrc                   int32
	XformSrc                     XForm
	BkColorSrc                   ColorRef
	UsageSrc                     uint32
	offBmiSrc, cbBmiSrc          uint32
	offBitsSrc, cbBitsSrc        uint32
	// only for EMR_STRETCHBLT
	cxSrc, cySrc int32

	BmiSrc  BitmapInfoHeader
	BitsSrc []byte
}

// unified reader function for EMR_BITBLT and EMR_STRETCHBLT
func (r *bitmapRecord) read(reader *bytes.Reader) (Recorder, error) {
	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.xDest); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.yDest); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cxDest); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cyDest); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.BitBltRasterOperation); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.xSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.ySrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.XformSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.BkColorSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.UsageSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.offBmiSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cbBmiSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.offBitsSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cbBitsSrc); err != nil {
		return nil, err
	}

	if r.Type == EMR_STRETCHBLT {
		if err := binary.Read(reader, binary.LittleEndian, &r.cxSrc); err != nil {
			return nil, err
		}

		if err := binary.Read(reader, binary.LittleEndian, &r.cySrc); err != nil {
			return nil, err
		}
	}

	// no bitmap data
	if r.offBmiSrc == 0 {
		return r, nil
	}

	// defined record size to skip UndefinedSpace
	var rsize uint32
	if r.Type == EMR_STRETCHBLT {
		rsize = 108
	} else if r.Type == EMR_BITBLT {
		rsize = 100
	}

	// BitmapBuffer
	// skipping UndefinedSpace1
	reader.Seek(int64(r.offBmiSrc-rsize), os.SEEK_CUR)
	if err := binary.Read(reader, binary.LittleEndian, &r.BmiSrc); err != nil {
		return nil, err
	}

	// skipping UndefinedSpace2
	reader.Seek(int64(r.offBitsSrc-rsize-r.BmiSrc.HeaderSize), os.SEEK_CUR)
	r.BitsSrc = make([]byte, r.cbBitsSrc)
	if _, err := reader.Read(r.BitsSrc); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *bitmapRecord) readImage() image.Image {

	// bytes per pixel
	bpp, ok := map[uint16]int{
		BI_BITCOUNT_1: 0,
		BI_BITCOUNT_3: 1,
		BI_BITCOUNT_5: 3,
		BI_BITCOUNT_4: 2,
		BI_BITCOUNT_6: 4,
	}[r.BmiSrc.BitCount]

	if !ok {
		fmt.Fprintln(os.Stderr, "emf: unsupported bitmap type", r.BmiSrc.BitCount)
		return nil
	}

	// src image width and height
	width, height := int(r.BmiSrc.Width), int(r.BmiSrc.Height)
	// bytes per line with padding to 4 bytes
	bpl := ((width*int(r.BmiSrc.BitCount) + 31) & 0xFFFFFFE0) / 8

	switch r.BmiSrc.BitCount {
	case BI_BITCOUNT_1:
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		// we don't read it
		colors := map[int][]byte{0: []byte{0, 0, 0, 0}, 1: []byte{255, 255, 255, 0}}

		mask := []byte{0x80, 0x40, 0x20, 0x10, 0x08, 0x04, 0x02, 0x01}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				p := img.PixOffset(x, height-y-1)
				c := colors[0]
				if (r.BitsSrc[y*bpl+x/8] & mask[x%8]) > 0 {
					c = colors[1]
				}
				img.Pix[p+0] = c[2]
				img.Pix[p+1] = c[1]
				img.Pix[p+2] = c[0]
				img.Pix[p+3] = 0xff
			}
		}
		return img

	case BI_BITCOUNT_3:
		img := image.NewGray(image.Rect(0, 0, width, height))
		ix := 0
		// BMP images are stored bottom-up
		for y := height - 1; y >= 0; y-- {
			b := r.BitsSrc[y*bpl : y*bpl+bpl]
			p := img.Pix[ix*img.Stride : ix*img.Stride+img.Stride]
			for i, j := 0, 0; i < len(p); i, j = i+1, j+bpp {
				p[i] = b[j]
			}
			ix = ix + 1
		}
		return img

	case BI_BITCOUNT_4:
		if r.BmiSrc.Compression != BI_RGB {
			fmt.Fprintln(os.Stderr, "emf: unsupported compression type", r.BmiSrc.Compression)
			return nil
		}

		img := image.NewRGBA(image.Rect(0, 0, width, height))
		ix := 0
		// BMP images are stored bottom-up
		for y := height - 1; y >= 0; y-- {
			b := r.BitsSrc[y*bpl : y*bpl+bpl]
			p := img.Pix[ix*img.Stride : ix*img.Stride+img.Stride]
			for i, j := 0, 0; i < len(p); i, j = i+4, j+bpp {
				// The relative intensities of red, green, and blue
				// are represented with 5 bits for each color component.
				c := uint16(b[j+1])<<8 | uint16(b[j])
				p[i+0] = uint8((c>>10)&0x001f) * 8
				p[i+1] = uint8((c>>5)&0x001f) * 8
				p[i+2] = uint8(c&0x001f) * 8
				p[i+3] = 0xff
			}
			ix = ix + 1
		}
		return img

	case BI_BITCOUNT_5, BI_BITCOUNT_6:
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		ix := 0
		// BMP images are stored bottom-up
		for y := height - 1; y >= 0; y-- {
			b := r.BitsSrc[y*bpl : y*bpl+bpl]
			p := img.Pix[ix*img.Stride : ix*img.Stride+img.Stride]
			for i, j := 0, 0; i < len(p); i, j = i+4, j+bpp {
				// color in BMP stored in BGR order
				p[i+0] = b[j+2]
				p[i+1] = b[j+1]
				p[i+2] = b[j+0]
				p[i+3] = 0xff
			}
			ix = ix + 1
		}
		return img
	}
	return nil
}

func (r *bitmapRecord) Draw(ctx *context) {
	img := r.readImage()
	if img == nil {
		return
	}

	// dest image rectangle
	tx, ty := ctx.GetMatrixTransform().GetTranslation()
	rect := image.Rect(
		int(r.Bounds.Left)+int(tx), int(r.Bounds.Top)+int(ty),
		int(r.Bounds.Right)+int(tx), int(r.Bounds.Bottom)+int(ty))

	// Record bounds often differs for 1px with image size.
	// Call scaling only if image size is bigger than record bounds because
	// this procedure is very expensive.
	if img.Bounds().Dx() > rect.Dx()+1 && img.Bounds().Dy() > rect.Dy()+1 {
		img = imaging.Resize(img, rect.Dx(), rect.Dy(), imaging.CatmullRom)
	}

	draw.Draw(ctx.img, rect, img, image.ZP, draw.Over)
}

type BitbltRecord struct {
	bitmapRecord
}

func readBitbltRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &BitbltRecord{}
	r.Record = Record{Type: EMR_BITBLT, Size: size}
	return r.read(reader)
}

type StretchbltRecord struct {
	bitmapRecord
}

func readStretchbltRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &StretchbltRecord{}
	r.Record = Record{Type: EMR_STRETCHBLT, Size: size}
	return r.read(reader)
}

type StretchdibitsRecord struct {
	// brings two unused fields: XformSrc and BkColorSrc
	bitmapRecord
}

func readStretchdibitsRecord(reader *bytes.Reader, size uint32) (Recorder, error) {
	r := &StretchdibitsRecord{}
	r.Record = Record{Type: EMR_STRETCHDIBITS, Size: size}

	if err := binary.Read(reader, binary.LittleEndian, &r.Bounds); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.xDest); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.yDest); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.xSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.ySrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cxSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cySrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.offBmiSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cbBmiSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.offBitsSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cbBitsSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.UsageSrc); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.BitBltRasterOperation); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cxDest); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &r.cyDest); err != nil {
		return nil, err
	}

	// BitmapBuffer
	// skipping UndefinedSpace1
	reader.Seek(int64(r.offBmiSrc-80), os.SEEK_CUR)
	if err := binary.Read(reader, binary.LittleEndian, &r.BmiSrc); err != nil {
		return nil, err
	}

	// skipping UndefinedSpace2
	reader.Seek(int64(r.offBitsSrc-80-r.BmiSrc.HeaderSize), os.SEEK_CUR)
	r.BitsSrc = make([]byte, r.cbBitsSrc)
	if _, err := reader.Read(r.BitsSrc); err != nil {
		return nil, err
	}

	return r, nil
}
