package escpos

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// Style defines the text formatting options for the printer
type Style struct {
	Bold          bool
	Width, Height uint8
	Reverse       bool
	Underline     uint8 // can be 0, 1 or 2
	UpsideDown    bool
	Rotate        bool
	Justify       Justify
}

type Justify uint8

// Justification constants
const (
	JustifyLeft   Justify = 0
	JustifyCenter Justify = 1
	JustifyRight  Justify = 2
)

// Underline constants
const (
	UnderlineNone   uint8 = 0 // No underline
	UnderlineSingle uint8 = 1 // Single underline
	UnderlineDouble uint8 = 2 // Double underline
)

// Font type constants
const (
	FontA uint8 = 0 // Font A (12x24)
	FontB uint8 = 1 // Font B (9x24)
)

// QR code error correction levels
const (
	QRCodeErrorCorrectionLevelL uint8 = 48 // 7% recovery capacity
	QRCodeErrorCorrectionLevelM uint8 = 49 // 15% recovery capacity
	QRCodeErrorCorrectionLevelQ uint8 = 50 // 25% recovery capacity
	QRCodeErrorCorrectionLevelH uint8 = 51 // 30% recovery capacity
)

// QR code model constants
const (
	QRCodeModel1 uint8 = 49 // Model 1 (older, smaller capacity)
	QRCodeModel2 uint8 = 50 // Model 2 (newer, enhanced functionality)
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

// Real-time status command constants
const (
	// Real-time status commands
	RT_STATUS_ONLINE byte = 1
	RT_STATUS_PAPER  byte = 4

	// Masks
	RT_MASK_NEAREND byte = 0x0C // bits 2 and 3
	RT_MASK_NOPAPER byte = 0x60 // bits 5 and 6
	RT_MASK_OFFLINE byte = 0x08
)

// ESC/POS command bytes
const (
	esc byte = 0x1B
	gs  byte = 0x1D
	fs  byte = 0x1C
	dle byte = 0x10 // Data Link Escape - used for real-time commands
)

// Image processing method constants
const (
	// ImageProcessDither applies Floyd-Steinberg dithering
	ImageProcessDither uint8 = 0
	// ImageProcessThreshold applies simple threshold-based conversion
	ImageProcessThreshold uint8 = 1
)

// Code page constants
const (
	CodePagePC437      uint8 = 0  // USA, Standard Europe
	CodePageKatakana   uint8 = 1  // Katakana
	CodePagePC850      uint8 = 2  // Multilingual
	CodePagePC860      uint8 = 3  // Portuguese
	CodePagePC863      uint8 = 4  // Canadian-French
	CodePagePC865      uint8 = 5  // Nordic
	CodePageISO8859_1  uint8 = 6  // Western European
	CodePageWPC1252    uint8 = 16 // Latin 1
	CodePagePC866      uint8 = 17 // Cyrillic #2
	CodePagePC852      uint8 = 18 // Latin 2
	CodePagePC858      uint8 = 19 // Euro
	CodePageIranII     uint8 = 20 // Iran II
	CodePageLatvian    uint8 = 21 // Latvian
	CodePageISO88596   uint8 = 22 // Arabic
	CodePageLCDTurkish uint8 = 24 // Turkish
	CodePageISO8859_15 uint8 = 25 // Latin 9
	CodePageCP1098     uint8 = 38 // Farsi
	CodePageCP864      uint8 = 40 // Arabic
	CodePageISO8859_2  uint8 = 41 // Latin 2
	CodePageCP1125     uint8 = 42 // Ukrainian
	CodePageCP1250     uint8 = 47 // Latin 2
	CodePageCP1251     uint8 = 48 // Cyrillic
	CodePageCP1253     uint8 = 49 // Greek
	CodePageCP1254     uint8 = 50 // Turkish
	CodePageCP1255     uint8 = 51 // Hebrew
	CodePageCP1256     uint8 = 52 // Arabic
	CodePageCP1257     uint8 = 53 // Baltic
	CodePageCP1258     uint8 = 54 // Vietnamese
	CodePageKZ1048     uint8 = 55 // Kazakhstan
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
	reader io.Reader // Added reader for status queries
	Style  Style
	config PrinterConfig
}

// New creates a new Escpos printer instance
func New(printer Printer) *Escpos {
	return &Escpos{
		dst:    bufio.NewWriter(printer),
		reader: printer,
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
	return e.WriteRaw([]byte(data))
}

// WriteGBK writes a string to the printer using GBK encoding
func (e *Escpos) WriteGBK(data string) (int, error) {
	return e.WriteWithEncoding(data, simplifiedchinese.GBK, CodePagePC437)
}

// WriteWEU writes a string to the printer using Western European encoding
func (e *Escpos) WriteWEU(data string) (int, error) {
	return e.WriteWithEncoding(data, charmap.CodePage850, CodePagePC850)
}

// WriteWithEncoding writes text after converting it from UTF-8 to the specified encoding
// and setting the appropriate code page on the printer
func (e *Escpos) WriteWithEncoding(data string, enc encoding.Encoding, codepage uint8) (int, error) {
	return e.WriteRawWithEncoding([]byte(data), enc)
}

// WriteRawWithEncoding writes raw bytes to the printer after converting them from UTF-8
// to the specified encoding
func (e *Escpos) WriteRawWithEncoding(data []byte, enc encoding.Encoding) (int, error) {
	// Create an encoder for the target encoding
	encoder := enc.NewEncoder()

	// The input data is already in UTF-8, no need to decode first
	// Just encode directly from UTF-8 to the target encoding
	encBytes, err := encoder.Bytes(data)
	if err != nil {
		// Handle unsupported characters
		encBytes, err = encoding.ReplaceUnsupported(encoder).Bytes(data)
		if err != nil {
			return 0, fmt.Errorf("failed to encode data: %w", err)
		}
	}

	// Write the converted text
	return e.WriteRaw(encBytes)
}

// SetSize sets the font size by specifying both height and width (1-8)
// The function applies the corresponding byte value based on the formula:
// c = (2 << 3) * (width - 1) + (height - 1)
func (e *Escpos) SetSize(height, width uint8) (int, error) {
	// Ensure values are between 1 and 8
	if width < 1 {
		width = 1
	} else if width > 8 {
		width = 8
	}

	if height < 1 {
		height = 1
	} else if height > 8 {
		height = 8
	}

	sizeByte := (2<<3)*(width-1) + (height - 1)

	// Update the style
	e.Style.Height = height
	e.Style.Width = width

	// Send the command to the printer
	return e.WriteRaw([]byte{gs, '!', sizeByte})
}

// SetJustify sets the justification for text
// Use JustifyLeft, JustifyCenter, or JustifyRight constants
func (e *Escpos) SetJustify(j Justify) (int, error) {
	if e.config.DisableJustify {
		return 0, fmt.Errorf("justification is disabled in the printer configuration")
	}
	if j > JustifyRight {
		j = JustifyLeft
	}
	// Update the style
	e.Style.Justify = j

	return e.WriteRaw([]byte{esc, 'a', byte(j)})
}

// SetBold sets the bold mode
// Use true for bold, false for normal
func (e *Escpos) SetBold(b bool) (int, error) {
	if e.config.DisableBold {
		return 0, fmt.Errorf("bold mode is disabled in the printer configuration")
	}
	return e.WriteRaw([]byte{esc, 'E', boolToByte(b)})
}

// SetUnderline sets the underline mode
// Use 0 for no underline, 1 for single underline, and 2 for double underline
func (e *Escpos) SetUnderline(u uint8) (int, error) {
	if e.config.DisableUnderline {
		return 0, fmt.Errorf("underline mode is disabled in the printer configuration")
	}
	if u > 2 {
		u = 0
	}
	return e.WriteRaw([]byte{esc, '-', u})
}

// SetUpsideDown sets the upside-down mode
// Use true for upside-down, false for normal
func (e *Escpos) SetUpsideDown(u bool) (int, error) {
	if e.config.DisableUpsideDown {
		return 0, fmt.Errorf("upside-down mode is disabled in the printer configuration")
	}
	return e.WriteRaw([]byte{esc, '{', boolToByte(u)})
}

// SetRotate sets the 90° clockwise rotation
// Use true for rotated, false for normal
func (e *Escpos) SetRotate(r bool) (int, error) {
	if e.config.DisableRotate {
		return 0, fmt.Errorf("rotation mode is disabled in the printer configuration")
	}
	return e.WriteRaw([]byte{esc, 'V', boolToByte(r)})
}

// SetReverse sets the reverse printing mode
// Use true for reverse, false for normal
func (e *Escpos) SetReverse(r bool) (int, error) {
	if e.config.DisableReverse {
		return 0, fmt.Errorf("reverse mode is disabled in the printer configuration")
	}
	return e.WriteRaw([]byte{gs, 'B', boolToByte(r)})
}

// SetFont sets the font type
// Use FontA (12x24) or FontB (9x24)
func (e *Escpos) SetFont(f uint8) (int, error) {
	if f > FontB {
		f = FontA
	}
	return e.WriteRaw([]byte{esc, 'M', f})
}

// SetHRIPosition sets the position of the HRI (Human Readable Interpretation) characters
// Use the HRIPosition constants
func (e *Escpos) SetHRIPosition(p uint8) (int, error) {
	if p > HRIPositionBoth {
		return 0, fmt.Errorf("invalid HRI position: must be between 0-3")
	}
	return e.WriteRaw([]byte{gs, 'H', p})
}

// SetHRIFont sets the HRI font
// false: Font A (12x24)
// true: Font B (9x24)
func (e *Escpos) SetHRIFont(p bool) (int, error) {
	return e.WriteRaw([]byte{gs, 'f', boolToByte(p)})
}

// SetBarcodeHeight sets the height for barcodes in dots (default: 162)
func (e *Escpos) SetBarcodeHeight(p uint8) (int, error) {
	return e.WriteRaw([]byte{gs, 'h', p})
}

// SetBarcodeWidth sets the width for barcodes (2-6, default: 3)
func (e *Escpos) SetBarcodeWidth(p uint8) (int, error) {
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
//
// Parameters:
//   - code: the data to encode (max 7089 characters for Model 2, 1167 for Model 1)
//   - model: QR code model to use (QRCodeModel1 or QRCodeModel2)
//   - size: size of QR code modules in dots (1-16)
//   - correctionLevel: error correction level with these options:
//   - QRCodeErrorCorrectionLevelL: Recovers 7% of data
//   - QRCodeErrorCorrectionLevelM: Recovers 15% of data
//   - QRCodeErrorCorrectionLevelQ: Recovers 25% of data
//   - QRCodeErrorCorrectionLevelH: Recovers 30% of data
//
// Returns the number of bytes written and any error encountered.
// Use Model 2 for most applications as it offers better capacity and features.
func (e *Escpos) QRCode(code string, model uint8, size uint8, correctionLevel uint8) (int, error) {
	// Check model capacity limits
	maxLength := 7089 // Default for Model 2
	if model == QRCodeModel1 {
		maxLength = 1167
	}

	if len(code) > maxLength {
		return 0, fmt.Errorf("QR code data too long (max %d characters for the selected model)", maxLength)
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

	// Validate model parameter
	if model != QRCodeModel1 && model != QRCodeModel2 {
		model = QRCodeModel2 // Default to Model 2 if invalid
	}

	var written int
	var err error

	// Set QR code model
	_, err = e.WriteRaw([]byte{gs, '(', 'k', 4, 0, 49, 65, model, 0})
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
		data, err := PrepareImageForPrinting(image, highDensityVertical, highDensityHorizontal)
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

// SetDefaultLineSpacing sets the line spacing to the default (1/6 inch)
func (e *Escpos) SetDefaultLineSpacing() (int, error) {
	return e.WriteRaw([]byte{esc, '2'})
}

// SetLineSpacing sets the line spacing to p/180 inch (ESC/POS)
func (e *Escpos) SetLineSpacing(p uint8) (int, error) {
	return e.WriteRaw([]byte{esc, '3', p})
}

// Initialize resets the printer to its default settings
func (e *Escpos) Initialize() (int, error) {
	return e.WriteRaw([]byte{esc, '@'})
}

// SetMotionUnits sets the horizontal (x) and vertical (y) motion units
// x: horizontal motion unit (25.4/x mm)
// y: vertical motion unit (25.4/y mm)
func (e *Escpos) SetMotionUnits(x, y uint8) (int, error) {
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

// SetCodePage sets the code page (character set) for the printer
// The list of available code pages varies by printer model
func (e *Escpos) SetCodePage(codepage uint8) (int, error) {
	return e.WriteRaw([]byte{esc, 't', codepage})
}

// QueryStatus sends a real-time status request to the printer and returns the response
// The parameter 'statusType' should be one of the RT_STATUS_* constants
func (e *Escpos) QueryStatus(statusType byte) ([]byte, error) {
	// Send the real-time status request
	_, err := e.WriteRaw([]byte{dle, 0x04, statusType})
	if err != nil {
		return nil, fmt.Errorf("failed to send status request: %w", err)
	}

	// Flush the buffer to ensure the command is sent immediately
	err = e.dst.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush status request: %w", err)
	}

	// Give the printer some time to respond
	time.Sleep(100 * time.Millisecond)

	// Read the response
	if e.reader == nil {
		return nil, fmt.Errorf("reader not available")
	}

	buf := make([]byte, 1)
	n, err := e.reader.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read status response: %w", err)
	}

	if n == 0 {
		return []byte{}, nil
	}

	return buf, nil
}

// IsOnline queries the online status of the printer
// Returns true if the printer is online, false otherwise
func (e *Escpos) IsOnline() (bool, error) {
	status, err := e.QueryStatus(RT_STATUS_ONLINE)
	if err != nil {
		return false, err
	}

	if len(status) == 0 {
		return false, nil // Assume offline if no response
	}

	b := status[0]

	// If bit 3 is set, printer is offline
	isOffline := b&RT_MASK_OFFLINE == RT_MASK_OFFLINE

	return !isOffline, nil // Return true if online
}

// PaperStatus queries the paper status using DLE EOT 4
// Returns:
// 2 = Paper is adequate
// 1 = Paper is low (near end)
// 0 = No paper
func (e *Escpos) PaperStatus() (int, error) {
	status, err := e.QueryStatus(RT_STATUS_PAPER)
	if err != nil {
		return 2, err // Assume paper is OK if error
	}

	if len(status) == 0 {
		return 2, nil // Assume paper is OK if no response
	}

	b := status[0]

	// Check end sensor first (bits 5-6)
	if b&RT_MASK_NOPAPER == RT_MASK_NOPAPER {
		return 0, nil // No paper
	}

	// Then check near-end sensor (bits 2-3)
	if b&RT_MASK_NEAREND == RT_MASK_NEAREND {
		return 1, nil // Low paper
	}

	return 2, nil // Paper is adequate
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
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
