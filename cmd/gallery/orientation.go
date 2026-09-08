package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"os"
)

// jpegOrientation reads only the EXIF orientation; no metadata is copied to previews.
func jpegOrientation(raw []byte) int {
	if len(raw) < 4 || raw[0] != 0xff || raw[1] != 0xd8 {
		return 1
	}
	for pos := 2; pos+4 <= len(raw); {
		if raw[pos] != 0xff {
			return 1
		}
		marker := raw[pos+1]
		if marker == 0xda || marker == 0xd9 {
			return 1
		}
		size := int(binary.BigEndian.Uint16(raw[pos+2 : pos+4]))
		if size < 2 || pos+2+size > len(raw) {
			return 1
		}
		data := raw[pos+4 : pos+2+size]
		if marker == 0xe1 && bytes.HasPrefix(data, []byte("Exif\x00\x00")) {
			return tiffOrientation(data[6:])
		}
		pos += 2 + size
	}
	return 1
}

func tiffOrientation(data []byte) int {
	if len(data) < 8 {
		return 1
	}
	var order binary.ByteOrder
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}
	if order.Uint16(data[2:4]) != 42 {
		return 1
	}
	offset := uint64(order.Uint32(data[4:8]))
	if offset+2 > uint64(len(data)) {
		return 1
	}
	count := int(order.Uint16(data[offset : offset+2]))
	for i := 0; i < count; i++ {
		pos := offset + 2 + uint64(i)*12
		if pos+12 > uint64(len(data)) {
			return 1
		}
		entry := data[pos : pos+12]
		if order.Uint16(entry[:2]) == 274 && order.Uint16(entry[2:4]) == 3 && order.Uint32(entry[4:8]) == 1 {
			value := int(order.Uint16(entry[8:10]))
			if value >= 1 && value <= 8 {
				return value
			}
			return 1
		}
	}
	return 1
}

func orientImage(src image.Image, orientation int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dw, dh := w, h
	if orientation >= 5 {
		dw, dh = h, w
	}
	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := x, y
			switch orientation {
			case 2:
				dx = w - 1 - x
			case 3:
				dx, dy = w-1-x, h-1-y
			case 4:
				dy = h - 1 - y
			case 5:
				dx, dy = y, x
			case 6:
				dx, dy = h-1-y, x
			case 7:
				dx, dy = h-1-y, w-1-x
			case 8:
				dx, dy = y, w-1-x
			}
			dst.Set(dx, dy, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func orientedPNG(raw []byte, orientation int) (string, error) {
	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp("", "gallery-orientation-*.png")
	if err != nil {
		return "", err
	}
	encodeErr := png.Encode(tmp, orientImage(src, orientation))
	closeErr := tmp.Close()
	if encodeErr != nil || closeErr != nil {
		_ = os.Remove(tmp.Name()) // Best-effort cleanup after a failed temporary conversion.
		if encodeErr != nil {
			return "", encodeErr
		}
		return "", closeErr
	}
	return tmp.Name(), nil
}
