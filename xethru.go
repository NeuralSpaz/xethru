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

package xethru

import (
	"io"
	"time"
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
