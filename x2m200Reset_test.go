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

// Reset Tests
package xethru

import (
	"bytes"
	"testing"
	"time"
)

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

		ok, err := client.Reset(time.Millisecond * 3)

		if ok != c.ok {
			t.Errorf("expected %+v, got %+v", c.ok, ok)
		}

		if err != c.err {
			t.Errorf("expected %+v, got %+v", c.err, err)
		}
	}

}
