// This file is inspired by what is done in the python-escpos and escpos-php libraries.
package escpos

import (
	"fmt"
	"github.com/kovidgoyal/imaging"
	"image"
	"image/color"
)

// printImageDither prints a dithered image to the printer.
// It uses Floyd-Steinberg dithering to convert the image to black and white.
// highDensityVertical and highDensityHorizontal control the density of the image.
// The image is rasterized and converted to a byte array for printing (header included).
// todo: add support for fragmentHeight, center, and maxWidth
func printImageDither(img image.Image, highDensityVertical bool, highDensityHorizontal bool) (data []byte, err error) {
	im, err := transformImage(img)
	if err != nil {
		return nil, err
	}

	densityByte := byte(0)
	if !highDensityHorizontal {
		densityByte += 1
	}
	if !highDensityVertical {
		densityByte += 2
	}

	raster := rasterizeImage(im)

	width, height := im.Bounds().Dx(), im.Bounds().Dy()
	widthBytes := (width + 7) / 8

	header := append([]byte{0x1D}, []byte("v0")...)
	header = append(header, densityByte)

	if res, err := intLowHigh(widthBytes, 2); err != nil {
		return nil, err
	} else {
		header = append(header, res...)
	}

	if res, err := intLowHigh(height, 2); err != nil {
		return nil, err
	} else {
		header = append(header, res...)
	}

	result := append(header, raster...)

	return result, nil
}

// transformImage converts an image to a pure black and white image using Floyd-Steinberg dithering.
func transformImage(imgSource interface{}) (*image.NRGBA, error) {
	var imgOriginal image.Image
	var err error

	switch source := imgSource.(type) {
	case image.Image:
		imgOriginal = source
	case string:
		imgOriginal, err = imaging.Open(source)
		if err != nil {
			return nil, err
		}
	}

	// convert to rgba
	rgba := imaging.Clone(imgOriginal)

	bounds := rgba.Bounds()
	white := imaging.New(bounds.Max.X, bounds.Max.Y, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	// We need to composite the rgba image over the white background using alpha
	result := imaging.OverlayCenter(white, rgba, 1.0)

	// Convert to grayscale
	gray := imaging.Grayscale(result)

	// Invert the image
	result = imaging.Invert(gray)

	// Convert to pure black and white and apply Floyd-Steinberg dithering
	result = applyFloydSteinbergDithering(result)

	return result, nil
}

// applyFloydSteinbergDithering applies Floyd-Steinberg dithering to an image.
// It also converts the image to a binary format (black and white).
// And reverses the colors (black becomes white and vice versa).
func applyFloydSteinbergDithering(img image.Image) *image.NRGBA {
	binary := imaging.New(img.Bounds().Dx(), img.Bounds().Dy(), color.White)
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	errors := make([][]float64, height)
	for i := range errors {
		errors[i] = make([]float64, width)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.At(x, y)
			r, _, _, _ := c.RGBA()
			// Convert from uint32 to float64 (0-255 range)
			oldPixel := float64(r>>8) + errors[y][x]
			newPixel := 0.0
			if oldPixel >= 128 {
				newPixel = 255.0
			}
			// Set the actual pixel
			if newPixel != 0 {
				binary.Set(x, y, color.Black)
			}

			// Distribute the error
			quantError := oldPixel - newPixel
			if x+1 < width {
				errors[y][x+1] += quantError * 7.0 / 16.0
			}
			if y+1 < height {
				if x-1 >= 0 {
					errors[y+1][x-1] += quantError * 3.0 / 16.0
				}
				errors[y+1][x] += quantError * 5.0 / 16.0
				if x+1 < width {
					errors[y+1][x+1] += quantError * 1.0 / 16.0
				}
			}
		}
	}

	return binary
}

// rasterizeImage convert binary image to bytes
func rasterizeImage(img *image.NRGBA) []byte {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// For binary images, we need 1 bit per pixel
	// Calculate bytes needed: width * height / 8 (rounded up)
	bytesPerRow := (width + 7) / 8
	dataSize := bytesPerRow * height
	data := make([]byte, dataSize)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get pixel color
			c := img.At(x, y)
			r, _, _, _ := c.RGBA()

			// If pixel is black (0), set the bit
			// In binary mode, 0 is black, and 1 is white
			if r == 0 {
				// Calculate byte position and bit position within byte
				bytePos := y*bytesPerRow + x/8
				bitPos := uint(7 - (x % 8)) // MSB first

				data[bytePos] |= 1 << bitPos
			}
		}
	}

	return data
}

// intLowHigh generates multiple bytes for a number: In lower and higher parts, or more parts as needed.
func intLowHigh(inpNumber int, outBytes int) ([]byte, error) {
	if outBytes < 1 || outBytes > 4 {
		return nil, fmt.Errorf("can only output 1-4 bytes")
	}

	maxInput := 256<<(outBytes*8) - 1
	if inpNumber < 0 || inpNumber > maxInput {
		return nil, fmt.Errorf("number too large. Can only output up to %d in %d bytes", maxInput, outBytes)
	}

	out := make([]byte, outBytes)
	for i := 0; i < outBytes; i++ {
		out[i] = byte(inpNumber % 256)
		inpNumber /= 256
	}

	return out, nil
}
