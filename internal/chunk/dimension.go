package chunk

import (
	"image"
	"math"
)

// CalculateChunkBounds takes the given *image.RGBA and calculates the optimal distribution of image chunks
// to encode the merkle tree data.
//
// The more chunks we anticipate the smaller they become, the more of them are there and the more data needs
// to be encoded in each chunk to store all the merkle tree data. So there is an optimum of the number of chunks.
// Basically we want the highest number of chunks where each individual one can still store all the necessary
// merkle information.
//
// The calculation is an iterative process. The calculation starts with the assumption that we want to use
// two chunks to encode the data. First it calculates the required amount of bits to encode all merkle nodes
// within one chunk. Then it calculates the total number of available bits per chunk. In the first iteration
// the number of available bits will usually be much larger than the required bits.
//
// If the amount of required bits exceeds the available least significant bits we stop and are sure we have found
// the maximum number of chunks that this image can be divided into.
//
// Beware that with one merkle tree leaf hash (256 bits) the side of the merkle node (1 byte) needs to be encoded
// and the number of leaf nodes (offset of 8) as well.
//
// As a last step we built a matrix of bounds that represent the chunks in the given image. Since the chunks may
// not divide the side lengths perfectly we need to handle the clipping as well.
func CalculateChunkBounds(rgba *image.RGBA) [][]image.Rectangle {

	chunk := Chunk{RGBA: rgba}

	// Calculate maximum number of chunks that this image can be divided into taken into account
	chunkCount := 0
	for {
		chunkCount += 2
		// neededBitsPerChunk answers the question: How many bits do we need to store the merkle tree leaves if
		// we had chunkCount many chunks. The more chunks -> the more merkle leaves -> the less data can be saved
		// into one chunk.

		// The number of hashes that need to be saved into each chunk based on the total chunk count.
		hashesPerChunk := int(math.Ceil(math.Log2(float64(chunkCount))))
		neededBitsPerChunk := hashesPerChunk*(HashBitLength+MerkleSideBitLength) + PathCountBitLength

		chunkCountX, chunkCountY := chunkDist(chunkCount)

		// guaranteed width and height of each chunk (could be more due to clipping
		chunkWidth := chunk.Width() / chunkCountX
		chunkHeight := chunk.Height() / chunkCountY

		// The available amount of bits in each chunk
		availableBitsPerChunk := chunkWidth * chunkHeight * 3

		// If we need more bits than are available we stop and decrement the chunk count to the last
		// "working" count.
		if neededBitsPerChunk > availableBitsPerChunk {
			chunkCount -= 2
			break
		}
	}

	// Calculate the number of chunks along the width and height
	chunkCountX, chunkCountY := chunkDist(chunkCount)

	// guaranteed width and height of each chunk
	chunkWidth := chunk.Width() / chunkCountX
	chunkHeight := chunk.Height() / chunkCountY

	// Add clippings (the side length to chunk count ratio will likely be rational so we add the remainder to the
	// side lengths equally.
	chunkWidthClippings := chunk.Width() % chunkCountX
	chunkHeightClippings := chunk.Height() % chunkCountY

	bounds := make([][]image.Rectangle, chunkCountX)
	for i := range bounds {
		bounds[i] = make([]image.Rectangle, chunkCountY)
	}

	cxOff := 0
	cyOff := 0
	for cx := 0; cx < chunkCountX; cx++ {

		cw := chunkWidth
		if cx < chunkWidthClippings {
			cw += 1
			cxOff = 0
		} else {
			cxOff = chunkWidthClippings
		}

		for cy := 0; cy < chunkCountY; cy++ {

			ch := chunkHeight
			if cy < chunkHeightClippings {
				ch += 1
				cyOff = 0
			} else {
				cyOff = chunkHeightClippings
			}

			bounds[cx][cy] = image.Rect(cw, ch, 0, 0).Add(image.Pt(cxOff+cx*cw, cyOff+cy*ch))
		}
	}

	return bounds
}

// chunkDist calculates the chunk distribution along the width and height.
// The aim is to get an evenly distributed field of chunks.
func chunkDist(count int) (int, int) {

	// Calculate optimal distribution of chunks along width and height
	factors := primeFactors(count)

	// Number of chunks along the width
	countX := factors[len(factors)-1]

	// Number of chunks along the height
	countY := 1
	if len(factors) > 1 {
		countY = factors[len(factors)-2]
	}

	// Evenly distribute the chunks along bot directions
	for i := len(factors) - 3; i >= 0; i-- {
		if countX > countY {
			countY *= factors[i]
		} else {
			countX *= factors[i]
		}
	}

	return countX, countY
}

// primeFactors returns what the name says ;)
// Src: https://siongui.github.io/2017/05/09/go-find-all-prime-factors-of-integer-number/
func primeFactors(n int) (pfs []int) {
	// Get the number of 2s that divide n
	for n%2 == 0 {
		pfs = append(pfs, 2)
		n = n / 2
	}

	// n must be odd at this point. so we can skip one element
	// (note i = i + 2)
	for i := 3; i*i <= n; i = i + 2 {
		// while i divides n, append i and divide n
		for n%i == 0 {
			pfs = append(pfs, i)
			n = n / i
		}
	}

	// This condition is to handle the case when n is a prime number
	// greater than 2
	if n > 2 {
		pfs = append(pfs, n)
	}

	return
}
