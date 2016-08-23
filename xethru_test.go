// xethru  Copyright (C) 2016
// This work is copyright no part may be reproduced by any process,
// nor may any other exclusive right be exercised, without the permission of
// NeuralSpaz aka Josh Gardiner 2016
// It is my intent that this will be released as open source at some
// future time. If you would like to contribute please contact me.
package xethru

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"testing"
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
		{[]byte{0x01, 0x02, 0x03}, ErrorPacketNoStartByte, []byte{0x1d, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x01, 0x02, 0x03}, ErrorPacketNotEndbyte, []byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7d}},
		{[]byte{0x01, 0x02, 0x03}, ErrorPacketBadCRC, []byte{0x7d, 0x01, 0x02, 0x03, 0x71, 0x7e}},
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

func NewTestDevice(name string) io.ReadWriter { return nil }

func TestPing(t *testing.T) {
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

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
