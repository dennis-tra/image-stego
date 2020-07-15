package chunk

import (
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"log"
	"path"

	"github.com/cbergoon/merkletree"
)

func Encode(filepath string, outdir string) error {
	filename := path.Base(filepath)

	log.Println("Opening image:", filepath)
	originalImg, err := OpenImageFile(filepath)
	if err != nil {
		return err
	}

	// copy original image for the checker pattern image to visualize the chunk bounds
	checkerImg := ImageToRGBA(originalImg.SubImage(originalImg.Bounds()))

	list := []merkletree.Content{}

	log.Println("Calculating bounds...")
	bounds := CalculateChunkBounds(originalImg)

	log.Println("Building merkle tree...")
	for _, boundsRow := range bounds {
		for _, bound := range boundsRow {
			list = append(list, &Chunk{
				RGBA: ImageToRGBA(originalImg.SubImage(bound)),
			})
		}
	}

	// Create a new Merkle Tree from the list of Content
	tree, err := merkletree.NewTree(list)
	if err != nil {
		return err
	}
	log.Println("Merkle Tree Root Hash:", hex.EncodeToString(tree.MerkleRoot()))

	log.Println("Drawing checker pattern overlay image...")
	for x, boundRow := range bounds {
		for y, bound := range boundRow {

			var clr color.RGBA
			if (x%2 == 0 && y%2 == 0) || (x%2 != 0 && y%2 != 0) {
				clr = color.RGBA{B: 255, A: 255}
			} else {
				clr = color.RGBA{R: 255, A: 255}
			}

			draw.DrawMask(
				checkerImg,
				bound,
				&image.Uniform{C: clr},
				image.Point{},
				&image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 80}},
				image.Point{},
				draw.Over,
			)
		}
	}

	checkerFilepath := path.Join(outdir, SetExtension(filename, ".checker.png"))
	log.Println("Saving checker pattern overlay image:", checkerFilepath)
	err = SaveImageFile(checkerFilepath, checkerImg)
	if err != nil {
		return err
	}

	log.Println("Encoding Merkle Tree information into LSBs of the image")
	encodedImg := image.NewRGBA(originalImg.Bounds())
	for x, boundsRow := range bounds {
		for y, bound := range boundsRow {

			chunk := list[x*len(boundsRow)+y].(*Chunk)

			paths, sides, err := tree.GetMerklePath(chunk)
			if err != nil {
				return err
			}

			buf := []byte{}
			buf = append(buf, uint8(len(paths)))
			for i, path := range paths {
				side := uint8(sides[i])
				buf = append(buf, side)
				buf = append(buf, path...)
			}

			_, err = chunk.Write(buf)
			if err != nil {
				return err
			}

			draw.Draw(encodedImg, bound, chunk, image.Point{}, draw.Src)
		}
	}

	encodedFilepath := path.Join(outdir, SetExtension(filename, ".png"))
	log.Println("Saving encoded image:", encodedFilepath)
	err = SaveImageFile(encodedFilepath, encodedImg)
	if err != nil {
		return err
	}

	return nil
}
