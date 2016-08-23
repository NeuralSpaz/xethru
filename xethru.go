// xethru  Copyright (C) 2016
// This work is copyright no part may be reproduced by any process,
// nor may any other exclusive right be exercised, without the permission of
// NeuralSpaz aka Josh Gardiner 2016
// It is my intent that this will be released as open source at some
// future time. If you would like to contribute please contact me.

package xethru

import (
	"encoding/binary"
	"errors"
	"io"
	"time"
)

// Flow Control bytes
// startByte + [data] + CRC + endByte
const (
	startByte = 0x7D
	endByte   = 0x7E
	escByte   = 0x7F
)

// type x2m200Framer struct {
// 	x2m200FrameWriter
// 	x2m200FrameReader
// 	// x2m200FrameCloser
// }

//
func NewReadWriter(r io.Reader, w io.Writer) framer {
	// if model == "x2m200" {
	// return x2m200Framer{x2m200FrameWriter{w}, x2m200FrameReader{r}, x2m200FrameCloser{nil}}
	return x2m200Frame{w: w, r: r}

	// }
}

// func NewReadWriteCloser(rwc io.ReadWriteCloser) io.ReadWriteCloser {
// 	// if model == "x2m200" {
// 	return &x2m200Framer{x2m200FrameWriter{rwc}, x2m200FrameReader{rwc}, x2m200FrameCloser{rwc}}
// 	// }
// }

type framer interface {
	io.Writer
	io.Reader
	Ping() (bool, error)
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

// type x2m200FrameCloser struct {
// 	c io.Closer
// }
//
// func (x *x2m200Frame) Close() error {
// 	return x.c.Close()
// }

// type x2m200FrameReader struct {
// 	r io.Reader
// }

func (x x2m200Frame) Read(b []byte) (n int, err error) {
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
			return 0, ErrorPacketNoStartByte
		}
		if crcByte != crc {
			return 0, ErrorPacketBadCRC
		}
		// delete startByte
		b = b[:0+copy(b[0:], b[1:])]
		// for i := 0; i < n; i++ {
		// 	fmt.Println(i)
		// }
		n--
		if n == 0 {
			return n, io.EOF
		}
		return n, nil
	}
	if err != nil {
		return 0, err
	}
	return 0, nil
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

const (
	x2m200PingCommand          = 0x01
	x2m200PingSeed             = 0xeeaaeaae
	x2m200PingResponseReady    = 0xaaeeaeea
	x2m200PingResponseNotReady = 0xaeeaeeaa
)

func (x x2m200Frame) Ping() (bool, error) {

	resp := ping(x)

	select {
	case <-time.After(time.Millisecond * 500):
		return false, errors.New("ping timeout")
	case r := <-resp:
		ok, err := isValidPingResponse(r)
		return ok, err
	}

	return false, nil

}

func ping(x framer) <-chan []byte {
	response := make(chan []byte)
	return response
}

func isValidPingResponse(b []byte) (bool, error) {
	// check response length is
	if len(b) != 5 {
		return false, errPingNotEnoughBytes
	}
	// Check response starts with Ping Byte
	if b[0] != x2m200PingCommand {
		return false, errPingDoesNotStartWithPingCMD
	}
	// check for valid response
	// x2m200PingResponseReady 0xaa ee ae ea
	// x2m200PingResponseNotReady = 0xae ea ee aa

	// maybe betterway to check this
	resp := binary.BigEndian.Uint32(b[1:])
	// log.Printf("const %x resp %x\n", x2m200PingResponseReady, resp)
	if resp != x2m200PingResponseReady && resp != x2m200PingResponseNotReady {
		return false, errPingDoesNotContainResponse
	}

	if resp == x2m200PingResponseNotReady {
		return false, nil
	}

	if resp == x2m200PingResponseReady {
		return true, nil
	}

	return false, errors.New("somthing went wrong")
}

var errPingDoesNotContainResponse = errors.New("ping response does not contain a valid ping response")
var errPingNotEnoughBytes = errors.New("ping response does not contain a valid ping response")
var errPingDoesNotStartWithPingCMD = errors.New("ping response does not start with ping response start byte")

//
//
//
//
//
//	// Build Request
// seed := make([]byte, 4)
// binary.BigEndian.PutUint32(seed, x2m200PingSeed)
// fmt.Printf("%x\n", seed)
// fmt.Printf("%x\n", x2m200PingSeed)
// command := append([]byte{x2m200PingCommand}, seed...)
// n, err := x.Write(command)
// if err != nil {
// 	fmt.Println(err, n)
// }
// // Read Response
//
// data := make([]byte, 56)
//
// n, err = x.Read(data)
//
// for n == 0 {
// 	n, err = x.Read(data)
// 	if err != nil {
// 		if err != io.EOF {
// 			fmt.Println(err)
// 		}
// 	}
// 	time.Sleep(time.Millisecond * 10)
// }
//
// fmt.Printf("Ping answer: %x\n", data)
//
// // fmt.Println("ping answer:", b)
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
