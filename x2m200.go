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
	"bufio"
	"errors"
	"io"
)

type x2m200Frame struct {
	w io.Writer
	r *bufio.Reader
	c io.Closer
	// frameBuffer []byte
}

type protocolError byte

const (
	notReconsied protocolError = 0x01
	crcFailed    protocolError = 0x02
	invaidAppID  protocolError = 0x03
)

func (x *x2m200Frame) Close() error {
	return x.c.Close()
}

func (x *x2m200Frame) Write(p []byte) (n int, err error) {

	p = append(p[:0], append([]byte{startByte}, p[0:]...)...)
	crc := checksum(&p)
	// not quite correct but works most of the time but need to ignor endByte that are not at end.
	for k := 0; k < len(p); k++ {
		if p[k] == endByte {
			p = append(p[:k], append([]byte{escByte}, p[k:]...)...)
			k++
		}
	}
	p = append(p, crc)
	p = append(p, endByte)
	return x.w.Write(p)
}

// Flow Control bytes
// startByte + [data] + CRC + endByte
const (
	startByte = 0x7D
	endByte   = 0x7E
	escByte   = 0x7F
	errorByte = 0x20
)

func (x *x2m200Frame) Read(b []byte) (n int, err error) {
	var s []byte
	header, err := x.r.Peek(1)
	if err != nil {
		return 0, io.EOF
	}
	if header[0] == startByte {
		s, err = x.r.ReadBytes(endByte)
		if err != nil {
			return 0, err
		}
		ok, p, verr := validator(s)
		// log.Println(verr)
		if ok && verr != nil {
			// log.Println(verr)
			return 0, verr
		}

		for !ok {
			// scan to next endByte
			s2, err := x.r.ReadBytes(endByte)
			if err != nil && verr != errPacketNotLongEnough {
				return 0, verr
			}
			s = append(s, s2...)
			ok, p, _ = validator(s)
		}
		if ok {
			n = copy(b, p)
			return n, nil
		}
	}
	return 0, errPacketNoStartByte
	// return 0, nil
}

func validator(b []byte) (bool, []byte, error) {
	// var k []byte
	// k = append(k, b...)

	const (
		wait = iota
		startAfterStart
		inMessage
		afterEsc
		done
	)
	var buf []byte
	state := wait
	for k, v := range b {
		switch state {
		case wait:
			if v == startByte {
				state = inMessage
				buf = append(buf, v)
			}
		case inMessage:
			if k == len(b)-1 {
				state = done
			} else if v == escByte {
				state = afterEsc
			} else {
				state = inMessage
				buf = append(buf, v)
			}
		case afterEsc:
			if v == endByte && k == len(b)-1 {
				return false, nil, errPacketNotLongEnough
			}
			state = inMessage
			buf = append(buf, v)
		}
	}
	var crcByte byte
	n := len(buf)
	crcByte, buf = buf[n-1], buf[:n-1]
	n--

	crc := checksum(&buf)

	if crcByte != crc {
		return false, nil, errPacketBadCRC
	}

	buf = buf[:0+copy(buf[0:], buf[1:])]
	n--

	if n > 1 {
		if buf[0] == errorByte {
			pError := protocolError(buf[1])
			switch pError {
			case notReconsied:
				return true, buf, errProtocolErrorNotReconsied
			case crcFailed:
				return true, buf, errProtocolErrorCRCfailed
			case invaidAppID:
				return true, buf, errProtocolErrorInvaidAppID
			}
		}
	}
	// log.Println("returning nil")
	// log.Printf("%#0x\n", buf)
	return true, buf, nil
}

var (
	errPacketNotLongEnough       = errors.New("not long enough")
	errPacketNoStartByte         = errors.New("no startbyte")
	errPacketBadCRC              = errors.New("failed checksum")
	errProtocolErrorNotReconsied = errors.New("protocol error command not recognised")
	errProtocolErrorCRCfailed    = errors.New("protocol error command bad crc")
	errProtocolErrorInvaidAppID  = errors.New("protocol error invalid app id")
)

// Calculated by XOR’ing all bytes from <START> + [Data].
// Note that the CRC is done after escape bytes is removed. This
// means that CRC is also calculated before adding escape bytes.
func checksum(p *[]byte) byte {
	var crc byte
	for _, b := range *p {
		crc = crc ^ b
	}
	return crc
}

var errChecksumInvalidPacketSTART = errors.New("invalid packet missing start")
