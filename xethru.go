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
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"time"
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
		return x2m200Frame{w: port, r: port}
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
	Ping(t time.Duration) (bool, error)
	Reset(t time.Duration) (bool, error)
}

type App interface {
	Load() (bool, error)
	Set() (bool, error)
	Exec() (bool, error)
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

const (
	x2m200PingCommand          = 0x01
	x2m200PingSeed             = 0xeeaaeaae
	x2m200PingResponseReady    = 0xaaeeaeea
	x2m200PingResponseNotReady = 0xaeeaeeaa
)

func (x x2m200Frame) Ping(t time.Duration) (bool, error) {
	resp := make(chan []byte)
	x.ping(resp)
	if t == 0 {
		t = time.Millisecond * 100
	}
	select {
	case <-time.After(t):

	case r := <-resp:
		ok, err := isValidPingResponse(r)
		return ok, err
	}

	return false, errPingTimeout

}

//
var errPingTimeout = errors.New("ping timeout")

//
func (x x2m200Frame) ping(response chan []byte) {
	go func() {
		// build ping command
		// find betterway to do this
		seed := make([]byte, 4)
		binary.BigEndian.PutUint32(seed, x2m200PingSeed)
		// fmt.Printf("seed %x\n", seed)
		cmd := []byte{x2m200PingCommand, seed[0], seed[1], seed[2], seed[3]}
		// Write to Framer
		n, err := x.Write(cmd)
		// x.w.Flush()
		if err != nil {
			log.Printf("Ping Write Error %v, number of bytes %d\n", err, n)
		}

		// Read from Framer
		b := make([]byte, 20)
		n, err = x.Read(b)
		if err != nil {
			log.Printf("Ping Read Error %v, number of bytes %d\n", err, n)
		}
		// retry
		for n == 0 {
			n, err = x.Read(b)
			if err != nil {
				log.Printf("Ping Read Error %v, number of bytes %d\n", err, n)
				log.Printf("bytes %x\n", b)
			}
		}
		// send response []byte back to caller
		response <- b[:n]

	}()

}

//
func isValidPingResponse(b []byte) (bool, error) {
	// check response length is
	if len(b) != 5 {
		return false, errPingNotEnoughBytes
	}
	// Check response starts with Ping Byte
	if b[0] != x2m200PingCommand {
		return false, errPingDoesNotStartWithPingCMD
	}
	// check for valid response first striping off startByte
	resp := binary.BigEndian.Uint32(b[1:])
	switch resp {
	case x2m200PingResponseNotReady:
		return false, nil
	case x2m200PingResponseReady:
		return true, nil
	default:
		return false, errPingDoesNotContainResponse
	}
}

//
var errPingDoesNotContainResponse = errors.New("ping response does not contain a valid ping response")
var errPingNotEnoughBytes = errors.New("ping response does not contain correct number of bytes")
var errPingDoesNotStartWithPingCMD = errors.New("ping response does not start with ping response start byte")

const (
	resetCmd      = 0x22
	resetAck      = 0x10
	systemMesg    = 0x30
	systemBooting = 0x10
	systemReady   = 0x11
)

func (x x2m200Frame) Reset(t time.Duration) (bool, error) {
	//TODO pullout comms timeouts to flags with sensible defaults
	if t == 0 {
		t = time.Millisecond * 100
	}
	response := make(chan []byte)
	done := make(chan bool)
	n, err := x.Write([]byte{resetCmd})

	if err != nil {
		log.Printf("Ping Write Error %v, number of bytes %d\n", err, n)
	}
	go func() {
		for {

			// for _ := range done {
			b := make([]byte, 20)
			n, err = x.Read(b)
			if err != nil {
				log.Printf("Ping Read Error %v, number of bytes %d\n", err, n)
			}
			// send response []byte back to caller
			response <- b[:n]
			d := <-done
			if d {
				close(done)
				return
			}

		}
	}()

	for {
		select {
		case <-time.After(t):
			return false, errResetTimeout
		case resp := <-response:
			ok, err := isValidResetResponse(resp)
			if err != nil {
				log.Printf("Error: %v, response: %x\n", err, resp)
			}
			if ok && err == nil {
				done <- true
				close(response)
				return true, nil
			}
			done <- false
		}
	}

}

var errResetTimeout = errors.New("reset timeout")

func isValidResetResponse(b []byte) (bool, error) {
	if len(b) == 0 {
		return false, errResetNotEnoughBytes
	}
	if bytes.Contains(b, []byte{systemMesg, systemReady}) {
		log.Println("System Ready")
		return true, nil
	}
	if bytes.Contains(b, []byte{systemMesg, systemBooting}) {
		log.Println("System Booting")
		return false, nil
	}
	if b[0] == resetAck {
		log.Println("Reset command confirmed")
		return false, nil
	}
	return false, errResetResponseError
}

var errResetNotEnoughBytes = errors.New("reset not enough bytes in response")
var errResetResponseError = errors.New("reset did not contain a correct response")
