package bit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLSB(t *testing.T) {
	tests := []struct {
		byte byte
		want bool
	}{
		{byte: 0b00000000, want: false},
		{byte: 0b00000001, want: true},
		{byte: 0b11111111, want: true},
		{byte: 0b11111110, want: false},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%08b should be %t", tt.byte, tt.want)
		t.Run(name, func(t *testing.T) {
			got := GetLSB(tt.byte)
			assert.Equal(t, tt.want, got, "GetLSB() = %v, want %v", got, tt.want)
		})
	}
}

func TestWithLSB(t *testing.T) {
	tests := []struct {
		byte byte
		bit  bool
		want byte
	}{
		{byte: 0b00000000, bit: false, want: 0b00000000},
		{byte: 0b00000000, bit: true, want: 0b00000001},
		{byte: 0b11111111, bit: true, want: 0b11111111},
		{byte: 0b11111111, bit: false, want: 0b11111110},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("setting LSB of %08b to %t should be %08b", tt.byte, tt.bit, tt.want)
		t.Run(name, func(t *testing.T) {
			got := WithLSB(tt.byte, tt.bit)
			assert.Equal(t, tt.want, got, "WithLSB() = %v, want %v", got, tt.want)
		})
	}
}
