package xethru

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
)

//
const (
	startByte          = 0x7D
	endByte            = 0x7E
	escByte            = 0x7F
	XTSSPCModResetByte = 0x22
)

// Sensor config
type Sensor struct {
	Conn   net.Conn
	mode   string
	ip     string
	port   string
	baud   int
	serial *os.File
	buf    []byte
}

// type packet struct {
// 	startpos  int
// 	datastart int
// 	dataend   int
// 	data      []byte
// 	escpos    []int
// 	endpos    int
// 	raw       []byte
// 	crc       []byte
// }

// type Breath struct {
// 	bpm      float64
// 	distance float64
// 	quality  float64
// }

// Config Errors
var errUnsupportedTransport = errors.New("unsupported transport type")
var errWindowsSerialConnection = errors.New("invalid serial port use COM")
var errLinuxSerialConnection = errors.New("invalid serial port use /dev/tty")
var errBaudRateNotanInterger = errors.New("baudrate is not an interger")
var errBaudRateNotgiven = errors.New("baudrate must be set")

// NewSensor use this to build configuration for XM200
func NewSensor(transport string, connection string) (Sensor, error) {
	if transport != "tcp" && transport != "serial" && transport != "udp" {
		return Sensor{}, errUnsupportedTransport
	}
	s := Sensor{}
	var host, port string
	var baud int
	var err error
	switch transport {
	case "tcp":
		host, port, err = net.SplitHostPort(connection)
		if err != nil {
			return Sensor{}, err
		}

		connection = net.JoinHostPort(host, port)
		fmt.Println(connection)
		s.Conn, err = net.Dial("tcp", connection)
		if err != nil {
			log.Println(err)
			return s, err
		}
		s.mode = "network"
	case "udp":
		host, port, err = net.SplitHostPort(connection)
		if err != nil {
			return Sensor{}, err
		}
		connection = net.JoinHostPort(host, port)
	case "serial":
		if runtime.GOOS == "windows" {
			if !strings.Contains(connection, "COM") {
				return Sensor{}, errWindowsSerialConnection
			}
		}
		if runtime.GOOS == "linux" {
			if !strings.Contains(connection, "/dev/tty") {
				return Sensor{}, errLinuxSerialConnection
			}
		}
		portbaud := strings.Split(connection, ":")
		if len(portbaud) < 2 {
			// fmt.Printf("Port Baud %+#v\n", portbaud)
			return Sensor{}, errBaudRateNotgiven
		}
		port = portbaud[0]
		baud, err = strconv.Atoi(portbaud[1])
		if err != nil {
			return Sensor{}, errBaudRateNotanInterger
		}
		fmt.Println(baud, port)
	}

	return s, nil
}

// Read Data from Sesor
// func (s *Sensor) Read() {}
//

// Write slice of []byes sensor without start, esc, crc or end bytes
// implements io.Writer
func (s *Sensor) Write(b []byte) (n int, err error) {
	return s.write(b)
}

// Actual Impeemtation of Write
func (s *Sensor) write(b []byte) (n int, err error) {
	b = append(b[:0], append([]byte{startByte}, b[0:]...)...)
	// Should be no error here ever as we just set the startByte
	crc, _ := checksum(&b)

	// escape non endBytes 7e's
	for k := 0; k < len(b); k++ {
		if b[k] == endByte {
			b = append(b[:k], append([]byte{escByte}, b[k:]...)...)
			k++
		}
	}

	// add crc
	b = append(b, byte(crc))
	// add endByte
	b = append(b, endByte)

	// actually write to interface here
	// file descriper if serial
	// net.Conn if tcp

	if s.mode == "network" {
		return s.Conn.Write(b)

	}

	if s.mode == "usb" {
		return s.serial.Write(b)
	}

	return len(b), nil
}

// Close closes the connections to sensor
func (s *Sensor) Close() {}

//
// func (s *Sensor) Reset() {}
//
// func (s *Sensor) Led() {}
//
// func (s *Sensor) Load() {}
//
// // func (p *packet) CRC() {}
//
// func BaseBand() {}
//
// func baseBand() {}
//
// func load() {}
//
// func led() {}
//
// func reset() {}
//
// func new() {}
//
// func read() {}
//
func write() {}

//
// func close() {}

// func crc() {}

// Ping
// The ping command can be used to check connection to the module, and verify module readyness.
// During the module boot procedure, when connecting using USB, it can be difficult to make sure you are able to receive the
// system status messages saying that the module is ready. In that case, the PING command comes in handy, where you can
// ask the module if it is ready to receive commands.
// Example: <Start> + <XTS_SPC_PING> + [XTS_DEF_PINGVAL(i)] + <CRC> + <End>
// Response: <Start> + <XTS_SPR_PONG> + [Pongval(i)] + <CRC> + <End>
// Protocol codes:
// Name 						Value 				Description
// XTS_SPC_PING 				0x01 		Ping command code
// XTS_DEF_PINGVAL 				0xeeaaeaae 	Ping seed value
// XTS_SPR_PONG 				0x01 		Pong responce code
// XTS_DEF_PONGVAL_READY 		0xaaeeaeea 	Module is ready
// XTS_DEF_PONGVAL_NOTREADY 	0xaeeaeeaa 	Module is not ready

// func (s *Sensor) ping() {
//
// }

// CRC is the the calcuated CRC checksum
type CRC byte

type packet []byte
type data []byte

// Packet contains all the bytes
// type Packet struct {
// 	p packet
// 	d data
// 	c CRC
// }

// Checksum
// Calculated by XOR’ing all bytes from <START> + [Data].
// Note that the CRC is done after escape bytes is removed. This
// means that CRC is also calculated before adding escape bytes.
func checksum(p *[]byte) (CRC, error) {
	if (*p)[0] != startByte {
		return 0x00, errChecksumInvalidPacketSTART
	}
	var crc byte
	for _, b := range *p {
		crc = crc ^ b
	}

	return CRC(crc), nil
}

var errChecksumInvalidPacketSTART = errors.New("invalid packet missing start")

// func checksum() {}
//
// // Data escaping
// // Escaping means that if the escape byte occurs in data,
// // the next byte is not <Start>, <End> or <Esc>, but intended byte
// // with same value as flags.
// // Example: 0x7D + 0x10 + 0x7F + 0x7E + 0x04 + 0xFF + 0x7E
// // Here the byte 0x7E in the middle is intended, and should not be read as a flag.
// // The 0x7E byte is prepended with the
// // escape byte 0x7F. After parsing for escape bytes, the data becomes:
// // 0x7D + 0x10 + 0x7E + 0x04 + 0xFF + 0x7E
//
// func parseEscapeChar(packet []byte) []byte { return nil }
