// This file is part of Xethru-Go - A Golang library for the xethru modules
//
// The MIT License (MIT)
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
	errorByte = 0x20
)

type protocolError byte

const (
	notReconsied protocolError = 0x01
	crcFailed    protocolError = 0x02
	invaidAppID  protocolError = 0x03
)

// Open Creates a x2m200 xethu serial protocol from a io.ReadWriter
// it implements io.Reader and io.Writer
func Open(device string, port io.ReadWriter) Framer {
	if device == "x2m200" {
		return x2m200Frame{w: port, r: port}
	} else {
		return port
	}
}

// CreateSplitReadWriter Used help with testing Framer
func CreateSplitReadWriter(w io.Writer, r io.Reader) Framer {
	return x2m200Frame{w: w, r: r}
}

// Framer is a wrapper for a serial protocol. it insets the start, crc and end bytes for you
type Framer interface {
	io.Writer
	io.Reader
	// Ping(t time.Duration) (bool, error)
}

type x2m200Frame struct {
	w io.Writer
	r io.Reader
}

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
	return
}

func (x x2m200Frame) Read(b []byte) (n int, err error) {
	// read from the reader
	n, err = x.r.Read(b)
	// should be at least 3 bytes (start,crc,end)
	if n > 3 {
		var last byte
		// pop endByte
		last, b = b[n-1], b[:n-1]
		n--
		if last != endByte {
			return n, errorPacketNotEndbyte
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
		crc, err := checksum(&b)
		if err != nil {
			return n, errorPacketNoStartByte
		}
		if crcByte != crc {
			return n, errorPacketBadCRC
		}
		// delete startByte
		b = b[:0+copy(b[0:], b[1:])]
		n--
		// check for errors byte
		if b[0] == errorByte {
			switch protocolError(b[1]) {
			case notReconsied:
				return n, protocolErrorNotReconsied
			case crcFailed:
				return n, protocolErrorCRCfailed
			case invaidAppID:
				return n, protocolErrorInvaidAppID
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
	errorPacketNoStartByte    = errors.New("no startbyte")
	errorPacketNotEndbyte     = errors.New("does not end with endbyte")
	errorPacketBadCRC         = errors.New("failed checksum")
	protocolErrorNotReconsied = errors.New("protocol error command not reconsied")
	protocolErrorCRCfailed    = errors.New("protocol error command bad crc")
	protocolErrorInvaidAppID  = errors.New("protocol error invalid app id")
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
