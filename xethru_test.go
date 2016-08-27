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
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

// testing helper
func NewXethruWriter(w io.Writer) io.Writer {
	return x2m200Frame{w: w}
}

func NewXethruReader(r io.Reader) io.Reader {
	return x2m200Frame{r: r}
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

func TestXethruWrite(t *testing.T) {

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

func TestXethruRead(t *testing.T) {

	cases := []struct {
		readback []byte
		err      error
		writeout []byte
	}{
		{[]byte{0x01, 0x02, 0x00}, nil, []byte{0x7d, 0x01, 0x02, 0x00, 0x7e, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, nil, []byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x00, 0x7c, 0x7f}, nil, []byte{0x7d, 0x00, 0x7c, 0x7f, 0x7e, 0x7e}},
		{[]byte{0x00, 0x01, 0x02, 0x03}, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x00, 0x01, 0x02, 0x7e}, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x7f, 0x7e, 0x00, 0x7e}},
		{[]byte{0x7e, 0x01, 0x02, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x01, 0x02, 0x7f, 0x7e, 0x7e, 0x7e}},
		{[]byte{0x7e, 0x7e, 0x02, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x02, 0x7f, 0x7e, 0x01, 0x7e}},
		{[]byte{0x7e, 0x7e, 0x7e, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7d, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, errorPacketNoStartByte, []byte{0x1d, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, errorPacketNotEndbyte, []byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7d}},
		{[]byte{0x01, 0x02, 0x03}, errorPacketBadCRC, []byte{0x7d, 0x01, 0x02, 0x03, 0x71, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, protocolErrorNotReconsied, []byte{0x7d, 0x20, 0x01, 0x5c, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, protocolErrorCRCfailed, []byte{0x7d, 0x20, 0x02, 0x5f, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, protocolErrorInvaidAppID, []byte{0x7d, 0x20, 0x03, 0x5e, 0x7e}},
		{[]byte{}, nil, []byte{0x7d, 0x7d, 0x7e}},
	}

	for _, c := range cases {
		r := bytes.NewReader(c.writeout)
		x := NewXethruReader(r)

		readback, err := ioutil.ReadAll(x)

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
		b := make([]byte, 256)
		n, err := sensor.Read(b)
		if err != nil {
			log.Printf("Ping Read Error %v, number of bytes %d\n", err, n)
		}
		// for {
		for n == 0 {
			n, err = sensor.Read(b)
			if err != nil {
				log.Printf("Ping Read Error %v, number of bytes %d\n", err, n)
				log.Printf("bytes %x\n", b)
			}
		}
		// log.Printf("sensorRecive %x\n", b[:n])
		sensorRecive <- b[:n]

		for {
			select {
			case <-time.After(time.Millisecond * 1000):
				return
			case p := <-sensorSend:
				n, err = sensor.Write(p)
				if err != nil {
					log.Printf("Ping Read Error %v, number of bytes %d\n", err, n)
					log.Printf("bytes %x\n", b)
				}
			}
		}

	}()

	return client, sensorSend, sensorRecive
}

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

		ok, err := client.Ping(c.timeout)

		if ok != c.ok {
			t.Errorf("expected %+v, got %+v", c.ok, ok)
		}

		if err != c.err {
			t.Errorf("expected %+v, got %+v", c.err, err)
		}
	}

}

func TestReset(t *testing.T) {

	cases := []struct {
		sensorRX  []byte
		sensorTX1 []byte
		sensorTX2 []byte
		sensorTX3 []byte
		delay     time.Duration
		ok        bool
		err       error
	}{
		{[]byte{resetCmd}, []byte{resetAck}, []byte{systemMesg, systemBooting}, []byte{systemMesg, systemReady}, time.Millisecond, true, nil},
		{[]byte{resetCmd}, []byte{resetAck}, []byte{systemMesg, systemBooting}, []byte{systemMesg, systemReady}, time.Millisecond * 5, false, errResetTimeout},
		{[]byte{resetCmd}, []byte{systemMesg, systemReady}, []byte{}, []byte{}, time.Millisecond, true, nil},
		{[]byte{resetCmd}, []byte{0x01, 0x02}, []byte{}, []byte{}, time.Millisecond, false, errResetTimeout},
	}

	for _, c := range cases {

		client, sensorSend, sensorRecive := newLoopBackXethru()

		go func() {
			b := <-sensorRecive
			time.Sleep(c.delay)
			// fmt.Printf("%x", b)
			if bytes.Contains(b, c.sensorRX) {
				sensorSend <- c.sensorTX1
				time.Sleep(c.delay)
				sensorSend <- c.sensorTX2
				time.Sleep(c.delay)
				sensorSend <- c.sensorTX3
				time.Sleep(c.delay)
			}
		}()

		ok, err := client.Reset(time.Millisecond * 2)

		if ok != c.ok {
			t.Errorf("expected %+v, got %+v", c.ok, ok)
		}

		if err != c.err {
			t.Errorf("expected %+v, got %+v", c.err, err)
		}
	}

}
