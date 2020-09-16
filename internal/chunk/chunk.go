package chunk

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"image"
	"io"

	"dennis-tra/image-stego/pkg/bit"

	"github.com/cbergoon/merkletree"
	"github.com/icza/bitio"
)

// Chunk is a wrapper around an image.RGBA struct that keeps track of
// the read and written bytes to the least significant bits of the underlying *image.RGBA.
type Chunk struct {
	*image.RGBA

	// The number of read bytes. Subsequent calls to read will continue where the last read left off.
	rOff int

	// The number of written bytes. Subsequent calls to write will continue where the last write left off.
	wOff int
}

// MaxPayloadSize returns the maximum number of bytes that can be written to this chunk
func (c *Chunk) MaxPayloadSize() int {
	return c.LSBCount() / 8
}

// Width is a short hand to return the width in pixels of the chunk
func (c *Chunk) Width() int {
	return c.Bounds().Dx()
}

// Height is a short hand to return the height in pixels of the chunk
func (c *Chunk) Height() int {
	return c.Bounds().Dy()
}

// PixelCount returns the total number of pixels
func (c *Chunk) PixelCount() int {
	return len(c.Pix) / 4
}

// LSBCount returns the total number of least significant bits (LSB) available for encoding a message.
// Currently only the RGB values are considered not the A.
func (c *Chunk) LSBCount() int {
	return c.PixelCount() * BitsPerPixel
}

// MinX in this context returns the starting value for iterating over the horizontal axis of the image
func (c *Chunk) MinX() int {
	return c.Bounds().Min.X
}

// MaxX in this context returns the ending value for iterating over the horizontal axis of the image
func (c *Chunk) MaxX() int {
	return c.Bounds().Max.X
}

// MinY in this context returns the starting value for iterating over the vertical axis of the image
func (c *Chunk) MinY() int {
	return c.Bounds().Min.Y
}

// MaxY in this context returns the ending value for iterating over the vertical axis of the image
func (c *Chunk) MaxY() int {
	return c.Bounds().Max.Y
}

// CalculateHash calculates the SHA256 hash of the 7 most significant bits. The least
// significant bit (LSB) is not considered in the hash generation as it is used to
// store the (derived) Merkle leaves/nodes.
// Note: From an implementation point of view the LSB is actually considered but
// always overwritten by a 0.
// This method (among Equal) lets Chunk conform to the merkletree.Content interface.
func (c *Chunk) CalculateHash() ([]byte, error) {

	h := sha256.New()

	for x := c.MinX(); x < c.MaxX(); x++ {
		for y := c.MinY(); y < c.MaxY(); y++ {

			rgba := c.RGBAAt(x, y)

			byt := []byte{
				bit.WithLSB(rgba.R, false),
				bit.WithLSB(rgba.G, false),
				bit.WithLSB(rgba.B, false),
			}
			if _, err := h.Write(byt); err != nil {
				return nil, err
			}
		}
	}

	return h.Sum(nil), nil
}

// Write writes the given bytes to the least significant bits of the chunk.
// It returns the number of bytes written from p and an error if one occurred.
// Consult the io.Writer documentation for the intended behaviour of this function.
// A byte from p is either written completely or not at all to the least significant bits.
// Subsequent calls to write will continue were the last write left off.
func (c *Chunk) Write(p []byte) (n int, err error) {
	r := bitio.NewReader(bytes.NewBuffer(p))

	defer func() { c.wOff += n }()

	for i := 0; i < len(p); i++ {

		bitOff := (c.wOff + i) * BitsPerByte

		// Stop early if there is not enough LSB space left
		if bitOff+7 >= len(c.Pix)-len(c.Pix)/4 {
			return n, io.EOF
		}

		// At this point we're sure that a whole byte can still be written
		for j := 0; j < BitsPerByte; j++ {

			bitVal, err := r.ReadBool()
			if err != nil {
				return n, err
			}

			idx := bitOff + j + (bitOff+j)/BitsPerPixel
			c.Pix[idx] = bit.WithLSB(c.Pix[idx], bitVal)
		}

		// As one byte was written increment the counter
		n += 1
	}

	return n, nil
}

// Read reads the amount of bytes given in p from the LSBs of the image chunk.
// It returns the number of bytes read from the least significant bits and an error if one occurred.
// p will contain the contents from the least significant bits after the call has finished.
func (c *Chunk) Read(p []byte) (n int, err error) {

	b := bytes.NewBuffer(p)
	w := bitio.NewWriter(b)

	b.Reset()

	defer func() {
		w.Close()
		c.rOff += n
	}()

	for i := 0; i < len(p); i++ {

		// calculate current read bit offset: static read offset from potential last run plus idx-var times bits in a byte
		bitOff := (c.rOff + i) * BitsPerByte

		// Stop early if there are not enough LSBs left
		if bitOff+BitsPerByte+(bitOff+BitsPerByte)/BitsPerPixel > len(c.Pix) {
			return n, io.EOF
		}

		// At this point we're sure that a whole byte can still be read
		for j := 0; j < BitsPerByte; j++ {

			idx := bitOff + j + (bitOff+j)/BitsPerPixel
			v := bit.GetLSB(c.Pix[idx])

			err := w.WriteBool(v)
			if err != nil {
				return n, err
			}
		}

		// As one whole byte was read increment the counter
		n += 1
	}

	return n, err
}

// Equals tests for equality of two Contents. It only considers the 7 most significant bits since the last bit contains
// the hash data of the other chunks and doesn't count to the equality.
func (c *Chunk) Equals(o merkletree.Content) (bool, error) {

	oc, ok := o.(*Chunk) // other chunk
	if !ok {
		return false, errors.New("invalid type casting")
	}

	if oc.Width() != c.Width() || oc.Height() != c.Height() {
		return false, nil
	}

	for x := c.MinX(); x < c.MaxX(); x++ {
		for y := c.MinY(); y < c.MaxY(); y++ {

			thisColor := c.RGBAAt(x, y)
			otherColor := oc.RGBAAt(x, y)

			if bit.WithLSB(thisColor.R, false) != bit.WithLSB(otherColor.R, false) {
				return false, nil
			}

			if bit.WithLSB(thisColor.G, false) != bit.WithLSB(otherColor.G, false) {
				return false, nil
			}

			if bit.WithLSB(thisColor.B, false) != bit.WithLSB(otherColor.B, false) {
				return false, nil
			}

			if bit.WithLSB(thisColor.A, false) != bit.WithLSB(otherColor.A, false) {
				return false, nil
			}
		}
	}

	return true, nil
}
