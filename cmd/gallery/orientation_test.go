package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestOrientImage(t *testing.T) {
	src := image.NewNRGBA(image.Rect(2, 4, 5, 6))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			src.SetNRGBA(x+2, y+4, color.NRGBA{R: uint8(y*3 + x + 1), A: 255})
		}
	}
	want := [][]uint8{
		{1, 2, 3, 4, 5, 6}, {3, 2, 1, 6, 5, 4},
		{6, 5, 4, 3, 2, 1}, {4, 5, 6, 1, 2, 3},
		{1, 4, 2, 5, 3, 6}, {4, 1, 5, 2, 6, 3},
		{6, 3, 5, 2, 4, 1}, {3, 6, 2, 5, 1, 4},
	}
	for i, pixels := range want {
		got := orientImage(src, i+1)
		width, height := 3, 2
		if i >= 4 {
			width, height = 2, 3
		}
		if got.Bounds().Dx() != width || got.Bounds().Dy() != height {
			t.Fatalf("orientation %d: wrong dimensions %v", i+1, got.Bounds())
		}
		for j, p := range pixels {
			r, _, _, _ := got.At(j%width, j/width).RGBA()
			if uint8(r>>8) != p {
				t.Errorf("orientation %d pixel %d: got %d want %d", i+1, j, r>>8, p)
			}
		}
	}
}

func TestJPEGOrientation(t *testing.T) {
	for _, order := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
		for orientation := 1; orientation <= 8; orientation++ {
			data := make([]byte, 26)
			copy(data, "II")
			if order == binary.BigEndian {
				copy(data, "MM")
			}
			order.PutUint16(data[2:], 42)
			order.PutUint32(data[4:], 8)
			order.PutUint16(data[8:], 1)
			order.PutUint16(data[10:], 274)
			order.PutUint16(data[12:], 3)
			order.PutUint32(data[14:], 1)
			order.PutUint16(data[18:], uint16(orientation))
			jpeg := append([]byte{0xff, 0xd8, 0xff, 0xe1, 0, 34}, []byte("Exif\x00\x00")...)
			jpeg = append(jpeg, data...)
			if got := jpegOrientation(jpeg); got != orientation {
				t.Fatalf("%v: got %d want %d", order, got, orientation)
			}
			for n := 0; n < len(jpeg); n++ {
				if got := jpegOrientation(jpeg[:n]); got != 1 {
					t.Fatalf("truncated JPEG: got %d", got)
				}
			}
		}
	}
}

func TestPrunePreviewsKeepsOnlyCurrentGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	keep := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-720.webp"
	obsolete := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-360.webp"
	for _, name := range []string{keep, obsolete, "personal-photo.webp"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := prunePreviews(dir, map[string]bool{keep: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, obsolete)); !os.IsNotExist(err) {
		t.Fatal("obsolete variant retained")
	}
	for _, name := range []string{keep, "personal-photo.webp"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatal("current or unrelated image removed")
		}
	}
}
