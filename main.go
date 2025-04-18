package escpos

import (
	"bufio"
	"fmt"
	"image"
	"io"

	"github.com/justinmichaelvieira/iconv"
)

// Style defines the text formatting options for the printer
type Style struct {
	Bold          bool
	Width, Height uint8
	Reverse       bool
	Underline     uint8 // can be 0, 1 or 2
	UpsideDown    bool
	Rotate        bool
	Justify       uint8
}

// Justification constants
const (
	JustifyLeft   uint8 = 0
	JustifyCenter uint8 = 1
	JustifyRight  uint8 = 2
)

// QR code error correction levels
const (
	QRCodeErrorCorrectionLevelL uint8 = 48 // 7% recovery capacity
	QRCodeErrorCorrectionLevelM uint8 = 49 // 15% recovery capacity
	QRCodeErrorCorrectionLevelQ uint8 = 50 // 25% recovery capacity
	QRCodeErrorCorrectionLevelH uint8 = 51 // 30% recovery capacity
)

// Barcode types
const (
	BarcodeUPCA    uint8 = 0
	BarcodeUPCE    uint8 = 1
	BarcodeEAN13   uint8 = 2
	BarcodeEAN8    uint8 = 3
	BarcodeCode39  uint8 = 4
	BarcodeITF     uint8 = 5
	BarcodeCodabar uint8 = 6
)

// HRI position constants
const (
	HRIPositionNone  uint8 = 0
	HRIPositionAbove uint8 = 1
	HRIPositionBelow uint8 = 2
	HRIPositionBoth  uint8 = 3
)

// ESC/POS command bytes
const (
	esc byte = 0x1B
	gs  byte = 0x1D
	fs  byte = 0x1C
)

// Image processing method constants
const (
	// ImageProcessDither applies Floyd-Steinberg dithering
	ImageProcessDither uint8 = 0
	// ImageProcessThreshold applies simple threshold-based conversion
	ImageProcessThreshold uint8 = 1
)

// PrinterConfig contains options to disable specific formatting features
type PrinterConfig struct {
	DisableUnderline  bool
	DisableBold       bool
	DisableReverse    bool
	DisableRotate     bool
	DisableUpsideDown bool
	DisableJustify    bool
}

// Escpos represents a ESC/POS printer connection
type Escpos struct {
	dst    *bufio.Writer
	Style  Style
	config PrinterConfig
}

// New creates a new Escpos printer instance
func New(dst io.Writer) *Escpos {
	return &Escpos{
		dst: bufio.NewWriter(dst),
		// Set default style values
		Style: Style{
			Bold:      false,
			Width:     1,
			Height:    1,
			Reverse:   false,
			Underline: 0,
			Justify:   JustifyLeft,
		},
	}
}

// SetConfig sets the printer configuration options
func (e *Escpos) SetConfig(conf PrinterConfig) {
	e.config = conf
}

// Print sends the buffered data to the printer
func (e *Escpos) Print() error {
	if err := e.dst.Flush(); err != nil {
		return fmt.Errorf("failed to send data to printer: %w", err)
	}
	return nil
}

// PrintAndCut sends the buffered data to the printer and performs a cut
func (e *Escpos) PrintAndCut() error {
	_, err := e.Cut()
	if err != nil {
		return fmt.Errorf("failed to perform cut: %w", err)
	}

	if err := e.dst.Flush(); err != nil {
		return fmt.Errorf("failed to send data to printer: %w", err)
	}
	return nil
}

// WriteRaw writes raw bytes directly to the printer
func (e *Escpos) WriteRaw(data []byte) (int, error) {
	if len(data) > 0 {
		return e.dst.Write(data)
	}
	return 0, nil
}

// Write prints a string using the current style settings
func (e *Escpos) Write(data string) (int, error) {
	var err error
	bytesWritten := 0

	// Bold
	if !e.config.DisableBold {
		_, err = e.WriteRaw([]byte{esc, 'E', boolToByte(e.Style.Bold)})
		if err != nil {
			return 0, fmt.Errorf("failed to set bold style: %w", err)
		}
	}

	// Underline
	if !e.config.DisableUnderline {
		_, err = e.WriteRaw([]byte{esc, '-', e.Style.Underline})
		if err != nil {
			return 0, fmt.Errorf("failed to set underline style: %w", err)
		}
	}

	// Reverse
	if !e.config.DisableReverse {
		_, err = e.WriteRaw([]byte{gs, 'B', boolToByte(e.Style.Reverse)})
		if err != nil {
			return 0, fmt.Errorf("failed to set reverse style: %w", err)
		}
	}

	// Rotate
	if !e.config.DisableRotate {
		_, err = e.WriteRaw([]byte{esc, 'V', boolToByte(e.Style.Rotate)})
		if err != nil {
			return 0, fmt.Errorf("failed to set rotate style: %w", err)
		}
	}

	// UpsideDown
	if !e.config.DisableUpsideDown {
		_, err = e.WriteRaw([]byte{esc, '{', boolToByte(e.Style.UpsideDown)})
		if err != nil {
			return 0, fmt.Errorf("failed to set upside-down style: %w", err)
		}
	}

	// Justify
	if !e.config.DisableJustify {
		_, err = e.WriteRaw([]byte{esc, 'a', e.Style.Justify})
		if err != nil {
			return 0, fmt.Errorf("failed to set justification: %w", err)
		}
	}

	// Width / Height
	_, err = e.WriteRaw([]byte{gs, '!', ((e.Style.Width - 1) << 4) | (e.Style.Height - 1)})
	if err != nil {
		return 0, fmt.Errorf("failed to set text size: %w", err)
	}

	// Write the actual text data
	n, err := e.WriteRaw([]byte(data))
	if err != nil {
		return bytesWritten, fmt.Errorf("failed to write text data: %w", err)
	}
	bytesWritten += n

	return bytesWritten, nil
}

// WriteGBK writes a string to the printer using GBK encoding
func (e *Escpos) WriteGBK(data string) (int, error) {
	gbk, err := iconv.ConvertString(data, iconv.GBK, iconv.UTF8)
	if err != nil {
		return 0, fmt.Errorf("failed to convert to GBK encoding: %w", err)
	}
	return e.Write(gbk)
}

// WriteWEU writes a string to the printer using Western European encoding
func (e *Escpos) WriteWEU(data string) (int, error) {
	weu, err := iconv.ConvertString(data, iconv.CP850, iconv.UTF8)
	if err != nil {
		return 0, fmt.Errorf("failed to convert to Western European encoding: %w", err)
	}
	return e.Write(weu)
}

// Bold sets the printer to print bold text
func (e *Escpos) Bold(p bool) *Escpos {
	e.Style.Bold = p
	return e
}

// Underline sets the underline style with thickness p (0-2 dots)
func (e *Escpos) Underline(p uint8) *Escpos {
	if p > 2 {
		p = 2
	}
	e.Style.Underline = p
	return e
}

// Reverse sets reverse printing (white text on black background)
func (e *Escpos) Reverse(p bool) *Escpos {
	e.Style.Reverse = p
	return e
}

// Justify sets text justification (alignment)
// Use JustifyLeft, JustifyCenter, or JustifyRight constants
func (e *Escpos) Justify(p uint8) *Escpos {
	if p > JustifyRight {
		p = JustifyLeft
	}
	e.Style.Justify = p
	return e
}

// Rotate toggles 90° clockwise rotation
func (e *Escpos) Rotate(p bool) *Escpos {
	e.Style.Rotate = p
	return e
}

// UpsideDown toggles upside-down printing
func (e *Escpos) UpsideDown(p bool) *Escpos {
	e.Style.UpsideDown = p
	return e
}

// Size sets the font size. Width and Height should be between 1 and 5.
func (e *Escpos) Size(width uint8, height uint8) *Escpos {
	// Ensure values are between 1 and 5
	if width < 1 {
		width = 1
	} else if width > 5 {
		width = 5
	}

	if height < 1 {
		height = 1
	} else if height > 5 {
		height = 5
	}

	e.Style.Width = width
	e.Style.Height = height
	return e
}

// HRIPosition sets the position of the HRI (Human Readable Interpretation) characters
// Use the HRIPosition constants
func (e *Escpos) HRIPosition(p uint8) (int, error) {
	if p > HRIPositionBoth {
		return 0, fmt.Errorf("invalid HRI position: must be between 0-3")
	}
	return e.WriteRaw([]byte{gs, 'H', p})
}

// HRIFont sets the HRI font
// false: Font A (12x24)
// true: Font B (9x24)
func (e *Escpos) HRIFont(p bool) (int, error) {
	return e.WriteRaw([]byte{gs, 'f', boolToByte(p)})
}

// BarcodeHeight sets the height for barcodes in dots (default: 162)
func (e *Escpos) BarcodeHeight(p uint8) (int, error) {
	return e.WriteRaw([]byte{gs, 'h', p})
}

// BarcodeWidth sets the width for barcodes (2-6, default: 3)
func (e *Escpos) BarcodeWidth(p uint8) (int, error) {
	if p < 2 {
		p = 2
	}
	if p > 6 {
		p = 6
	}
	return e.WriteRaw([]byte{gs, 'w', p})
}

// UPCA prints a UPC-A barcode
// code must be 11-12 digits
func (e *Escpos) UPCA(code string) (int, error) {
	return e.Barcode(BarcodeUPCA, code)
}

// UPCE prints a UPC-E barcode
// code must be 11-12 digits
func (e *Escpos) UPCE(code string) (int, error) {
	return e.Barcode(BarcodeUPCE, code)
}

// EAN13 prints an EAN-13 barcode
// code must be 12-13 digits
func (e *Escpos) EAN13(code string) (int, error) {
	return e.Barcode(BarcodeEAN13, code)
}

// EAN8 prints an EAN-8 barcode
// code must be 7-8 digits
func (e *Escpos) EAN8(code string) (int, error) {
	return e.Barcode(BarcodeEAN8, code)
}

// CODE39 prints a CODE39 barcode
func (e *Escpos) CODE39(code string) (int, error) {
	return e.Barcode(BarcodeCode39, code)
}

// ITF prints an ITF barcode
func (e *Escpos) ITF(code string) (int, error) {
	return e.Barcode(BarcodeITF, code)
}

// CODABAR prints a CODABAR barcode
func (e *Escpos) CODABAR(code string) (int, error) {
	return e.Barcode(BarcodeCodabar, code)
}

// Barcode is a generic function to print barcodes
// barcodeType: one of the Barcode* constants
// code: the data to encode
func (e *Escpos) Barcode(barcodeType uint8, code string) (int, error) {
	// Validate barcode type
	if barcodeType > BarcodeCodabar {
		return 0, fmt.Errorf("invalid barcode type: %d", barcodeType)
	}

	// Validate code based on barcode type
	switch barcodeType {
	case BarcodeUPCA, BarcodeUPCE:
		if len(code) != 11 && len(code) != 12 {
			return 0, fmt.Errorf("UPC code should have 11 or 12 digits")
		}
		if !onlyDigits(code) {
			return 0, fmt.Errorf("UPC code can only contain digits")
		}
	case BarcodeEAN13:
		if len(code) != 12 && len(code) != 13 {
			return 0, fmt.Errorf("EAN-13 code should have 12 or 13 digits")
		}
		if !onlyDigits(code) {
			return 0, fmt.Errorf("EAN-13 code can only contain digits")
		}
	case BarcodeEAN8:
		if len(code) != 7 && len(code) != 8 {
			return 0, fmt.Errorf("EAN-8 code should have 7 or 8 digits")
		}
		if !onlyDigits(code) {
			return 0, fmt.Errorf("EAN-8 code can only contain digits")
		}
	case BarcodeITF:
		if len(code) < 2 || len(code)%2 != 0 {
			return 0, fmt.Errorf("ITF code must have an even number of digits (at least 2)")
		}
		if !onlyDigits(code) {
			return 0, fmt.Errorf("ITF code can only contain digits")
		}
	}

	byteCode := append([]byte(code), 0)
	return e.WriteRaw(append([]byte{gs, 'k', barcodeType}, byteCode...))
}

// QRCode prints a QR code
// code: the data to encode (max 7089 characters)
// model: QR code model (false for model 1, true for model 2)
// size: size in dots (1-16)
// correctionLevel: error correction level (use QRCodeErrorCorrectionLevel* constants)
func (e *Escpos) QRCode(code string, model bool, size uint8, correctionLevel uint8) (int, error) {
	if len(code) > 7089 {
		return 0, fmt.Errorf("QR code data too long (max 7089 characters)")
	}

	// Validate and adjust parameters
	if size < 1 {
		size = 1
	} else if size > 16 {
		size = 16
	}

	if correctionLevel < QRCodeErrorCorrectionLevelL || correctionLevel > QRCodeErrorCorrectionLevelH {
		correctionLevel = QRCodeErrorCorrectionLevelL
	}

	var m byte = 49 // Model 1
	var written int
	var err error

	// Set QR code model
	if model {
		m = 50 // Model 2
	}
	_, err = e.WriteRaw([]byte{gs, '(', 'k', 4, 0, 49, 65, m, 0})
	if err != nil {
		return 0, fmt.Errorf("failed to set QR code model: %w", err)
	}

	// Set QR code size
	_, err = e.WriteRaw([]byte{gs, '(', 'k', 3, 0, 49, 67, size})
	if err != nil {
		return 0, fmt.Errorf("failed to set QR code size: %w", err)
	}

	// Set QR code error correction level
	_, err = e.WriteRaw([]byte{gs, '(', 'k', 3, 0, 49, 69, correctionLevel})
	if err != nil {
		return 0, fmt.Errorf("failed to set QR code error correction level: %w", err)
	}

	// Store the data in the buffer
	var codeLength = len(code) + 3
	var pL, pH byte
	pH = byte(codeLength / 256)
	pL = byte(codeLength % 256)

	written, err = e.WriteRaw(append([]byte{gs, '(', 'k', pL, pH, 49, 80, 48}, []byte(code)...))
	if err != nil {
		return written, fmt.Errorf("failed to store QR code data: %w", err)
	}

	// Print the buffer
	_, err = e.WriteRaw([]byte{gs, '(', 'k', 3, 0, 49, 81, 48})
	if err != nil {
		return written, fmt.Errorf("failed to print QR code: %w", err)
	}

	return written, nil
}

// PrintImage prints an image to the printer
// Deprecated: Use PrintImageWithProcessing instead
func (e *Escpos) PrintImage(image image.Image) (int, error) {
	return e.PrintImageWithProcessing(image, ImageProcessThreshold, false, false)
}

// PrintImageWithProcessing prints an image to the printer using the specified processing method
// Multiple parameters are available to control the image processing:
//   - image: the image to print
//   - processMethod: the image processing method to use (ImageProcessDither or ImageProcessThreshold)
//   - highDensityVertical: if true, use high density vertical printing (only for dithered images)
//   - highDensityHorizontal: if true, use high density horizontal printing (only for dithered images)
//
// Returns the number of bytes written and any error encountered
func (e *Escpos) PrintImageWithProcessing(image image.Image, processMethod uint8, highDensityVertical bool, highDensityHorizontal bool) (int, error) {
	switch processMethod {
	case ImageProcessDither:
		data, err := printImageDither(image, true, true)
		if err != nil {
			return 0, fmt.Errorf("failed to transform dithered image: %w", err)
		}
		return e.WriteRaw(data)

	case ImageProcessThreshold:
		// Use the traditional threshold-based conversion
		xL, xH, yL, yH, data := printImage(image)
		return e.WriteRaw(append([]byte{gs, 'v', 48, 0, xL, xH, yL, yH}, data...))

	default:
		return 0, fmt.Errorf("unknown image processing method: %d", processMethod)
	}

}

// PrintNVBitImage prints a pre-stored bit image with index p and mode
// p: image index (1-based)
// mode: print mode (0-3)
func (e *Escpos) PrintNVBitImage(p uint8, mode uint8) (int, error) {
	if p == 0 {
		return 0, fmt.Errorf("NV bit image index must be at least 1")
	}
	if mode > 3 {
		return 0, fmt.Errorf("NV bit image mode must be between 0-3")
	}

	return e.WriteRaw([]byte{fs, 'd', p, mode})
}

// LineFeed sends a newline to the printer
func (e *Escpos) LineFeed() (int, error) {
	return e.Write("\n")
}

// LineFeedN prints and feeds the paper p lines
func (e *Escpos) LineFeedN(p uint8) (int, error) {
	return e.WriteRaw([]byte{esc, 'd', p})
}

// DefaultLineSpacing sets the line spacing to the default (1/6 inch)
func (e *Escpos) DefaultLineSpacing() (int, error) {
	return e.WriteRaw([]byte{esc, '2'})
}

// LineSpacing sets the line spacing to p/180 inch (ESC/POS)
func (e *Escpos) LineSpacing(p uint8) (int, error) {
	return e.WriteRaw([]byte{esc, '3', p})
}

// Initialize resets the printer to its default settings
func (e *Escpos) Initialize() (int, error) {
	return e.WriteRaw([]byte{esc, '@'})
}

// MotionUnits sets the horizontal (x) and vertical (y) motion units
// x: horizontal motion unit (25.4/x mm)
// y: vertical motion unit (25.4/y mm)
func (e *Escpos) MotionUnits(x, y uint8) (int, error) {
	return e.WriteRaw([]byte{gs, 'P', x, y})
}

// Cut feeds the paper to the cutting position and cuts it
func (e *Escpos) Cut() (int, error) {
	return e.WriteRaw([]byte{gs, 'V', 'A', 0x00})
}

// PartialCut performs a partial paper cut
func (e *Escpos) PartialCut() (int, error) {
	return e.WriteRaw([]byte{gs, 'V', 'B', 0x00})
}

// OpenDrawer opens the cash drawer connected to the printer
// pin: pin number (0 or 1)
// time: pulse duration (1-8) * 100ms
func (e *Escpos) OpenDrawer(pin uint8, time uint8) (int, error) {
	if pin > 1 {
		pin = 0
	}
	if time < 1 {
		time = 1
	} else if time > 8 {
		time = 8
	}
	return e.WriteRaw([]byte{esc, 'p', pin, time, time})
}

// SelectCharacterCodeTable selects the character code table
// table: code table number (0-255)
func (e *Escpos) SelectCharacterCodeTable(table uint8) (int, error) {
	return e.WriteRaw([]byte{esc, 't', table})
}

// boolToByte converts a boolean to a byte (0x00 or 0x01)
func boolToByte(b bool) byte {
	if b {
		return 0x01
	}
	return 0x00
}

// onlyDigits checks if a string contains only digits
func onlyDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
