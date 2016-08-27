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
	"encoding/binary"
	"errors"
	"flag"
	"log"
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
	x2 := xethru.Open("x2m200", port)

	for i := 0; i < 10; i++ {
		ok, err := Ping(x2, *pingTimeout)
		if err != nil {
			log.Fatalf("Error Communicating with Device: %v", err)
		}
		if !ok {
			log.Fatal("Device Not Ready")
		}
		log.Println("Got Pong")

		time.Sleep(*pingTimeout)
	}
}

const (
	x2m200PingCommand          = 0x01
	x2m200PingSeed             = 0xeeaaeaae
	x2m200PingResponseReady    = 0xaaeeaeea
	x2m200PingResponseNotReady = 0xaeeaeeaa
)

//
// // Ping takes a time.Durration and waits for a maxium of that time before
// // timing out, usefull for confirming configurations is working
// // a true return with no error means the the xethru module is ready to
// // to accept other commands.
func Ping(x xethru.Framer, t time.Duration) (bool, error) {
	resp := make(chan []byte)
	ping(x, resp)
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
func ping(x xethru.Framer, response chan []byte) {
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
