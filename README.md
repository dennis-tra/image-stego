# image-stego

Steganography-based image integrity - verify the integrity of individual parts of an image without the need of secondary information.

![Verified image with marked manipulated parts](docs/porsche.overlay.png)

## Motivation

Tamper-proof timestamping of digital content (like images) is based on generating a cryptographic hash of the file and persisting it in a blockchain (e.g. transferring money to a bitcoin address derived from that hash). The corresponding block of that transaction contains a timestamp that is practically impossible to alter. At a later point in time a third party can proof the existence by hashing the data, deriving the bitcoin address and verify him/herself the presence of that address in the blockchain. If the address is found the third party can be sure that the data has not been manipulated after the corresponding timestamp of the block.

Especially if the digital content is an image there is the disadvantage of just needing to change one pixel (actually just one bit) and the resulting hash will be completely different although the original image will be perceptually identical. For example one wouldn't be able to proof that a particular image was created earlier than claimed. In most other use-cases I'm aware of this is a huge advantage though.

## One step forward

While the following approach won't completely solve the aforementioned problem it may be a step forward and may cause thought for others.

What is the approach about in one sentence:

> It uses steganography to embed merkle tree nodes into chunks of an image, so that the integrity of each individual chunk can be verified on its own.

[Steganography](https://en.wikipedia.org/wiki/Steganography) in this context means using the least significant bits (LSB) of the image to encode information. There are other techniques but this was the most straight forward for me to implement.

By having individual chunks one can still proof the integrity of parts of the image while seeing which areas have been tampered with.

## The Approach

As a first step the image is divided in a set of chunks. These cannot be arbitrarily small thouch because the smaller they are the more data needs to be stored in each one but the less storage space each has. There's an optimum of in how many chunks the image should be divided into.

After the chunk count has been calculated the first seven most significant bits of each chunk are hashed. This will result in set of hashes that are now considered as merkle tree leafs. Theses leafs are combined to derive the merkle root hash. This hash can now be embedded into a blockchain.

Each chunk gets now the missing merkle tree information encoded into its least significant bits so that it holds all information necessary to reconstruct the merkle tree root hash.

## Example

Let's consider a squared image that is divided in four equal chunks. The algorithm examines each chunk separately by looping through all pixels and calculating the hash of the seven most significant bits of the 8-Bit RGB (and A) values of each pixel in the chunk. This will give the hash values <img src="https://latex.codecogs.com/svg.latex?H_1" />, <img src="https://latex.codecogs.com/svg.latex?H_2" />, <img src="https://latex.codecogs.com/svg.latex?H_3" /> and <img src="https://latex.codecogs.com/svg.latex?H_4" />. In the picture below the considered bits are printed faint in the top right corner.

![Illustration of the idea](./docs/illustration.png)

Those chunk hashes are now taken as merkle tree leafs and used to construct the merkle tree root hash like in the picture above in the bottom right.

The hash <img src="https://latex.codecogs.com/svg.latex?H_{1234}" /> should be the one to be persisted in a blockchain to be able to proof the existence.

For each chunk to be independently verifiable those merkle nodes are taken that are necessary to reconstruct the merkle root and embeded in the least significant bits of the chunk (denoted in yellow above). The yellow bar at the bottom illustrates the set of least significant bits and data that's saved in them.

E.g. for hash <img src="https://latex.codecogs.com/svg.latex?H_1" /> the hashes <img src="https://latex.codecogs.com/svg.latex?H_2" /> and <img src="https://latex.codecogs.com/svg.latex?H_{34}" /> are necessary to recalculate the merkle tree root hash <img src="https://latex.codecogs.com/svg.latex?H_{1234}" />. The hash <img src="https://latex.codecogs.com/svg.latex?H_1" /> doesn't need to be saved in the LSBs because it can and should be derived from the seven most significant bits of the image data itself.

Having prepared the image in such a way each chunk holds enough information to independently verify its integrity. If an adversery was to manipulate parts of the image the corresponding chunks would become invalidated (e.g. the root hashes wouldn't equal the others) and those chunks can be identified. Other chunks can still be proofen to not having been manipulated which wouldn't have been the case if the just whole image was hashed.

## The Results

Original image (w: 1038px, h: 435px):

![Original image](data/porsche.jpg)

Encoded image:

![Encoded image](docs/porsche.png)

Chunking of the original image:

![Encoded image with chunk overlay](docs/porsche.checker.png)

Manipulated image:

![Manipulated image](docs/porsche.tampered.png)

Decoded image:

![Verified image with marked manipulated parts](docs/porsche.overlay.png)


## Reproduction

To reproduce the results build the tool running the following command:

```shell
go build -o stego cmd/stego/main.go
```

Then take the example image and encode it:

```shell
./stego -e -o="out" data/porsche.jpg
```

Now you'll find the encoded image in the `./out/` folder along with an image that shows the chunks. Run the following command to verify that the image has not been tampered with:

```shell
./stego -d out/porsche.png
```

You should see the following output:

```text
2020/09/16 19:05:44 ...
2020/09/16 19:05:44 This image has not been tampered with. All chunks have the same Merkle Root: 278cba1daf96d84165f8aa69d184e63df5c79f3a4c31cc6864e148c0317c713d
```

Manipulate the image and run the above command again (don't save the image as JPEG as the data in the LSBs wouldn't survive the compression):

```shell
./stego -d out/porsche.png
```

You should see the following output:

```text
...
2020/09/16 19:10:30 Found multiple Merkle Roots. This image has been tampered with! RootHashes:
2020/09/16 19:10:30 Count       Root
2020/09/16 19:10:30     1       1323d18f3aab27b4414535825b9b755c61b19fbd2ce7f2a39688e95e5f32fe15
...
2020/09/16 08:10:30   413       278cba1daf96d84165f8aa69d184e63df5c79f3a4c31cc6864e148c0317c713d
...
2020/09/16 08:10:30 Drawing overlay image of altered regions...
2020/09/16 08:10:30 Saving overlay image: out/porsche.overlay.png
```

## Limitations

There are several limitations that come to my mind I just want to list here:

- Currently only lossless image file formats are supported as least significant bits wouldn't survive a jpeg compression. There are steganography approaches that address exactly this problem though.
- The original image needs to be altered
- It's actually not necessary to embed the merkle tree information in the image itself but to save it separately (maybe header information or a separate file). However having all verification information in one place has its advantages too.
- Cropping is not supported yet because there needs to be a mechanism to find the chunk dimensions independently of the image size.
- The `Chunk` struct implements the Gos `Writer` interface to encode data in the LSBs. This means only whole bytes can be written which leads to wasted space for meta information like: 1. How many hashes are encoded in this chunk, 2. which side should this hash be appended/prepended to to calculate the root hash. Especially the latter information is a simple boolean flag which wastes a whole byte.
- I have no idea of a compelling use case. Maybe the HN community has some ideas. Follow the corresponding posts:


## Timestamped commits

([Verify here](https://verify.originstamp.com/)):

- 9b54dc5b4b912b0c3f5944c1bd7ac008b16beb6e
