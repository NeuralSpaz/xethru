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
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"testing"
	"time"
)

// testing helper
func NewXethruWriter(w io.Writer) io.Writer {
	return &x2m200Frame{w: w}
}

func NewXethruReader(r io.Reader) io.Reader {
	return &x2m200Frame{r: bufio.NewReader(r)}
}

// CreateSplitReadWriter Used help with testing Framer
func CreateSplitReadWriter(w io.Writer, r io.Reader) Framer {
	return &x2m200Frame{w: w, r: bufio.NewReader(r)}
}

func x2m200ProtocolwithTransit(in []byte) ([]byte, []byte, error) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	xw := NewXethruWriter(w)

	_, err := xw.Write(in)
	w.Flush()
	if err != nil {
		return nil, nil, err
	}
	transit := b.Bytes()
	xr := NewXethruReader(&b)

	readback, err := ioutil.ReadAll(xr)
	if err != nil {
		return nil, nil, err
	}
	return readback, transit, err
}

func TestX2M200Write(t *testing.T) {

	cases := []struct {
		b      []byte
		n      int
		err    error
		writen []byte
	}{
		{[]byte{0x01, 0x02, 0x00}, 6, nil, []byte{0x7d, 0x01, 0x02, 0x00, 0x7e, 0x7e}},
		{[]byte{0x00, 0x7c, 0x7f}, 6, nil, []byte{0x7d, 0x00, 0x7c, 0x7f, 0x7e, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, 6, nil, []byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x00, 0x01, 0x02, 0x03}, 7, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x00, 0x01, 0x02, 0x7e}, 8, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x7f, 0x7e, 0x00, 0x7e}},
		{[]byte{0x7e, 0x01, 0x02, 0x7e}, 9, nil, []byte{0x7d, 0x7f, 0x7e, 0x01, 0x02, 0x7f, 0x7e, 0x7e, 0x7e}},
		{[]byte{0x7e, 0x7e, 0x02, 0x7e}, 10, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x02, 0x7f, 0x7e, 0x01, 0x7e}},
		{[]byte{0x7e, 0x7e, 0x7e, 0x7e}, 11, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7d, 0x7e}},
		{[]byte{0x01, 0xee, 0xaa, 0xea, 0xae}, 8, nil, []byte{0x7d, 0x01, 0xee, 0xaa, 0xea, 0xae, 0x7c, 0x7e}},
	}
	for _, c := range cases {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		x := NewXethruWriter(w)
		n, err := x.Write(c.b)
		w.Flush()

		if err != c.err {
			t.Errorf("Expected: %v, got %v\n", c.err, err)
		}
		if n != c.n {
			t.Errorf("Expected: %d, got %d\n", c.n, n)
		}
		if string(b.Bytes()) != string(c.writen) {
			t.Errorf("Expected: %d, got %d\n", c.writen, b.Bytes())
		}
	}
}

func TestX2M200Read(t *testing.T) {

	cases := []struct {
		readback []byte
		err      error
		writeout []byte
	}{
		{[]byte{0x01, 0x02, 0x00}, nil, []byte{0x7d, 0x01, 0x02, 0x00, 0x7f, 0x7e, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, nil, []byte{0x7d, 0x01, 0x02, 0x03, 0x7f, 0x7d, 0x7e}},
		{[]byte{0x00, 0x7c, 0x7f}, nil, []byte{0x7d, 0x00, 0x7c, 0x7f, 0x7f, 0x7e, 0x7e}},
		{[]byte{0x00, 0x01, 0x02, 0x03}, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x00, 0x01, 0x02, 0x7e}, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x7f, 0x7e, 0x00, 0x7e}},
		{[]byte{0x7e, 0x01, 0x02, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x01, 0x02, 0x7f, 0x7e, 0x7e, 0x7e}},
		{[]byte{0x7e, 0x7e, 0x02, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x02, 0x7f, 0x7e, 0x01, 0x7e}},
		{[]byte{0x7e, 0x7e, 0x7e, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7d, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, errPacketNoStartByte, []byte{0x1d, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{}, io.EOF, []byte{}},
		{[]byte{}, io.EOF, []byte{0x7d}},
		{[]byte{0x01, 0x02, 0x03}, errPacketBadCRC, []byte{0x7d, 0x01, 0x02, 0x03, 0x71, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, errProtocolErrorNotReconsied, []byte{0x7d, 0x20, 0x01, 0x5c, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, errProtocolErrorCRCfailed, []byte{0x7d, 0x20, 0x02, 0x5f, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, errProtocolErrorInvaidAppID, []byte{0x7d, 0x20, 0x03, 0x5e, 0x7e}},
		{[]byte{}, nil, []byte{0x7d, 0x7d, 0x7e}},
	}

	for _, c := range cases {
		r := bytes.NewReader(c.writeout)
		x := NewXethruReader(r)

		b := make([]byte, 1024)
		n, err := x.Read(b)
		readback := b[:n]

		if err != c.err {
			t.Errorf("Expected: %s, got %s\n", c.err, err)
		}

		if err == nil {
			if string(readback) != string(c.readback) {
				t.Errorf("Expected: %x, got %x\n", c.readback, readback)
			}
		}
	}
}

func newLoopBackXethru() (Framer, chan []byte, chan []byte) {
	sensorReader, clientWriter := io.Pipe()
	clientReader, sensorWriter := io.Pipe()
	client := CreateSplitReadWriter(clientWriter, clientReader)
	sensor := CreateSplitReadWriter(sensorWriter, sensorReader)

	sensorSend := make(chan []byte)
	sensorRecive := make(chan []byte)

	go func() {
		defer close(sensorSend)
		for {
			select {
			case <-time.After(time.Millisecond * 1000):
				return
			case p := <-sensorSend:
				_, err := sensor.Write(p)
				if err != nil {
					return
				}
			}
		}
	}()

	go func() {
		defer close(sensorRecive)
		for {
			b := make([]byte, 256)
			n, err := sensor.Read(b)
			if err != nil {
				return
			}
			sensorRecive <- b[:n]
		}
	}()

	return client, sensorSend, sensorRecive
}
