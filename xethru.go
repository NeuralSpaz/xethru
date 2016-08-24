// Copyright © 2016 Josh Gardiner aka NeuralSpaz on github.com
//
// This file is part of Xethru-Go - A Golang library for the xethru modules
//

package xethru

import (
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

func NewReadWriter(r io.Reader, w io.Writer) framer {
	return x2m200Frame{w: w, r: r}
}

type framer interface {
	io.Writer
	io.Reader
	Ping(t time.Duration) (bool, error)
	LoadApp(config AppConfig) App
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
	if n > 0 {
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
		return false, errPingTimeout
	case r := <-resp:
		ok, err := isValidPingResponse(r)
		return ok, err
	}

	return false, nil

}

var errPingTimeout = errors.New("ping timeout")

func (x x2m200Frame) ping(response chan []byte) {
	// response := make(chan []byte)
	go func() {
		// build ping command
		// find betterway to do this
		seed := make([]byte, 4)
		binary.BigEndian.PutUint32(seed, x2m200PingSeed)
		// fmt.Printf("seed %x\n", seed)
		cmd := []byte{x2m200PingCommand, seed[0], seed[1], seed[2], seed[3]}
		// Write to framer
		n, err := x.Write(cmd)
		// x.w.Flush()
		if err != nil {
			log.Printf("Ping Write Error %v, number of bytes %d\n", err, n)
		}

		// Read from framer
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
			time.Sleep(time.Millisecond * 100)
		}
		// send response []byte back to caller
		response <- b[:n]

	}()

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

var errPingDoesNotContainResponse = errors.New("ping response does not contain a valid ping response")
var errPingNotEnoughBytes = errors.New("ping response does not contain correct number of bytes")
var errPingDoesNotStartWithPingCMD = errors.New("ping response does not start with ping response start byte")

type AppConfig struct {
	Name        string
	ZoneStart   float64
	ZoneEnd     float64
	LEDMode     string
	Sensitivity float64
	Output      io.Writer
	parser      func()
}

type App interface {
	GetStatus() error
	Reset() error
	parser([]byte) (bool, error)
}

func (x x2m200Frame) LoadApp(config AppConfig) App { return SleepingApp{} }

func StartApp(app App) (bool, error)         { return false, nil }
func SetLED(app App) (bool, error)           { return false, nil }
func SetDetectionZone(app App) (bool, error) { return false, nil }
func SetSensitivity(app App) (bool, error)   { return false, nil }

type status int

const (
	breathing status = iota
	movement
	tracking
	noMovement
	initializing
	reserved
	unknown
)

type SleepingApp struct {
	framer
	Status        status
	RPM           float64
	Distance      float64
	SignalQuality float64
	MovementSlow  float64
	MovementFast  float64
}

func (a SleepingApp) GetStatus() error              { return nil }
func (a SleepingApp) Reset() error                  { return nil }
func (a SleepingApp) String() string                { return "" }
func (a SleepingApp) Parser(b []byte) (bool, error) { return false, nil }

type RespirationApp struct{}

func (a RespirationApp) GetStatus() error              { return nil }
func (a RespirationApp) Reset() error                  { return nil }
func (a RespirationApp) String() string                { return "" }
func (a RespirationApp) Parser(b []byte) (bool, error) { return false, nil }

type PresenceApp struct{}

func (a PresenceApp) GetStatus() error              { return nil }
func (a PresenceApp) Reset() error                  { return nil }
func (a PresenceApp) String() string                { return "" }
func (a PresenceApp) Parser(b []byte) (bool, error) { return false, nil }

type BaseBandApp struct{}

func (a BaseBandApp) GetStatus() error              { return nil }
func (a BaseBandApp) Reset() error                  { return nil }
func (a BaseBandApp) String() string                { return "" }
func (a BaseBandApp) Parser(b []byte) (bool, error) { return false, nil }

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
