package escpos

import (
	"fmt"
	"io"
	"math"

	"golang.org/x/text/encoding"
)

const (
	esc                               = 0x1B
	gs                                = 0x1D
	lf                                = 0x0A
	QRCodeErrorCorrectionLevelL uint8 = 48
	QRCodeErrorCorrectionLevelM uint8 = 49
	QRCodeErrorCorrectionLevelQ uint8 = 50
	QRCodeErrorCorrectionLevelH uint8 = 51
)

type Escpos struct {
	dev io.ReadWriter
	enc *encoding.Encoder
}

// Init printer cleaning your old configs
func (e *Escpos) Init() {
	e.dev.Write([]byte{esc, 0x40})
}

// Feed skip one line of paper
func (e *Escpos) Feed() {
	e.dev.Write([]byte{lf})
}

// Feed skip n lines of paper
func (e *Escpos) FeedN(n byte) {
	e.dev.Write([]byte{esc, 0x64, n})
}

// SelfTest start self test of printer
func (e *Escpos) SelfTest() {
	e.dev.Write([]byte{gs, 0x28, 0x41, 0x02, 0x00, 0x00, 0x02})
}

// Write print text
func (e *Escpos) Write(text string) error {
	str, err := encoding.ReplaceUnsupported(e.enc).String(text)
	if err != nil {
		return err
	}

	e.dev.Write([]byte(str))
	return nil
}

func (e *Escpos) Writeln(text string) {
	e.Write(text + "\n")
}

// Print QRCode
func (e *Escpos) QRCode(code string, model bool, size uint8, correctionLevel uint8) (int, error) {
	if len(code) > 7089 {
		return 0, fmt.Errorf("the code is too long, it's length should be smaller than 7090")
	}
	if size < 1 {
		size = 1
	}
	if size > 16 {
		size = 16
	}
	var m byte = 49
	var err error

	if model {
		m = 50
	}
	_, err = e.dev.Write([]byte{gs, '(', 'k', 4, 0, 49, 65, m, 0})
	if err != nil {
		return 0, err
	}

	// set the qr code size
	_, err = e.dev.Write([]byte{gs, '(', 'k', 3, 0, 49, 67, size})
	if err != nil {
		return 0, err
	}
	// set the qr code error correction level
	if correctionLevel < 48 {
		correctionLevel = 48
	}

	if correctionLevel > 51 {
		correctionLevel = 51
	}

	_, err = e.dev.Write([]byte{gs, '(', 'k', 3, 0, 49, 69, size})
	if err != nil {
		return 0, err
	}

	// store the data in the buffer
	// we now write stuff to the printer, so lets save it for returning
	// pL and pH define the size of the data. Data ranges from 1 to (pL + pH*256)-3
	// 3 < pL + pH*256 < 7093
	var codeLength = len(code) + 3
	var pL, pH byte
	pH = byte(int(math.Floor(float64(codeLength) / 256)))
	pL = byte(codeLength - 256*int(pH))

	written, err := e.dev.Write(append([]byte{gs, '(', 'k', pL, pH, 49, 80, 48}, []byte(code)...))
	if err != nil {
		return written, err
	}

	// finally print the buffer
	_, err = e.dev.Write([]byte{gs, '(', 'k', 3, 0, 49, 81, 48})
	if err != nil {
		return written, err
	}

	return written, nil
}

// New create new Escpos struct and set default enconding
func New(dev io.ReadWriter, codepage int) *Escpos {
	escpos := &Escpos{dev: dev}
	// default for russian text
	// установить енкодер
	escpos.Charset(CharsetPC866)

	// установить кодовую страницу в принтере
	escpos.SetCodePage(codepage)
	return escpos
}
