package escpos

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockPrinter implements the Printer interface for testing
type MockPrinter struct {
	buf    bytes.Buffer
	status []byte
}

func (m *MockPrinter) Close() error {
	return nil
}

func (m *MockPrinter) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func (m *MockPrinter) Read(p []byte) (n int, err error) {
	if len(m.status) > 0 {
		n = copy(p, m.status)
		return n, nil
	}
	return 0, nil
}

func (m *MockPrinter) Bytes() []byte {
	return m.buf.Bytes()
}

func (m *MockPrinter) SetStatus(status []byte) {
	m.status = status
}

func NewMockPrinter() *MockPrinter {
	return &MockPrinter{}
}

// TestNew tests creating a new Escpos instance
func TestNew(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	assert.NotNil(t, p)
	assert.NotNil(t, p.dst)
	assert.NotNil(t, p.reader)
}

// TestWriteRaw tests writing raw bytes to the printer
func TestWriteRaw(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	data := []byte{0x1B, 0x40} // ESC @
	n, err := p.WriteRaw(data)

	assert.NoError(t, err)
	assert.Equal(t, 2, n)

	// Flush to ensure data is written
	err = p.Print()
	assert.NoError(t, err)

	// Check the written data
	assert.Equal(t, data, mock.Bytes())
}

// TestWrite tests writing a string to the printer
func TestWrite(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	text := "Hello, Printer!"
	n, err := p.Write(text)

	assert.NoError(t, err)
	assert.Equal(t, len(text), n)

	err = p.Print()
	assert.NoError(t, err)

	assert.Equal(t, []byte(text), mock.Bytes())
}

// TestPrint tests flushing data to the printer
func TestPrint(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.Write("Test")
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	assert.Equal(t, []byte("Test"), mock.Bytes())
}

// TestPrintAndCut tests printing and cutting
func TestPrintAndCut(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.Write("Test")
	assert.NoError(t, err)

	err = p.PrintAndCut()
	assert.NoError(t, err)

	// Should contain both the text and the cut command
	expected := append([]byte("Test"), []byte{gs, 'V', 'A', 0x00}...)
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetSize tests setting the font size
func TestSetSize(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	// Test with height=2, width=2
	_, err := p.SetSize(2, 2)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	// The byte value for 2x2 should be: (2<<3)*(2-1) + (2-1) = 16 + 1 = 17
	expected := []byte{gs, '!', 17}
	assert.Equal(t, expected, mock.Bytes())

	// Test with invalid values (should be clamped)
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.SetSize(0, 9)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	// Should be clamped to height=1, width=8
	// This results in (2<<3)*(8-1) + (1-1) = 112 + 0 = 112
	expected = []byte{gs, '!', 112}
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetJustify tests setting text justification
func TestSetJustify(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	// Test center justification
	_, err := p.SetJustify(JustifyCenter)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, 'a', 1}
	assert.Equal(t, expected, mock.Bytes())

	// Test with disabled justify
	mock = NewMockPrinter()
	p = New(mock)
	p.SetConfig(PrinterConfig{DisableJustify: true})

	_, err = p.SetJustify(JustifyRight)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "justification is disabled")
}

// TestSetBold tests setting bold mode
func TestSetBold(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	// Test bold on
	_, err := p.SetBold(true)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, 'E', 1}
	assert.Equal(t, expected, mock.Bytes())

	// Test bold off
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.SetBold(false)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected = []byte{esc, 'E', 0}
	assert.Equal(t, expected, mock.Bytes())

	// Test with disabled bold
	mock = NewMockPrinter()
	p = New(mock)
	p.SetConfig(PrinterConfig{DisableBold: true})

	_, err = p.SetBold(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bold mode is disabled")
}

// TestSetUnderline tests setting underline mode
func TestSetUnderline(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	// Test underline single
	_, err := p.SetUnderline(1)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, '-', 1}
	assert.Equal(t, expected, mock.Bytes())

	// Test with invalid value (should be clamped)
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.SetUnderline(3)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected = []byte{esc, '-', 0} // Clamped to 0
	assert.Equal(t, expected, mock.Bytes())

	// Test with disabled underline
	mock = NewMockPrinter()
	p = New(mock)
	p.SetConfig(PrinterConfig{DisableUnderline: true})

	_, err = p.SetUnderline(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "underline mode is disabled")
}

// TestBarcode tests printing barcodes
func TestBarcode(t *testing.T) {
	// Test valid EAN13
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.EAN13("1234567890128")
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := append([]byte{gs, 'k', BarcodeEAN13}, append([]byte("1234567890128"), 0)...)
	assert.Equal(t, expected, mock.Bytes())

	// Test invalid EAN13 (not digits)
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.EAN13("12345X7890128")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only contain digits")

	// Test invalid EAN13 (wrong length)
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.EAN13("123456789")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "should have 12 or 13 digits")
}

// TestQRCode tests printing QR codes
func TestQRCode(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.QRCode("https://example.com", QRCodeModel2, 5, QRCodeErrorCorrectionLevelM)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	// Check that the commands were written in the correct sequence
	output := mock.Bytes()

	// Should contain the model select command
	modelCmd := []byte{gs, '(', 'k', 4, 0, 49, 65, QRCodeModel2, 0}
	assert.Contains(t, string(output), string(modelCmd))

	// Should contain the size command
	sizeCmd := []byte{gs, '(', 'k', 3, 0, 49, 67, 5}
	assert.Contains(t, string(output), string(sizeCmd))

	// Should contain the error correction command
	errCmd := []byte{gs, '(', 'k', 3, 0, 49, 69, QRCodeErrorCorrectionLevelM}
	assert.Contains(t, string(output), string(errCmd))

	// Should contain the data command
	dataCmd := append([]byte{gs, '(', 'k'}, []byte{byte(len("https://example.com") + 3), 0, 49, 80, 48}...)
	dataCmd = append(dataCmd, []byte("https://example.com")...)
	assert.Contains(t, string(output), string(dataCmd))

	// Should contain the print command
	printCmd := []byte{gs, '(', 'k', 3, 0, 49, 81, 48}
	assert.Contains(t, string(output), string(printCmd))

	// Test with invalid QR code model
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.QRCode("test", 48, 5, QRCodeErrorCorrectionLevelM) // 48 is invalid
	assert.NoError(t, err)                                        // Should default to Model 2

	err = p.Print()
	assert.NoError(t, err)

	// Verify it defaulted to Model 2
	output = mock.Bytes()
	modelCmd = []byte{gs, '(', 'k', 4, 0, 49, 65, QRCodeModel2, 0}
	assert.Contains(t, string(output), string(modelCmd))
}

// TestCut tests cutting the paper
func TestCut(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.Cut()
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{gs, 'V', 'A', 0x00}
	assert.Equal(t, expected, mock.Bytes())
}

// TestPartialCut tests partial cutting the paper
func TestPartialCut(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.PartialCut()
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{gs, 'V', 'B', 0x00}
	assert.Equal(t, expected, mock.Bytes())
}

// TestOpenDrawer tests opening the cash drawer
func TestOpenDrawer(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.OpenDrawer(0, 2)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, 'p', 0, 2, 2}
	assert.Equal(t, expected, mock.Bytes())

	// Test with invalid pin (should be clamped)
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.OpenDrawer(2, 2) // Pin 2 is invalid
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected = []byte{esc, 'p', 0, 2, 2} // Defaults to pin 0
	assert.Equal(t, expected, mock.Bytes())

	// Test with invalid time (should be clamped)
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.OpenDrawer(1, 10) // Time 10 is too high
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected = []byte{esc, 'p', 1, 8, 8} // Clamped to 8
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetSelectCodePage tests setting the code page
func TestSetSelectCodePage(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetSelectCodePage(CodePagePC850)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, 't', CodePagePC850}
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetLineFeed tests line feeds
func TestSetLineFeed(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetLineFeed()
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte("\n")
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetLineFeedN tests multiple line feeds
func TestSetLineFeedN(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetLineFeedN(3)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, 'd', 3}
	assert.Equal(t, expected, mock.Bytes())
}

// TestInitialize tests initializing the printer
func TestInitialize(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.Initialize()
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, '@'}
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetHRIPosition tests setting the HRI position
func TestSetHRIPosition(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetHRIPosition(HRIPositionBelow)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{gs, 'H', HRIPositionBelow}
	assert.Equal(t, expected, mock.Bytes())

	// Test invalid HRI position (should be clamped)
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.SetHRIPosition(5) // Invalid position
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid HRI position")
}

// TestSetHRIFont tests setting the HRI font
func TestSetHRIFont(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetHRIFont(true)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{gs, 'f', 1}
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetBarcodeHeight tests setting the barcode height
func TestSetBarcodeHeight(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetBarcodeHeight(100)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{gs, 'h', 100}
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetBarcodeWidth tests setting the barcode width
func TestSetBarcodeWidth(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetBarcodeWidth(4)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{gs, 'w', 4}
	assert.Equal(t, expected, mock.Bytes())

	// Test invalid width (should be clamped)
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.SetBarcodeWidth(1) // Too small
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected = []byte{gs, 'w', 2} // Clamped to 2
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetDefaultLineSpacing tests setting default line spacing
func TestSetDefaultLineSpacing(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetDefaultLineSpacing()
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, '2'}
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetLineSpacing tests setting custom line spacing
func TestSetLineSpacing(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetLineSpacing(30)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{esc, '3', 30}
	assert.Equal(t, expected, mock.Bytes())
}

// TestSetMotionUnits tests setting motion units
func TestSetMotionUnits(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	_, err := p.SetMotionUnits(10, 20)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	expected := []byte{gs, 'P', 10, 20}
	assert.Equal(t, expected, mock.Bytes())
}

// Helper function to create a simple test image
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Draw a diagonal line
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x == y || x == height-y {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}
	return img
}

// TestPrintImageWithProcessing tests printing images
func TestPrintImageWithProcessing(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	img := createTestImage(64, 64)

	_, err := p.PrintImageWithProcessing(img, ImageProcessDither, true, true)
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	// Simply check the output contains data (actual format checking would be too complex)
	assert.Greater(t, len(mock.Bytes()), 10)
}

func TestQueryStatus(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	// Test valid status response
	mock.SetStatus([]byte{0x08}) // Example status byte
	status, err := p.QueryStatus(RT_STATUS_ONLINE)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x08}, status)

	// Test no response
	mock.SetStatus([]byte{}) // No status byte
	status, err = p.QueryStatus(RT_STATUS_ONLINE)
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, status)
}

func TestIsOnline(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	// Test when printer is online
	mock.SetStatus([]byte{0x00}) // Bit 3 (offline) is not set
	online, err := p.IsOnline()
	assert.NoError(t, err)
	assert.True(t, online)

	// Test when printer is offline
	mock.SetStatus([]byte{0x08}) // Bit 3 (offline) is set
	online, err = p.IsOnline()
	assert.NoError(t, err)
	assert.False(t, online)

	// Test no response
	mock.SetStatus([]byte{}) // No status byte
	online, err = p.IsOnline()
	assert.NoError(t, err)
	assert.False(t, online)
}

func TestPaperStatus(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	// Test paper adequate
	mock.SetStatus([]byte{0x00}) // No paper-related bits set
	status, err := p.PaperStatus()
	assert.NoError(t, err)
	assert.Equal(t, 2, status)

	// Test paper low (near end)
	mock.SetStatus([]byte{0x0C}) // Bits 2 and 3 (near end) are set
	status, err = p.PaperStatus()
	assert.NoError(t, err)
	assert.Equal(t, 1, status)

	// Test no paper
	mock.SetStatus([]byte{0x60}) // Bits 5 and 6 (no paper) are set
	status, err = p.PaperStatus()
	assert.NoError(t, err)
	assert.Equal(t, 0, status)

	// Test no response
	mock.SetStatus([]byte{}) // No status byte
	status, err = p.PaperStatus()
	assert.NoError(t, err)
	assert.Equal(t, 2, status) // Assume paper is adequate
}

// TestWriteWithEncoding tests writing with different encodings
func TestWriteWithEncoding(t *testing.T) {
	mock := NewMockPrinter()
	p := New(mock)

	// This is just a basic test since we can't test the actual encoding without mocking
	// the encoding functions themselves
	_, err := p.WriteGBK("测试")
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	// Check that something was written (actual content would depend on encoding)
	assert.Greater(t, len(mock.Bytes()), 0)

	// Test WriteWEU
	mock = NewMockPrinter()
	p = New(mock)

	_, err = p.WriteWEU("áéíóú")
	assert.NoError(t, err)

	err = p.Print()
	assert.NoError(t, err)

	assert.Greater(t, len(mock.Bytes()), 0)
}

// TestUtilityFunctions tests the utility functions
func TestUtilityFunctions(t *testing.T) {
	// Test boolToByte
	assert.Equal(t, byte(0x01), boolToByte(true))
	assert.Equal(t, byte(0x00), boolToByte(false))

	// Test onlyDigits
	assert.True(t, onlyDigits("1234567890"))
	assert.False(t, onlyDigits("123abc456"))
	assert.False(t, onlyDigits(""))
}
