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

// Xethru-GO a driver for the xethru sensor modules

// Package xethru: An open source implementation driver for xethru sensor modules.
// The current state of the api is still unstable and under active development.
// Contributions are welcome.
package xethru

import (
	"bufio"
	"io"
	"time"
)

// Open Creates a x2m200 xethu serial protocol from a io.ReadWriter
// it implements io.Reader and io.Writer
func Open(device string, port io.ReadWriteCloser) Framer {
	// fmt.Println("New instance of Xethru")
	// if device == "x2m200" {
	x := &x2m200Frame{
		w: port,
		r: bufio.NewReader(port),
		c: port,
	}
	// TODO: disable all feeds
	return x
}

// Framer is a wrapper for a serial protocol. it inserts the start, crc and end bytes for you
type Framer interface {
	io.Writer
	io.Reader
	io.Closer
	Reset() (bool, error)
}

type Module struct {
	f                  Framer
	AppID              [4]byte
	LEDMode            ledMode
	DetectionZoneStart float32
	DetectionZoneEnd   float32
	Sensitivity        uint32
	Timeout            time.Duration
	Data               chan interface{}
	// parser             func(b []byte) (interface{}, error)
}
