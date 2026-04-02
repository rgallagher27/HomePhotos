package imaging

import (
	"bytes"
	"encoding/binary"
	"image/jpeg"
	"testing"
)

// buildSyntheticTIFF creates a minimal TIFF file with an embedded JPEG.
// Structure: TIFF header → IFD0 with JPEGInterchangeFormat/Length tags → JPEG data
func buildSyntheticTIFF(t *testing.T, jpegData []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	bo := binary.LittleEndian

	// TIFF header: byte order (II) + magic (42) + IFD0 offset (8)
	buf.Write([]byte("II"))
	binary.Write(&buf, bo, uint16(42))
	binary.Write(&buf, bo, uint32(8)) // IFD starts right after header

	// IFD0: 2 entries
	numEntries := uint16(2)
	binary.Write(&buf, bo, numEntries)

	// Calculate where JPEG data will be placed
	// IFD0 starts at offset 8
	// IFD size: 2 (num entries) + 2*12 (entries) + 4 (next IFD offset) = 30
	jpegDataOffset := uint32(8 + 2 + 2*12 + 4)

	// Entry 1: JPEGInterchangeFormat (0x0201), type LONG (4), count 1, value = offset
	binary.Write(&buf, bo, uint16(0x0201))
	binary.Write(&buf, bo, uint16(4)) // LONG
	binary.Write(&buf, bo, uint32(1)) // count
	binary.Write(&buf, bo, jpegDataOffset)

	// Entry 2: JPEGInterchangeFormatLength (0x0202), type LONG (4), count 1, value = length
	binary.Write(&buf, bo, uint16(0x0202))
	binary.Write(&buf, bo, uint16(4)) // LONG
	binary.Write(&buf, bo, uint32(1)) // count
	binary.Write(&buf, bo, uint32(len(jpegData)))

	// Next IFD offset (0 = no more IFDs)
	binary.Write(&buf, bo, uint32(0))

	// JPEG data
	buf.Write(jpegData)

	return buf.Bytes()
}

func TestExtractEmbeddedJPEG(t *testing.T) {
	// Create a small JPEG to embed
	img := createTestImage(100, 80)
	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 75}); err != nil {
		t.Fatalf("encode test jpeg: %v", err)
	}
	jpegData := jpegBuf.Bytes()

	tiffData := buildSyntheticTIFF(t, jpegData)
	r := bytes.NewReader(tiffData)

	extracted, err := ExtractEmbeddedJPEG(r)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	// Decode the extracted JPEG to verify it's valid
	decoded, err := jpeg.Decode(extracted)
	if err != nil {
		t.Fatalf("decode extracted: %v", err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 80 {
		t.Errorf("size = %dx%d, want 100x80", bounds.Dx(), bounds.Dy())
	}
}

func TestExtractEmbeddedJPEG_InvalidFile(t *testing.T) {
	garbage := []byte("this is not a tiff file at all")
	r := bytes.NewReader(garbage)
	_, err := ExtractEmbeddedJPEG(r)
	if err == nil {
		t.Error("expected error for invalid file")
	}
}

func TestExtractEmbeddedJPEG_EmptyFile(t *testing.T) {
	r := bytes.NewReader(nil)
	_, err := ExtractEmbeddedJPEG(r)
	if err == nil {
		t.Error("expected error for empty file")
	}
}
