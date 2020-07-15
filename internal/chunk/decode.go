package chunk

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"log"
	"path"
)

func Decode(filepath string) error {

	log.Println("Opening image:", filepath)
	probeImg, err := OpenImageFile(filepath)
	if err != nil {
		return err
	}

	log.Println("Calculating bounds...")
	bounds := CalculateChunkBounds(probeImg)

	log.Println("Calculating Merkle tree roots for every chunk...")

	// rootHashes is a map from the root hash of a chunk to a list of indices where this root hash can be found
	rootHashes := map[string][]ChunkIndex{}
	for x, boundRow := range bounds {
		for y, bound := range boundRow {

			chunk := &Chunk{
				RGBA: ImageToRGBA(probeImg.SubImage(bound)),
			}

			// First byte contains the number of hashes in this chunk (called paths in the merkletree package)
			pathCount := make([]byte, 1)
			_, err := chunk.Read(pathCount)
			if err != nil {
				return err
			}

			chunkHash, _ := chunk.CalculateHash()
			prevHash := chunkHash
			for i := 0; i < int(pathCount[0]); i++ {
				// The order in which the hashes should be concatenated to calculate the composite hash
				side := make([]byte, 1)

				// The hash data for the new composite hash
				data := make([]byte, 32)

				// EOFs can happen if pathCount is wrong due to image manipulation
				// of that specific chunk. pathCount could be way larger than
				// the maximum chunk payload, therefore an EOF can happen.
				_, err := chunk.Read(side)
				if err != nil {
					break
				}

				_, err = chunk.Read(data)
				if err != nil {
					break
				}

				hsh := sha256.New()

				if side[0] == 0 {
					prevHash = append(data, prevHash...)
				} else if side[0] == 1 {
					prevHash = append(prevHash, data...)
				} else {
					break
				}

				hsh.Write(prevHash)
				prevHash = hsh.Sum(nil)
			}

			// This is the root hash for the chunk at hand
			rootHash := hex.EncodeToString(prevHash)

			// persist chunk root hash index to color it in later on if there is only one with this hash.
			if _, exists := rootHashes[rootHash]; !exists {
				rootHashes[rootHash] = []ChunkIndex{}
			}
			rootHashes[rootHash] = append(rootHashes[rootHash], ChunkIndex{x, y})
		}
	}

	// Find the root hash that appeared multiple times
	rootCount := 0
	merkleRoot := ""
	for chunkRoot, indices := range rootHashes {
		if len(indices) > rootCount {
			rootCount = len(indices)
			merkleRoot = chunkRoot
		}
	}

	if len(rootHashes) == 1 {
		log.Println("This image has not been tampered with. All chunks have the same Merkle Root:", merkleRoot)
		return nil
	}

	log.Println("Found multiple Merkle Roots. This image has been tampered with! RootHashes:")

	log.Println("Count\tRoot")
	for root, indexes := range rootHashes {
		log.Printf("%5d\t%s\n", len(indexes), root)
	}

	log.Println("Drawing overlay image of altered regions...")

	overlayImg := ImageToRGBA(probeImg.SubImage(probeImg.Bounds()))
	for root, indices := range rootHashes {
		for _, idx := range indices {

			if root == merkleRoot {
				continue
			}

			draw.DrawMask(
				overlayImg,
				bounds[idx.x][idx.y],
				&image.Uniform{C: color.RGBA{R: 255, A: 255}},
				image.Point{},
				&image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 80}},
				image.Point{},
				draw.Over,
			)
		}
	}

	overlayFilepath := path.Join(path.Dir(filepath), SetExtension(path.Base(filepath), ".overlay.png"))
	log.Println("Saving overlay image:", overlayFilepath)
	err = SaveImageFile(overlayFilepath, overlayImg)
	if err != nil {
		return err
	}

	return nil
}

// ChunkIndex holds the index of a chunk in the bounds map.
type ChunkIndex struct {
	x int
	y int
}
