package bit

// WithLSB returns the given byte with the least significant bit (LSB) set to
// the given bit value, while true means 1 and false means 0.
func WithLSB(b byte, bit bool) byte {
	if bit {
		return b | 1
	} else {
		return b & 0xFE
	}
}

// GetLSB given a byte, will return the least significant bit of that byte
func GetLSB(b byte) bool {
	return b%2 != 0
}
