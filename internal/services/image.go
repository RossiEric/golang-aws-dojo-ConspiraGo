package services

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"

	kingpin "github.com/alecthomas/kingpin/v2"
)

var (
	windowSize  = kingpin.Flag("size", "Window size as a percentage.").Short('s').Default("5").Float64()
	percentile  = kingpin.Flag("percentile", "Window percentile.").Short('p').Default("90").Float64()
	targetValue = kingpin.Flag("target", "Target value when scaling output.").Short('t').Default("240").Int()
	files       = kingpin.Arg("files", "Images to process.").Required().ExistingFiles()
)

type Bound struct {
	Left   int
	Top    int
	Bottom int
	Right  int
}

type Image interface {
	RemoveTransparency(ctx context.Context, reader io.Reader) (io.Reader, error)
	Slice(ctx context.Context, reader io.Reader, bound Bound) (io.Reader, error)
}

type imageImpl struct {
}

// Slice implements Image.
func (*imageImpl) Slice(ctx context.Context, reader io.Reader, bound Bound) (io.Reader, error) {
	top := bound.Top
	left := bound.Left
	bottom := bound.Bottom
	right := bound.Right

	im, _, err := image.Decode(reader)

	if err != nil {
		return nil, err
	}

	bounds := im.Bounds()

	isValid := image.Rect(int(left), int(top), int(right), int(bottom)).In(bounds)

	if !isValid {
		return nil, errors.New("invalid slice")
	}

	dst := image.NewRGBA(image.Rect(left, top, right, bottom))

	// Copia os pixels da região da imagem original para a nova imagem
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			srcColor := im.At(x, y)
			dst.Set(x, y, srcColor)
		}
	}

	var resultBuffer bytes.Buffer
	err = png.Encode(&resultBuffer, dst) // Utilizando o encoder PNG
	if err != nil {
		return nil, err
	}

	return &resultBuffer, nil
}

// RemoveTransparency implements imageImpl.
func (t *imageImpl) RemoveTransparency(ctx context.Context, reader io.Reader) (io.Reader, error) {
	im, _, err := image.Decode(reader)

	if err != nil {
		return nil, err
	}

	return t.processFile(im)
}

func (t *imageImpl) ensureGray(im image.Image) (*image.Gray, bool) {
	switch im := im.(type) {
	case *image.Gray:
		return im, true
	default:
		dst := image.NewGray(im.Bounds())
		draw.Draw(dst, im.Bounds(), im, image.ZP, draw.Src)
		return dst, false
	}
}

func (t *imageImpl) histogramPercentile(hist []int, n int, p float64) (int, error) {
	if p <= 0.5 {
		m := int(float64(n) * p)
		for v, c := range hist {
			m -= c
			if m <= 0 {
				return v, nil
			}
		}
	} else {
		m := int(float64(n) * (1 - p))
		for v := 255; v >= 0; v-- {
			m -= hist[v]
			if m <= 0 {
				return v, nil
			}
		}
	}

	return 0, errors.New("histogramPercentile: invalid percentile")
}

func (t *imageImpl) columnPercentiles(im *image.Gray, p float64, x, r int) ([]int, error) {
	b := im.Bounds()
	x0 := x - r
	x1 := x + r + 1
	if x0 < b.Min.X {
		x0 = b.Min.X
	}
	if x1 > b.Max.X {
		x1 = b.Max.X
	}
	y0 := b.Min.Y
	y1 := b.Max.Y
	result := make([]int, b.Dy())
	hist := make([]int, 256)
	n := 0
	for y := y0; y < y0+r; y++ {
		i := im.PixOffset(x0, y)
		for x := x0; x < x1; x++ {
			hist[im.Pix[i]]++
			i++
			n++
		}
	}
	for y := y0 + r; y < y1; y++ {
		yy := y - r - r
		if yy >= 0 {
			i := im.PixOffset(x0, yy)
			for x := x0; x < x1; x++ {
				hist[im.Pix[i]]--
				i++
				n--
			}
		}
		i := im.PixOffset(x0, y)
		for x := x0; x < x1; x++ {
			hist[im.Pix[i]]++
			i++
			n++
		}

		val, err := t.histogramPercentile(hist, n, p)

		if err != nil {
			return []int{}, err
		}

		result[y-r] = val
	}
	for y := y1; y < y1+r; y++ {
		yy := y - r - r
		i := im.PixOffset(x0, yy)
		for x := x0; x < x1; x++ {
			hist[im.Pix[i]]--
			i++
			n--
		}

		val, err := t.histogramPercentile(hist, n, p)

		if err != nil {
			return []int{}, err
		}

		result[y-r] = val
	}
	return result, nil
}

func (t *imageImpl) imagePercentile(im *image.Gray, p float64) (int, error) {
	hist := make([]int, 256)
	b := im.Bounds()
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		i := im.PixOffset(b.Min.X, y)
		for x := b.Min.X; x < b.Max.X; x++ {
			hist[im.Pix[i]]++
			i++
			n++
		}
	}
	return t.histogramPercentile(hist, n, p)
}

func (t *imageImpl) processFile(src image.Image) (io.Reader, error) {
	s := *windowSize / 100
	p := *percentile / 100
	t1 := float64(*targetValue)

	im, _ := t.ensureGray(src)
	dst := image.NewGray(im.Bounds())
	gradient := image.NewGray(im.Bounds())
	level := image.NewGray(im.Bounds())

	w := im.Bounds().Size().X
	h := im.Bounds().Size().Y
	size := int(math.Sqrt(float64(w*h))*s + 0.5)

	for x := 0; x < w; x++ {
		column, err := t.columnPercentiles(im, p, x, size/2)

		if err != nil {
			return nil, err
		}

		for y, a := range column {
			i := im.PixOffset(x, y)
			v := float64(im.Pix[i])
			v = (v / float64(a)) * t1
			if v < 0 {
				v = 0
			}
			if v > 255 {
				v = 255
			}
			dst.Pix[i] = uint8(v)
			gradient.Pix[i] = uint8(a)
		}
	}

	a, err := t.imagePercentile(dst, 0.0001)

	if err != nil {
		return nil, err
	}

	lo := float64(a)

	b, err := t.imagePercentile(dst, 0.97)

	if err != nil {
		return nil, err
	}

	hi := float64(b)

	m := 255 / (hi - lo)

	for i, v := range dst.Pix {
		nv := int((float64(v)-lo)*m + 0.5)
		if nv < 0 {
			nv = 0
		}
		if nv > 255 {
			nv = 255
		}
		level.Pix[i] = uint8(nv)
	}

	var buffer bytes.Buffer

	err = png.Encode(&buffer, im)

	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buffer.Bytes()), nil

}

func NewImage() Image {
	return &imageImpl{}
}
