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
	"io"
	"log"
	"testing"
	"time"

	"github.com/NeuralSpaz/xethru"
)

func newLoopBackXethru() (xethru.Framer, chan []byte, chan []byte) {
	sensorReader, clientWriter := io.Pipe()
	clientReader, sensorWriter := io.Pipe()
	client := xethru.CreateSplitReadWriter(clientWriter, clientReader)
	sensor := xethru.CreateSplitReadWriter(sensorWriter, sensorReader)

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

		ok, err := Ping(client, c.timeout)

		if ok != c.ok {
			t.Errorf("expected %+v, got %+v", c.ok, ok)
		}

		if err != c.err {
			t.Errorf("expected %+v, got %+v", c.err, err)
		}
	}

}