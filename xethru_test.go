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
	}
	for _, c := range cases {
		crc, err := c.p.Checksum()
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
