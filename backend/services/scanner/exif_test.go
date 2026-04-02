package scanner

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestExtractEXIF_NoEXIF(t *testing.T) {
	// Pass a non-EXIF file (just random bytes)
	data := bytes.NewReader([]byte("not an image file at all"))
	result, err := ExtractEXIF(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Width != nil {
		t.Errorf("width = %v, want nil", result.Width)
	}
	if result.CapturedAt != nil {
		t.Errorf("captured_at = %v, want nil", result.CapturedAt)
	}
	if result.CameraMake != "" {
		t.Errorf("camera_make = %q, want empty", result.CameraMake)
	}
	if result.ISO != nil {
		t.Errorf("iso = %v, want nil", result.ISO)
	}
	if result.GPSLatitude != nil {
		t.Errorf("gps_latitude = %v, want nil", result.GPSLatitude)
	}
}

func TestExtractEXIF_EmptyData(t *testing.T) {
	data := bytes.NewReader([]byte{})
	result, err := ExtractEXIF(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestExtractEXIF_MinimalJPEGNoEXIF(t *testing.T) {
	// Minimal valid JPEG (SOI + EOI markers) — no EXIF segment
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xD9}
	data := bytes.NewReader(jpeg)
	result, err := ExtractEXIF(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CameraMake != "" {
		t.Errorf("camera_make = %q, want empty", result.CameraMake)
	}
}

func TestExtractEXIF_WithEXIFData(t *testing.T) {
	// Build a minimal JPEG with EXIF APP1 segment containing IFD0 tags
	jpeg := buildTestJPEGWithEXIF(t)
	data := bytes.NewReader(jpeg)
	result, err := ExtractEXIF(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The minimal EXIF we build includes Make and Model
	if result.CameraMake != "TestMake" {
		t.Errorf("camera_make = %q, want %q", result.CameraMake, "TestMake")
	}
	if result.CameraModel != "TestModel" {
		t.Errorf("camera_model = %q, want %q", result.CameraModel, "TestModel")
	}
}

// buildTestJPEGWithEXIF constructs a minimal JPEG with an APP1/EXIF segment
// containing Make and Model IFD tags.
func buildTestJPEGWithEXIF(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer

	// JPEG SOI marker
	buf.Write([]byte{0xFF, 0xD8})

	// Build EXIF/TIFF data
	var tiff bytes.Buffer

	// TIFF header: little-endian
	tiff.Write([]byte("II"))                           // byte order: little-endian
	binary.Write(&tiff, binary.LittleEndian, uint16(42)) // magic number
	binary.Write(&tiff, binary.LittleEndian, uint32(8))  // offset to IFD0

	// IFD0 at offset 8
	makeVal := "TestMake\x00"   // null-terminated
	modelVal := "TestModel\x00" // null-terminated

	numEntries := uint16(2)
	binary.Write(&tiff, binary.LittleEndian, numEntries)

	// IFD entry size: 12 bytes each
	// After IFD: 2 (count) + 2*12 (entries) + 4 (next IFD) = 30 bytes
	// Data starts at offset 8 + 30 = 38
	dataOffset := uint32(8 + 2 + 2*12 + 4)

	// Tag 0x010F = Make (271), type=2 (ASCII), count=len, offset
	binary.Write(&tiff, binary.LittleEndian, uint16(0x010F))
	binary.Write(&tiff, binary.LittleEndian, uint16(2)) // ASCII
	binary.Write(&tiff, binary.LittleEndian, uint32(len(makeVal)))
	binary.Write(&tiff, binary.LittleEndian, dataOffset)

	// Tag 0x0110 = Model (272), type=2 (ASCII), count=len, offset
	binary.Write(&tiff, binary.LittleEndian, uint16(0x0110))
	binary.Write(&tiff, binary.LittleEndian, uint16(2)) // ASCII
	binary.Write(&tiff, binary.LittleEndian, uint32(len(modelVal)))
	binary.Write(&tiff, binary.LittleEndian, dataOffset+uint32(len(makeVal)))

	// Next IFD offset (0 = no more IFDs)
	binary.Write(&tiff, binary.LittleEndian, uint32(0))

	// Write data values
	tiff.WriteString(makeVal)
	tiff.WriteString(modelVal)

	// Build APP1 segment
	exifHeader := []byte("Exif\x00\x00")
	app1Data := append(exifHeader, tiff.Bytes()...)
	app1Len := uint16(len(app1Data) + 2) // +2 for the length field itself

	buf.Write([]byte{0xFF, 0xE1}) // APP1 marker
	binary.Write(&buf, binary.BigEndian, app1Len)
	buf.Write(app1Data)

	// JPEG EOI marker
	buf.Write([]byte{0xFF, 0xD9})

	return buf.Bytes()
}
