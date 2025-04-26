# About escpos [![GoDoc](https://godoc.org/github.com/schawnndev/escpos?status.svg)](https://godoc.org/github.com/schawnndev/escpos)
[![Go Reference](https://pkg.go.dev/badge/github.com/schawnndev/escpos.svg)](https://pkg.go.dev/github.com/schawnndev/escpos)

This is a [Golang](http://www.golang.org/project) package that provides
[ESC-POS](https://en.wikipedia.org/wiki/ESC/P) library functions to help with
sending control codes to a ESC-POS thermal printer.

It implements the protocol described in [this Command Manual](https://pos-x.com/download/escpos-programming-manual/)

## Current featureset
  * [x] Initializing the Printer
  * [x] Toggling Underline mode
  * [x] Toggling Bold text
  * [x] Toggling upside-down character printing
  * [x] Toggling Reverse mode
  * [x] Line spacing settings
  * [x] Rotated characters
  * [x] Align text
  * [x] Default ASCII Charset, Western Europe and GBK encoding
  * [x] Character size settings
  * [x] UPC-A, UPC-E, EAN13, EAN8 Barcodes
  * [x] QR Codes
  * [x] Standard printing mode
  * [x] Image Printing
  * [x] Printing of predefined NV images
  * [x] Cash drawer control

## Installation ##

Install the package via the following:

    go get -u github.com/schawnndev/escpos

## Usage ##

The escpos package can be used as the following:

```go
package main

import (
	"github.com/schawnndev/escpos"
)

func main() {
	nwPrinter, err := escpos.NewNetworkPrinter("192.168.8.40:9100")
	if err != nil {
		println(err.Error())
		return
	}
	defer nwPrinter.Close()
	
	p := escpos.New(nwPrinter)
	p.SetConfig(escpos.ConfigEpsonTMT20II)

	p.SetBold(true)
	p.SetSize(2, 2)
	p.Write("Hello World")
	p.LineFeed()
	
	p.SetBold(false)
	p.SetUnderline(escpos.UnderlineSingle)
	p.SetJustify(escpos.JustifyCenter)
	p.Write("this is underlined")
	p.LineFeed()
	p.QRCode("https://github.com/schawnndev/escpos", escpos.QRCodeModel2, 3, escpos.QRCodeErrorCorrectionLevelL)

	// You need to use either p.Print() or p.PrintAndCut() at the end to send the data to the printer.
	p.PrintAndCut()
}
```

## Setting Printer Parameters ##

The library provides a consistent naming convention for functions that set parameters, using the `Set` prefix:

```go
// Setting text parameters
p.SetBold(true)
p.SetUnderline(escpos.UnderlineDouble)
p.SetUpsideDown(true)
p.SetReverse(true)
p.SetJustify(escpos.JustifyCenter)
p.SetSize(2, 2)
p.SetFont(escpos.FontB)

// Setting line parameters
p.SetLineSpacing(30)
p.SetDefaultLineSpacing()

// Setting barcode parameters
p.SetBarcodeHeight(100)
p.SetBarcodeWidth(4)
p.SetHRIPosition(escpos.HRIPositionBelow)
p.SetHRIFont(true)

// Other control functions
p.SetMotionUnits(10, 20)
p.SetCodePage(escpos.CodePagePC850)
```

## Disable features ##

As the library sets all the styling parameters again for each call of Write, you might run into compatibility issues. Therefore it is possible to deactivate features.
To do so, use a predefined config (available for all printers listed under [Compatibility](#Compatibility)) right after the escpos.New call

```go
p := escpos.New(socket)
p.SetConfig(escpos.ConfigEpsonTMT20II) // predefined config for the Epson TM-T20II

// or for example

p.SetConfig(escpos.PrinterConfig{DisableUnderline: true})
```

## Other Printer Sources ##

If you want to use other printer sources, you can implement the `Printer` interface provided by the library.
The `Printer` interface defines the basic methods (`Read`, `Write`, and `Close`) required for communication with a printer.
This allows flexibility to connect to printers using different protocols or sources, such as serial, network, or custom implementations.

For example, to implement a serial printer connection, you can do the following:

```go
package escpos

import "go.bug.st/serial"

type serialPrinter struct {
	port serial.Port
}

func NewSerialPrinter(portName string, baudRate int) (Printer, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	return &serialPrinter{
		port: port,
	}, nil
}

func (sp *serialPrinter) Read(p []byte) (n int, err error) {
	return sp.port.Read(p)
}

func (sp *serialPrinter) Write(p []byte) (n int, err error) {
	return sp.port.Write(p)
}
```

## Compatibility ##

This is a (not complete) list of supported and tested devices.

| Manufacturer | Model    | Styling   | Barcodes | QR Codes | Images |
|--------------|----------| --------- | -------- | ------ | ------ |
| Epson        | TM-T20II | ✅        | ✅        | ✅     | ✅     |
| Epson        | TM-T88II | ☑️<br/>UpsideDown Printing not supported  | ✅        |       | ✅     |
| Munbyn       | ITPP047P | ✅  | ✅        |  ✅    | ✅     |
