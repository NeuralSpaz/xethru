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
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"log"
	"testing"
	"time"

	"github.com/NeuralSpaz/xethru"
	"github.com/jacobsa/go-serial/serial"
)

func main() {
	log.Println("X2M200 Ping Demo")
	commPort := flag.String("commPort", "/dev/ttyUSB0", "the comm port you wish to use")
	baudrate := flag.Uint("baudrate", 115200, "the baud rate for the comm port you wish to use")
	pingTimeout := flag.Duration("pingTimeout", time.Millisecond*300, "timeout for ping command")
	flag.Parse()

	options := serial.OpenOptions{
		PortName:        *commPort,
		BaudRate:        *baudrate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	defer port.Close()
	x2 := xethru.Open(port)

	for i := 0; i < 10; i++ {
		ok, err := x2.Ping(*pingTimeout)
		if err != nil {
			log.Fatalf("Error Communicating with Device: %v", err)
		}
		if !ok {
			log.Fatal("Device Not Ready")
		}
		log.Println("Got Pong")

		time.Sleep(*pingTimeout)
	}

	//
	// appconfig := xetheu.AppConfig{
	// 	Name:        "Resp",
	// 	ZoneStart:   0.5,
	// 	ZoneEnd:     1.5,
	// 	LEDMode:     Full,
	// 	Sensitivity: 10,
	// 	Output:      os.Stdout,
	// }
	//
	// app, err := x2.LoadApp(appconfig)
	// if err != nil {
	// 	log.Fatalf("Error Loading App: %v", err)
	// }
	// app.Start()
}

const (
	x2m200PingCommand          = 0x01
	x2m200PingSeed             = 0xeeaaeaae
	x2m200PingResponseReady    = 0xaaeeaeea
	x2m200PingResponseNotReady = 0xaeeaeeaa
)

// Ping takes a time.Durration and waits for a maxium of that time before
// timing out, usefull for confirming configurations is working
// a true return with no error means the the xethru module is ready to
// to accept other commands.
func Ping(x x2m200Frame, t time.Duration) (bool, error) {
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

var errPingTimeout = errors.New("ping timeout")

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
			// time.Sleep(time.Millisecond * 100)
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

type app interface {
	GetStatus() error
	Reset() error
	parser([]byte) (bool, error)
}

func (x x2m200Frame) LoadApp(config AppConfig) app { return SleepingApp{} }

func StartApp(a app) (bool, error) { return false, nil }
func SetLED(a app) (bool, error)   { return false, nil }

const (
	appCommandByte      = 0x10
	appSet              = 0x10
	appAck              = 0x10
	appSetDetectionZone = 0x96a10a1c
)

func SetDetectionZone(a app) (bool, error) { return false, nil }
func SetSensitivity(a app) (bool, error)   { return false, nil }

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
	Framer
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
func (a SleepingApp) parser(b []byte) (bool, error) { return false, nil }

type RespirationApp struct {
	Framer
	Status        status
	RPM           float64
	Distance      float64
	SignalQuality float64
	Movement      float64
}

func (a RespirationApp) GetStatus() error              { return nil }
func (a RespirationApp) Reset() error                  { return nil }
func (a RespirationApp) String() string                { return "" }
func (a RespirationApp) parser(b []byte) (bool, error) { return false, nil }

type PresenceApp struct {
	Framer
	Status        status
	RPM           float64
	Distance      float64
	SignalQuality float64
	MovementSlow  float64
	MovementFast  float64
}

func (a PresenceApp) GetStatus() error              { return nil }
func (a PresenceApp) Reset() error                  { return nil }
func (a PresenceApp) String() string                { return "" }
func (a PresenceApp) parser(b []byte) (bool, error) { return false, nil }

type BaseBandApp struct {
	Framer
	Counter           int32
	NumOfBins         int32
	BinLength         float64
	SamplingFrequency float64
	CarrierFrequency  float64
	RangeOffset       float64
	SigI              []float64
	SigQ              []float64
}

func (a BaseBandApp) GetStatus() error              { return nil }
func (a BaseBandApp) Reset() error                  { return nil }
func (a BaseBandApp) String() string                { return "" }
func (a BaseBandApp) parser(b []byte) (bool, error) { return false, nil }

//
//

// Outputs the baseband amplitude and phase data of the application.
// Example: <Start> + <XTS_SPR_APPDATA> + [XTS_ID_BASEBAND_AMPLITUDE_PHASE(i)] + [Counter(i)] + [NumOfBins(i)] + [BinLength(f)] + [SamplingFrequency(f)] + [CarrierFrequency(f)] + [RangeOffset(f)] + [Amplitude(f)]
// + ... + [Phase(f)] + ... + <CRC> + <End>

// Example:
// <Start> + <XTS_SPC_DIR_COMMAND> + <XTS_SDC_APP_SETINT> + [XTS_SACR_OUTPUTBASEBAND(i)] + [Length(i)] + [EnableCode(i)] + <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>

func (a *BaseBandApp) Enable() error {
	n, err := a.Write([]byte{scpDirCommand, sdcAppSetInit, sacrOutputBasebad, outputlength, sacrOutputBasebadPhaseAmp})
	if err != nil {
		log.Println(err)
		return err
	}

	b := make([]byte, 20)
	n, err = a.Read(b)
	if err != nil {
		log.Printf("Ping Read Error %v, number of bytes %d\n", err, n)
		return err
	}
	ok, err := isBaseBandPhaseAmpResponseValid(b[:n])
	if !ok {
		return err
	}
	return nil
}

func isBaseBandPhaseAmpResponseValid(b []byte) (bool, error) {
	if len(b) == 0 {
		return false, errSetBaseBandPhaseAmpNotEnoughBytes
	}
	if b[0] == sprAck {
		return true, nil
	}
	return false, errSetBaseBandPhaseAmpInvalidResponse
}

var (
	errSetBaseBandPhaseAmpInvalidResponse = errors.New("invalid response")
	errSetBaseBandPhaseAmpNotEnoughBytes  = errors.New("Not enough bytes in resposnse")
)

const (
	scpDirCommand             = 0x90
	sdcAppSetInit             = 0x71
	sacrOutputBasebad         = 0x10
	sacrOutputBasebadOff      = 0x00
	sacrOutputBasebadPhaseAmp = 0x02
	outputlength              = 0x01
	sprAck                    = 0x10
)

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
func TestIsValidPingResponse(t *testing.T) {
	cases := []struct {
		b   []byte
		err error
		ok  bool
	}{
		{[]byte{0x01}, errPingNotEnoughBytes, false},
		{[]byte{0x02, 0x00, 0x00, 0x00, 0x00}, errPingDoesNotStartWithPingCMD, false},
		{[]byte{0x01, 0x01, 0x02, 0x03, 0x04}, errPingDoesNotContainResponse, false},
		{[]byte{0x01, 0xae, 0xea, 0xee, 0xaa}, nil, false},
		{[]byte{0x01, 0xaa, 0xee, 0xae, 0xea}, nil, true},
	}
	for _, c := range cases {
		ok, err := isValidPingResponse(c.b)

		if err != c.err {
			t.Errorf("expected %+v, got %+v", c.err, err)
		}
		if ok != c.ok {
			t.Errorf("expected %+v, got %+v", c.ok, ok)
		}
	}
}

func TestPing(t *testing.T) {

	cases := []struct {
		ok         bool
		err        error
		delaymS    time.Duration
		sensorSend []byte
		timeout    time.Duration
	}{
		{true, nil, time.Millisecond * 1, []byte{0x01, 0xaa, 0xee, 0xae, 0xea}, time.Millisecond * 2},
		{true, nil, time.Millisecond * 1, []byte{0x01, 0xaa, 0xee, 0xae, 0xea}, 0},
		{false, nil, time.Millisecond * 1, []byte{0x01, 0xae, 0xea, 0xee, 0xaa}, time.Millisecond * 2},
		{false, errPingTimeout, time.Millisecond * 4, []byte{0x01, 0xaa, 0xee, 0xae, 0xea}, time.Millisecond * 2},
		{false, errPingDoesNotContainResponse, time.Millisecond * 1, []byte{0x01, 0x02, 0x02, 0x02, 0x02}, time.Millisecond * 2},
		{false, errPingNotEnoughBytes, time.Millisecond * 1, []byte{0x01, 0x02, 0x02}, time.Millisecond * 2},
		{false, errPingDoesNotStartWithPingCMD, time.Millisecond * 1, []byte{0x50, 0x02, 0x02, 0x02, 0x04}, time.Millisecond * 2},
	}

	for _, c := range cases {

		client, sensorSend, sensorRecive := newLoopBackXethru()

		go func() {
			b := <-sensorRecive
			time.Sleep(c.delaymS)
			// fmt.Printf("%x", b)
			if bytes.Contains(b, []byte{0x01, 0xee, 0xaa, 0xea, 0xae}) {
				sensorSend <- c.sensorSend
			}
		}()

		ok, err := Ping(client, c.timeout)

		if ok != c.ok {
			t.Errorf("expected %+v, got %+v", c.ok, ok)
		}

		if err != c.err {
			t.Errorf("expected %+v, got %+v", c.err, err)
		}
	}

}

func TestEnableBaseBandPhaseAmp(t *testing.T) {

	cases := []struct {
		sensorSend   []byte
		sensorRecive []byte
		err          error
	}{
		{[]byte{sprAck}, []byte{scpDirCommand, sdcAppSetInit, sacrOutputBasebad, outputlength, sacrOutputBasebadPhaseAmp}, nil},
		{[]byte{0x00}, []byte{scpDirCommand, sdcAppSetInit, sacrOutputBasebad, outputlength, sacrOutputBasebadPhaseAmp}, errSetBaseBandPhaseAmpInvalidResponse},
		{[]byte{}, []byte{scpDirCommand, sdcAppSetInit, sacrOutputBasebad, outputlength, sacrOutputBasebadPhaseAmp}, errSetBaseBandPhaseAmpNotEnoughBytes},
	}

	for _, c := range cases {

		client, sensorSend, sensorRecive := newLoopBackXethru()

		a := BaseBandApp{client, 0, 0, 0, 0, 0, 0, nil, nil}

		go func() {
			b := <-sensorRecive
			if string(b) != string(c.sensorRecive) {
				t.Errorf("sensorRecive expected %x, got %x", c.sensorRecive, b)
			}
			sensorSend <- c.sensorSend
		}()

		err := a.Enable()
		if err != c.err {
			t.Errorf("expected %+v, got %+v", c.err, err)
		}
	}

}

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
//
//
//
//
//
