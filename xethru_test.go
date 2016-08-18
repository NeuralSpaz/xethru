package xethru

import (
	"fmt"
	"runtime"
	"testing"
)

func TestNewSensor(t *testing.T) {
	if runtime.GOOS == "linux" {
		cases := []struct {
			transport, connection string
			err                   error
		}{
			{"tcp", "192.168.1.1:8080", nil},
			{"udp", "192.168.1.1:8080", nil},
			{"serial", "/dev/ttyUSB0:9600", nil},
			{"serial", "/dev/ttyUSB0", errBaudRateNotgiven},
			{"serial", "/dev/ttyUSB0:junk", errBaudRateNotanInterger},
			{"serial", "COM10:9600", errLinuxSerialConnection},
			{"junk", "morejunk", errUnsupportedTransport},
		}
		for _, c := range cases {
			_, err := NewSensor(c.transport, c.connection)
			if err != c.err {
				t.Errorf("could not create sensor transport: %v, connection: %v; got %+#v wanted %+#v\n", c.transport, c.connection, err, c.err)
			}
		}
		netcases := []struct {
			transport, connection string
			err                   string
		}{
			{"udp", "192.168.1.18080", "missing port in address 192.168.1.18080"},
			{"tcp", "192.168.1.18080", "missing port in address 192.168.1.18080"},
		}
		for _, c := range netcases {
			_, err := NewSensor(c.transport, c.connection)
			strerror := fmt.Sprintf("%s", err)
			if strerror != c.err {
				t.Errorf("could not create sensor transport: %v, connection: %v; got %+#v wanted %+#v\n", c.transport, c.connection, strerror, c.err)
			}
		}
	}
}

func TestChecksum(t *testing.T) {
	cases := []struct {
		p   packet
		crc CRC
		err error
	}{
		{[]byte{0x00, 0x01, 0x02}, 0x00, errChecksumInvalidPacketSTART},
		{[]byte{startByte, 0x01, 0x02}, 0x7E, nil},
		{[]byte{startByte, 0x01, 0x02, 0x03}, 0x7D, nil},
		{[]byte{startByte, 0x01, 0x02, 0xFF}, 0x81, nil},
		{[]byte{startByte, 0x01, 0x02, 0x7F}, 0x01, nil},
	}
	for _, c := range cases {
		crc, err := c.p.checksum()
		if err != c.err {
			t.Errorf("Expected: %v, got %v\n", c.err, err)
		}
		// if valid != c.valid {
		// 	t.Errorf("Expected: %v, got %v\n", c.valid, valid)
		// }
		if crc != c.crc {
			t.Errorf("Expected: %X, got %X\n", c.crc, crc)
		}
	}
}

func TestWrite(t *testing.T) {
	cases := []struct {
		b   []byte
		n   int
		err error
	}{
		{[]byte{0x01, 0x02, 0x03}, 6, nil},
		{[]byte{0x00, 0x01, 0x02, 0x03}, 7, nil},
		{[]byte{0x00, 0x01, 0x02, 0x7e}, 8, nil},
		{[]byte{0x7e, 0x01, 0x02, 0x7e}, 9, nil},
	}
	s, err := NewSensor("serial", "/dev/ttyUSB0:9600")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		n, err := s.write(c.b)
		if err != c.err {
			t.Errorf("Expected: %v, got %v\n", c.err, err)
		}
		// if valid != c.valid {
		// 	t.Errorf("Expected: %v, got %v\n", c.valid, valid)
		// }
		if n != c.n {
			t.Errorf("Expected: %d, got %d\n", c.n, n)
		}
	}

}
