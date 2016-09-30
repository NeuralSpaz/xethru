// Copyright (c) 2016 Josh Gardiner aka NeuralSpaz on github.com
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package xethru a open source implementation driver for xethru sensor modules.
// The current state of this library is still unstable and under active development.
// Contributions are welcome.
// To use with the X2M200 module you will first need to create a
// serial io.ReadWriter (there is an examples in the example dir)
// then you can use Open to create a new x2m200 device that
// will handle all the start, end, crc and escaping for you.
package xethru

import (
	"errors"
	"io"
)

type x2m200Frame struct {
	w io.Writer
	r io.Reader
}

// Flow Control bytes
// startByte + [data] + CRC + endByte
const (
	startByte = 0x7D
	endByte   = 0x7E
	escByte   = 0x7F
	errorByte = 0x20
)

type protocolError byte

const (
	notReconsied protocolError = 0x01
	crcFailed    protocolError = 0x02
	invaidAppID  protocolError = 0x03
)

func (x x2m200Frame) Write(p []byte) (n int, err error) {

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

	// fmt.Printf("DEBUG Write: ")
	// for i := 0; i < n; i++ {
	// 	fmt.Printf("%#02x", p[i])
	// }
	// fmt.Printf("\n")
	return
}

func (x x2m200Frame) Read(b []byte) (n int, err error) {
	// read from the reader
	n, err = x.r.Read(b)
	// x.r.buffio.ReadByte
	// fmt.Printf("DEBUG Read: ")
	// for i := 0; i < n; i++ {
	// 	fmt.Printf("%#02x", b[i])
	// }
	// fmt.Printf("\n")
	// should be at least 3 bytes (start,crc,end)
	if n > 3 {
		var last byte
		// pop endByte
		last, b = b[n-1], b[:n-1]
		n--
		if last != endByte {
			return 0, errPacketNotEndbyte
		}

		// pop crcByte to check later
		var crcByte byte
		crcByte, b = b[n-1], b[:n-1]
		n--

		// delete escBytes
		for i := 0; i < (n - 1); i++ {
			if b[i] == escByte && b[i+1] == endByte {
				b = b[:i+copy(b[i:], b[i+1:])]
				n--
			}
		}
		// check crcbyte
		crc, crcerr := checksum(&b)
		if crcerr != nil {
			return 0, errPacketNoStartByte
		}
		if crcByte != crc {
			return 0, errPacketBadCRC
		}
		// delete startByte
		b = b[:0+copy(b[0:], b[1:])]
		n--
		// check for errors byte
		if b[0] == errorByte {
			switch protocolError(b[1]) {
			case notReconsied:
				return 0, errProtocolErrorNotReconsied
			case crcFailed:
				return 0, errProtocolErrorCRCfailed
			case invaidAppID:
				return 0, errProtocolErrorInvaidAppID
			}
		}
		return n, nil
	}
	if err != nil {
		return 0, err
	}
	return 0, nil
}

var (
	errPacketNoStartByte         = errors.New("no startbyte")
	errPacketNotEndbyte          = errors.New("does not end with endbyte")
	errPacketBadCRC              = errors.New("failed checksum")
	errProtocolErrorNotReconsied = errors.New("protocol error command not reconsied")
	errProtocolErrorCRCfailed    = errors.New("protocol error command bad crc")
	errProtocolErrorInvaidAppID  = errors.New("protocol error invalid app id")
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
