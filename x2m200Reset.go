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
	"errors"
	"io"
	"log"
)

const (
	resetCmd = 0x22
)

// Reset should be the first to be called when connecting to the X2M200 sensor
func (x x2m200Frame) Reset() (bool, error) {
	last := "disableBaseBand"

disableBaseBand:
	// log.Println("disableBaseBand")
	// Disbale Basebands
	//  ([]byte{0x90, 0x71, 0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00})
	n, err := x.Write([]byte{0x90, 0x71, 0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	last = "disableBaseBand"
	goto reRead

disableRespiration:
	// log.Println("Disable Respiration")
	// Disbale Basebands
	//    				   {0x90, 0x71, 0x11, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	// n, err = x.Write([]byte{0x90, 0x71, 0x11, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	n, err = x.Write([]byte{0x20, 0x11})
	last = "disableRespiration"
	goto reRead

reset:
	// log.Println("Reset")
	n, err = x.Write([]byte{resetCmd})
	if err != nil {
		// log.Printf("Reset Write Error %v, number of bytes %d\n", err, n)
		return false, err
	}
	last = "reset"
	goto reRead

reRead:
	b := make([]byte, 2048)
	n, err = x.Read(b)
	// log.Println("Reading")
	if err != nil {
		// log.Printf("Reset read Error %v, number of bytes %d\n", err, n)
		if err == io.EOF {
			return true, nil
		}
	}
	if n == 0 {
		goto reset
	}
	state, err := parse(b[:n])
	if err != nil {
		// log.Printf("Parse read Error %v, state %#+v \n", err, state)
		return false, err
	}
	// log.Printf("Debug state %#+v \n", state)
	switch state.(type) {
	case SystemMessage:
		s := state.(SystemMessage)
		if s.Message == "Command Ack'ed" {
			if last == "reset" {
				// log.Println("Yay we got there")
				return true, nil
			}
			goto reset
		}

		// return x.Reset()
	case BaseBandAmpPhase:
		goto disableBaseBand
	case Respiration:
		goto disableRespiration
	// case Sleep:
	// 	// TODO DISABLE Sleep
	default:
		log.Printf("\n\n%#+v\n\n", state)
		goto reRead

	}

	return false, nil
}

var errResetNotEnoughBytes = errors.New("reset not enough bytes in response")
var errResetResponseError = errors.New("reset did not contain a correct response")
