package imaging

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	errNoEmbeddedJPEG = errors.New("no embedded JPEG found")
	errInvalidTIFF    = errors.New("invalid TIFF header")
)

type jpegBlob struct {
	offset uint32
	length uint32
}

// ExtractEmbeddedJPEG parses the TIFF/IFD structure of an ARW or DNG file
// and returns a reader over the largest embedded JPEG preview.
func ExtractEmbeddedJPEG(r io.ReadSeeker) (io.Reader, error) {
	bo, err := readByteOrder(r)
	if err != nil {
		return nil, err
	}

	// Read magic number (42 for TIFF)
	var magic uint16
	if err := binary.Read(r, bo, &magic); err != nil {
		return nil, fmt.Errorf("read magic: %w", err)
	}
	if magic != 42 {
		return nil, errInvalidTIFF
	}

	// Read offset to first IFD
	var ifdOffset uint32
	if err := binary.Read(r, bo, &ifdOffset); err != nil {
		return nil, fmt.Errorf("read ifd0 offset: %w", err)
	}

	var blobs []jpegBlob

	// Walk IFD chain
	for ifdOffset != 0 {
		found, subIFDs, nextOffset, err := readIFD(r, bo, ifdOffset)
		if err != nil {
			break
		}
		blobs = append(blobs, found...)

		// Also check SubIFDs for larger previews
		for _, subOffset := range subIFDs {
			subFound, _, _, err := readIFD(r, bo, subOffset)
			if err != nil {
				continue
			}
			blobs = append(blobs, subFound...)
		}

		ifdOffset = nextOffset
	}

	if len(blobs) == 0 {
		return nil, errNoEmbeddedJPEG
	}

	// Pick the largest blob
	best := blobs[0]
	for _, b := range blobs[1:] {
		if b.length > best.length {
			best = b
		}
	}

	// Verify JPEG SOI marker
	if _, err := r.Seek(int64(best.offset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek to jpeg: %w", err)
	}
	var soi [2]byte
	if _, err := io.ReadFull(r, soi[:]); err != nil {
		return nil, fmt.Errorf("read soi: %w", err)
	}
	if soi[0] != 0xFF || soi[1] != 0xD8 {
		return nil, errNoEmbeddedJPEG
	}

	// Return a section reader over the JPEG data
	if _, err := r.Seek(int64(best.offset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek to jpeg start: %w", err)
	}
	return io.LimitReader(r, int64(best.length)), nil
}

func readByteOrder(r io.Reader) (binary.ByteOrder, error) {
	var bom [2]byte
	if _, err := io.ReadFull(r, bom[:]); err != nil {
		return nil, fmt.Errorf("read byte order: %w", err)
	}
	switch string(bom[:]) {
	case "II":
		return binary.LittleEndian, nil
	case "MM":
		return binary.BigEndian, nil
	default:
		return nil, errInvalidTIFF
	}
}

const (
	tagJPEGOffset    = 0x0201 // JPEGInterchangeFormat
	tagJPEGLength    = 0x0202 // JPEGInterchangeFormatLength
	tagSubIFDs = 0x014A
)

// readIFD reads an IFD at the given offset and returns:
// - found JPEG blobs
// - SubIFD offsets to follow
// - offset to the next IFD (0 if none)
func readIFD(r io.ReadSeeker, bo binary.ByteOrder, offset uint32) ([]jpegBlob, []uint32, uint32, error) {
	if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, nil, 0, err
	}

	var numEntries uint16
	if err := binary.Read(r, bo, &numEntries); err != nil {
		return nil, nil, 0, err
	}

	// Sanity check
	if numEntries > 500 {
		return nil, nil, 0, fmt.Errorf("too many IFD entries: %d", numEntries)
	}

	var jpegOffset, jpegLength uint32
	var hasJPEGOffset, hasJPEGLength bool
	var subIFDs []uint32

	for i := 0; i < int(numEntries); i++ {
		var tag, typ uint16
		var count, value uint32
		if err := binary.Read(r, bo, &tag); err != nil {
			return nil, nil, 0, err
		}
		if err := binary.Read(r, bo, &typ); err != nil {
			return nil, nil, 0, err
		}
		if err := binary.Read(r, bo, &count); err != nil {
			return nil, nil, 0, err
		}
		if err := binary.Read(r, bo, &value); err != nil {
			return nil, nil, 0, err
		}

		switch tag {
		case tagJPEGOffset:
			jpegOffset = value
			hasJPEGOffset = true
		case tagJPEGLength:
			jpegLength = value
			hasJPEGLength = true
		case tagSubIFDs:
			if count == 1 {
				subIFDs = append(subIFDs, value)
			} else {
				// Multiple SubIFD offsets stored at the value offset
				curPos, _ := r.Seek(0, io.SeekCurrent)
				if _, err := r.Seek(int64(value), io.SeekStart); err == nil {
					for j := uint32(0); j < count && j < 10; j++ {
						var subOff uint32
						if err := binary.Read(r, bo, &subOff); err != nil {
							break
						}
						subIFDs = append(subIFDs, subOff)
					}
				}
				r.Seek(curPos, io.SeekStart)
			}
		}
	}

	var blobs []jpegBlob
	if hasJPEGOffset && hasJPEGLength && jpegLength > 0 {
		blobs = append(blobs, jpegBlob{offset: jpegOffset, length: jpegLength})
	}

	// Read next IFD offset
	var nextOffset uint32
	if err := binary.Read(r, bo, &nextOffset); err != nil {
		nextOffset = 0
	}

	return blobs, subIFDs, nextOffset, nil
}
