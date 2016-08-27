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
	"errors"
	"log"
	"time"
)

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
