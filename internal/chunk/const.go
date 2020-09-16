package chunk

const (
	// BitsPerPixel says how many colors should be used for LSB encoding.
	// 1 - only R, 2 - R and G, 3 - R, G and B, 4 - R, G, B and A
	BitsPerPixel = 3

	// The number of bits occupied by one SHA256 hash.
	HashBitLength = 256

	// The number of bits occupied by the side information of a merkle tree leaf.
	MerkleSideBitLength = 8

	// The number of bits occupied by the information of how many merkle tree leaves are encoded in the chunk.
	PathCountBitLength = 8

	// The number of bits in a byte.
	BitsPerByte = 8
)
