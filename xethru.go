package xethru

import (
	"errors"
	"io"
	"log"
)

// Flow Control bytes
// startByte + [data] + CRC + endByte
const (
	startByte = 0x7D
	endByte   = 0x7E
	escByte   = 0x7F
)

// Sensorer implements io.Writer and io.Reader
type Sensorer interface {
	// NewXethruWriter(w io.Writer) io.Writer
	// Write(p []byte) (n int, err error)
	// io.Reader
	// Read(p []byte) (n int, err error)
	// Checksum(p []byte) (crc byte, err error)
}

type xethruFrameWriter struct {
	w io.Writer
}

// NewXethruWriter builds data frame
func NewXethruWriter(w io.Writer) io.Writer {
	return &xethruFrameWriter{w}
}

func (x *xethruFrameWriter) Write(p []byte) (n int, err error) {

	p = append(p[:0], append([]byte{startByte}, p[0:]...)...)
	crc, _ := checksum(p)
	for k := 0; k < len(p); k++ {
		if p[k] == endByte {
			p = append(p[:k], append([]byte{escByte}, p[k:]...)...)
			k++
		}
	}
	p = append(p, crc)
	p = append(p, endByte)

	n, err = x.w.Write(p)
	return
}

type xethruFrameReader struct {
	r      io.Reader
	err    error
	toRead []byte
}

func NewXethruReader(r io.Reader) io.Reader {
	return &xethruFrameReader{
		r:      r,
		err:    nil,
		toRead: make([]byte, 512),
	}
}

func (x *xethruFrameReader) Read(b []byte) (n int, err error) {
	for {
		n, x.err = x.r.Read(x.toRead)
		if n > 0 {
			x.toRead, n, err = xethuProtocol(x.toRead, n)
			if err != nil {
				log.Println(err)
			}
			copy(b, x.toRead)
			if len(x.toRead) == 0 {
				return n, io.EOF
			}
			return n, nil
		}
		if x.err != nil {
			return 0, x.err
		}
	}
	return 0, io.EOF
}

func xethuProtocol(b []byte, n int) ([]byte, int, error) {
	if len(b) == 0 {
		return nil, 0, errors.New("zero size")
	}
	if b[0] != startByte {
		return nil, 0, errors.New("no startbyte")
	}
	var last byte
	last, b = b[n-1], b[:n-1]
	n--
	if last != endByte {
		return nil, 0, errors.New("does not end with endbyte")
	}
	for i := 0; i < n; i++ {
		if b[i] == escByte {
			b = b[:i+copy(b[i:], b[i+1:])]
			n--
		}
	}

	// pop off crc byte
	var crcByte byte
	crcByte, b = b[len(b)-1], b[:len(b)-1]
	n--

	crc, err := checksum(b)
	if err != nil {
		log.Println(err)
	}

	//
	if crcByte != crc {
		return nil, 0, errors.New("failed checksum")
	}

	// cut startByte off
	b = b[1:]
	n--
	return b, n, nil
}

// Calculated by XOR’ing all bytes from <START> + [Data].
// Note that the CRC is done after escape bytes is removed. This
// means that CRC is also calculated before adding escape bytes.
func checksum(p []byte) (byte, error) {
	// fmt.Printf("byte to check sum %x\n", p)
	if (p)[0] != startByte {
		return 0x00, errChecksumInvalidPacketSTART
	}
	var crc byte
	for _, b := range p {
		crc = crc ^ b
	}

	return crc, nil
}

var errChecksumInvalidPacketSTART = errors.New("invalid packet missing start")
