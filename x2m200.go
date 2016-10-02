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
	"bufio"
	"errors"
	"io"
	"log"
)

type x2m200Frame struct {
	w io.Writer
	r *bufio.Reader
	// buf          []byte
	// rpos, wpos   int // buf read and write positions
	// err          error
	// lastByte     int
	// lastRuneSize int
}

type protocolError byte

const (
	notReconsied protocolError = 0x01
	crcFailed    protocolError = 0x02
	invaidAppID  protocolError = 0x03
)

func (x *x2m200Frame) Write(p []byte) (n int, err error) {

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

// const headerSize int = 2

// Flow Control bytes
// startByte + [data] + CRC + endByte
const (
	startByte = 0x7D
	endByte   = 0x7E
	escByte   = 0x7F
	errorByte = 0x20
)

func (x *x2m200Frame) Read(b []byte) (n int, err error) {
	// read from the reader
	// var n int
	// var p byte
	var s []byte
	// var n int
	header, err := x.r.Peek(1)
	if header[0] == startByte {
		// log.Printf("\n\n\n\nWe Got a startByte")
		s, err = x.r.ReadBytes(endByte)

		// fmt.Printf("DEBUG Read: ")
		// for i := 0; i < len(s); i++ {
		// 	fmt.Printf("%#02x ", s[i])
		// }
		// fmt.Printf("\n")

		if err != nil {
			log.Println("DEBUG READ BYTE ERROR ", err)
			// n = copy(b, s)
			return 0, err
		}

		ok, _, p := validator(s)
		// log.Printf("validator ok=%v err=%v\n", ok, err)

		for !ok {
			// scan to next endByte
			s2, err := x.r.ReadBytes(endByte)
			if err != nil {
				log.Println("DEBUG READ BYTE ERROR ", err)
				return 0, err
			}
			// fmt.Printf("DEBUG Read S2: ")
			// for i := 0; i < len(s2); i++ {
			// 	fmt.Printf("%#02x ", s2[i])
			// }
			// fmt.Printf("\n")
			//
			// fmt.Printf("DEBUG Read Before Append: ")
			// for i := 0; i < len(s); i++ {
			// 	fmt.Printf("%#02x ", s[i])
			// }
			// fmt.Printf("\n")

			s = append(s, s2...)

			// fmt.Printf("DEBUG Read After Append:  ")
			// for i := 0; i < len(s); i++ {
			// 	fmt.Printf("%#02x ", s[i])
			// }
			// fmt.Printf("\n")
			ok, _, p = validator(s)
			// log.Printf("validator take 2 ok=%v err=%v\n", ok, err)
		}
		if ok {
			n = copy(b, p)
			return n, nil
		}
		// if err != nil {
		// 	log.Println("DEBUG validator ERROR ", err)
		// }
		// for !ok {
		// 	b = append(b, s...)
		// 	s2, err := x.r.ReadBytes(endByte)
		// 	if err != nil {
		// 		log.Println("DEBUG READ BYTE ERROR ", err)
		// 	}
		// 	b = append(b, s...)
		// 	ok, err = validator(b)
		// 	if err != nil {
		// 		log.Println("DEBUG validator ERROR ", err)
		// 	}
		// }
		// if ok {
		// 	n = copy(b, s)
		// 	return n, nil
		// }

	} else {
		return 0, errPacketNoStartByte
	}
	return 0, nil
	// for {
	// 	n, err = x.r.ReadBytes(endByte)
	// }

}

func validator(b []byte) (bool, error, []byte) {
	var k []byte
	k = append(k, b...)

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
				// fmt.Printf("DEBUG validator startByte:              ")
				// for i := 0; i < len(buf); i++ {
				// 	fmt.Printf("%#02x ", buf[i])
				// }
				// fmt.Printf("\n")
			}
		case inMessage:
			// fmt.Printf("k = %d, len = %d\n", k, len(b))
			if k == len(b)-1 {
				state = done
			} else if v == escByte {
				state = afterEsc
			} else {
				state = inMessage
				buf = append(buf, v)
				// fmt.Printf("DEBUG validator inMessage:              ")
				// for i := 0; i < len(buf); i++ {
				// 	fmt.Printf("%#02x ", buf[i])
				// }
				// fmt.Printf("\n")
			}
		case afterEsc:
			if v == endByte && k == len(b)-1 {
				// fmt.Printf("DEBUG errPacketNotLongEnough:               ")
				// for i := 0; i < len(buf); i++ {
				// 	fmt.Printf("%#02x ", buf[i])
				// }
				// fmt.Printf("\n")
				return false, errPacketNotLongEnough, nil
			}
			state = inMessage
			buf = append(buf, v)
			// fmt.Printf("DEBUG validator afterEsc:               ")
			// for i := 0; i < len(buf); i++ {
			// 	fmt.Printf("%#02x ", buf[i])
			// }
			// fmt.Printf("\n")
		}
	}

	// fmt.Printf("DEBUG before crc:                       ")
	// for i := 0; i < len(buf); i++ {
	// 	fmt.Printf("%#02x ", buf[i])
	// }
	// fmt.Printf("\n")

	var crcByte byte

	n := len(buf)

	crcByte, buf = buf[n-1], buf[:n-1]
	n--

	crc, crcerr := checksum(&buf)
	// log.Printf("CRC %#02x, CRCBYTE %#02x", crc, crcByte)
	if crcerr != nil {
		return false, errPacketNoStartByte, nil
	}

	if crcByte != crc {
		return false, errPacketBadCRC, nil
	}

	buf = buf[:0+copy(buf[0:], buf[1:])]
	n--

	// fmt.Printf("DEBUG validator:              ")
	// for i := 0; i < len(buf); i++ {
	// 	fmt.Printf("%#02x ", buf[i])
	// }
	// fmt.Printf("\n")

	return true, nil, buf

	// fmt.Printf("DEBUG validator:              ")
	// for i := 0; i < len(k); i++ {
	// 	fmt.Printf("%#02x ", k[i])
	// }
	// fmt.Printf("\n")
	//
	// n := len(k)
	// // var last kyte
	// // pop endByte
	// _, k = k[n-1], k[:n-1]
	// n--
	//
	// fmt.Printf("DEBUG validator pop endbyte:  ")
	// for i := 0; i < len(k); i++ {
	// 	fmt.Printf("%#02x ", k[i])
	// }
	// fmt.Printf("\n")
	//
	// // pop crcByte to check later
	// var crcByte byte
	// crcByte, k = k[n-1], k[:n-1]
	// n--
	//
	// fmt.Printf("DEBUG validator pop crcbyte:  ")
	// for i := 0; i < len(k); i++ {
	// 	fmt.Printf("%#02x ", k[i])
	// }
	// fmt.Printf("\n")
	//
	// // delete escBytes
	// for i := 0; i < (n - 1); i++ {
	// 	if k[i] == escByte && k[i+1] == endByte {
	// 		k = k[:i+copy(k[i:], k[i+1:])]
	// 		n--
	// 	}
	// }
	//
	// fmt.Printf("DEBUG validator del escbytes: ")
	// for i := 0; i < len(k); i++ {
	// 	fmt.Printf("%#02x ", k[i])
	// }
	// fmt.Printf("\n")
	//
	// // check crcbyte
	// crc, crcerr := checksum(&k)
	// if crcerr != nil {
	// 	return false, errPacketNoStartByte, nil
	// }
	// log.Printf("DEBUG validator CRC, crc %#02x, crcbyte %#02x ", crc, crcByte)
	// if crcByte != crc {
	// 	return false, errPacketBadCRC, nil
	// }
	// // delete startByte
	// k = k[:0+copy(k[0:], k[1:])]
	// n--
	// // check for errors byte
	// if k[0] == errorByte {
	// 	switch protocolError(k[1]) {
	// 	case notReconsied:
	// 		return false, errProtocolErrorNotReconsied, nil
	// 	case crcFailed:
	// 		return false, errProtocolErrorCRCfailed, nil
	// 	case invaidAppID:
	// 		return false, errProtocolErrorInvaidAppID, nil
	// 	}
	// }
	//

}

var (
	errPacketNotLongEnough       = errors.New("not long enough")
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
