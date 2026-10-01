// Writes a Windows .ico that contains logo.png scaled to the sizes
// Explorer and the taskbar actually use.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	pngPath := flag.String("png", "logo.png", "source PNG")
	icoPath := flag.String("ico", "logo.ico", "output ICO")
	flag.Parse()
	f, err := os.Open(*pngPath)
	if err != nil {
		fatal(err)
	}
	src, err := png.Decode(f)
	f.Close()
	if err != nil {
		fatal(err)
	}
	var blobs [][]byte
	for _, size := range []int{16, 24, 32, 48, 64, 128, 256} {
		buf := &bytes.Buffer{}
		if err := png.Encode(buf, fit(src, size)); err != nil {
			fatal(err)
		}
		blobs = append(blobs, buf.Bytes())
	}
	if err := os.WriteFile(*icoPath, packICO(blobs), 0o644); err != nil {
		fatal(err)
	}
}

func fit(src image.Image, size int) *image.NRGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		y0 := y * sh / size
		y1 := (y + 1) * sh / size
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < size; x++ {
			x0 := x * sw / size
			x1 := (x + 1) * sw / size
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var r, g, bl, a, n uint64
			for yy := y0; yy < y1; yy++ {
				for xx := x0; xx < x1; xx++ {
					R, G, B, A := src.At(b.Min.X+xx, b.Min.Y+yy).RGBA()
					r += uint64(R)
					g += uint64(G)
					bl += uint64(B)
					a += uint64(A)
					n++
				}
			}
			c := color.NRGBAModel.Convert(color.RGBA64{
				R: uint16(r / n),
				G: uint16(g / n),
				B: uint16(bl / n),
				A: uint16(a / n),
			}).(color.NRGBA)
			dst.SetNRGBA(x, y, c)
		}
	}
	return dst
}

func packICO(pngs [][]byte) []byte {
	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, uint16(0))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint16(len(pngs)))
	offset := 6 + 16*len(pngs)
	for _, p := range pngs {
		w, h := pngSize(p)
		wb, hb := byte(w), byte(h)
		if w >= 256 {
			wb = 0
		}
		if h >= 256 {
			hb = 0
		}
		out.WriteByte(wb)
		out.WriteByte(hb)
		out.WriteByte(0)
		out.WriteByte(0)
		_ = binary.Write(&out, binary.LittleEndian, uint16(1))
		_ = binary.Write(&out, binary.LittleEndian, uint16(32))
		_ = binary.Write(&out, binary.LittleEndian, uint32(len(p)))
		_ = binary.Write(&out, binary.LittleEndian, uint32(offset))
		offset += len(p)
	}
	for _, p := range pngs {
		out.Write(p)
	}
	return out.Bytes()
}

func pngSize(p []byte) (int, int) {
	cfg, err := png.DecodeConfig(bytes.NewReader(p))
	if err != nil {
		fatal(err)
	}
	return cfg.Width, cfg.Height
}

func fatal(err error) {
	os.Stderr.WriteString(err.Error() + "\n")
	os.Exit(1)
}
