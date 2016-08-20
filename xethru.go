// xethru  Copyright (C) 2016
// This work is copyright no part may be reproduced by any process,
// nor may any other exclusive right be exercised, without the permission of
// NeuralSpaz aka Josh Gardiner 2016
// It is my intent that this will be released as open source at some
// future time. If you would like to contribute please contact me.

package xethru

import (
	"errors"
	"io"
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
	// cant be error from checksum at we just set the startByte
	crc, _ := checksum(&p)
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
	r io.Reader
	// err    error
	// toRead []byte
}

func NewXethruReader(r io.Reader) io.Reader {
	return &xethruFrameReader{r}
	// r:      r,
	// err:    nil,
	// toRead: make([]byte, 128),
	// }
}

func (x *xethruFrameReader) Read(b []byte) (n int, err error) {
	// read from the reader
	n, err = x.r.Read(b)
	if n > 0 {
		var last byte
		// pop endByte
		last, b = b[n-1], b[:n-1]
		n--
		if last != endByte {
			return 0, ErrorPacketNotEndbyte
		}
		// delete escBytes
		for i := 0; i < n; i++ {
			if b[i] == escByte {
				b = b[:i+copy(b[i:], b[i+1:])]
				n--
			}
		}
		var crcByte byte
		// pop crcbyte
		crcByte, b = b[len(b)-1], b[:len(b)-1]
		n--
		crc, err := checksum(&b)
		if err != nil {
			return 0, ErrorPacketNoStartByte
		}
		if crcByte != crc {
			return 0, ErrorPacketBadCRC
		}
		// delete startByte
		b = b[:0+copy(b[0:], b[1:])]
		n--
		if n == 0 {
			return n, io.EOF
		}
		return n, nil
	}
	if err != nil {
		return 0, err
	}
	return 0, io.EOF
}

var (
	ErrorPacketNoStartByte = errors.New("no startbyte")
	ErrorPacketNotEndbyte  = errors.New("does not end with endbyte")
	ErrorPacketBadCRC      = errors.New("failed checksum")
)

// Calculated by XOR’ing all bytes from <START> + [Data].
// Note that the CRC is done after escape bytes is removed. This
// means that CRC is also calculated before adding escape bytes.
func checksum(p *[]byte) (byte, error) {
	// fmt.Printf("byte to check sum %x\n", p)
	if (*p)[0] != startByte {
		return 0x00, errChecksumInvalidPacketSTART
	}
	var crc byte
	for _, b := range *p {
		crc = crc ^ b
	}

	return crc, nil
}

var errChecksumInvalidPacketSTART = errors.New("invalid packet missing start")
