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

// Reset

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

// Reset should be the first to be called when connecting to the X2M200 sensor
func (x x2m200Frame) Reset() (bool, error) {

	n, err := x.Write([]byte{resetCmd})

	if err != nil {
		log.Printf("Reset Write Error %v, number of bytes %d\n", err, n)
	}

	b := make([]byte, 4096)
	n, err = x.Read(b)
	if err != nil {
		log.Printf("Reset Read Error %v, number of bytes %d\n", err, n)
	}
	ok, err := isValidResetResponse(b[:n])
	if err != nil {
		log.Printf("Reset Read Error %v, number of bytes %d\n", err, n)
	}
	for !ok {
		time.Sleep(time.Millisecond * 5)
		b := make([]byte, 1024)
		n, err = x.Read(b)
		if err != nil {
			log.Printf("Reset Read Error %v, number of bytes %d\n", err, n)
		}
		ok, err = isValidResetResponse(b[:n])
		if err != nil {
			log.Printf("Reset Read Error %v, number of bytes %d\n", err, n)
		}
	}
	return true, nil
}

var errResetTimeout = errors.New("reset timeout")

func isValidResetResponse(b []byte) (bool, error) {
	log.Printf("RESET DEBUG: %#02x", b)
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
		return true, nil
	}
	return false, errResetResponseError
}

var errResetNotEnoughBytes = errors.New("reset not enough bytes in response")
var errResetResponseError = errors.New("reset did not contain a correct response")
