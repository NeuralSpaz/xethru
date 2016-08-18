package xethru

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
)

//
const (
	START             = 0x7D
	END               = 0x7E
	ESC               = 0x7F
	XTS_SPC_MOD_RESET = 0x22
)

type Sensor struct {
	Conn string
	buf  []byte
}

type packet struct {
	startpos  int
	datastart int
	dataend   int
	data      []byte
	escpos    []int
	endpos    int
	raw       []byte
	crc       []byte
}

type Breath struct {
	bpm      float64
	distance float64
	quality  float64
}

var UnsupportedTransport = errors.New("unsupported transport type")
var WindowsSerialConnection = errors.New("invalid serial port use COM")
var LinuxSerialConnection = errors.New("invalid serial port use /dev/tty")
var BaudRateNotanInterger = errors.New("baudrate is not an interger")
var BaudRateNotgiven = errors.New("baudrate must be set")

func NewSensor(transport string, connection string) (Sensor, error) {
	if transport != "tcp" && transport != "serial" && transport != "udp" {
		return Sensor{}, UnsupportedTransport
	}
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
	case "udp":
		host, port, err = net.SplitHostPort(connection)
		if err != nil {
			return Sensor{}, err
		}
		connection = net.JoinHostPort(host, port)
	case "serial":
		if runtime.GOOS == "windows" {
			if !strings.Contains(connection, "COM") {
				return Sensor{}, WindowsSerialConnection
			}
		}
		if runtime.GOOS == "linux" {
			if !strings.Contains(connection, "/dev/tty") {
				return Sensor{}, LinuxSerialConnection
			}
		}
		portbaud := strings.Split(connection, ":")
		if len(portbaud) < 2 {
			// fmt.Printf("Port Baud %+#v\n", portbaud)
			return Sensor{}, BaudRateNotgiven
		}
		port = portbaud[0]
		baud, err = strconv.Atoi(portbaud[1])
		if err != nil {
			return Sensor{}, BaudRateNotanInterger
		}
		fmt.Println(baud)
	}

	s := Sensor{Conn: connection}

	return s, nil
}

// Read Data from Sesor
func (s *Sensor) Read() {}

func (s *Sensor) Write() {}

func (s *Sensor) Close() {}

func (s *Sensor) Reset() {}

func (s *Sensor) Led() {}

func (s *Sensor) Load() {}

// func (p *packet) CRC() {}

func BaseBand() {}

func baseBand() {}

func load() {}

func led() {}

func reset() {}

func new() {}

func read() {}

func write() {}

func close() {}

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

func (s *Sensor) ping() {

}

// Checksum
// Calculated by XOR’ing all bytes from <START> + [Data].
// Note that the CRC is done after escape bytes is removed. This
// means that CRC is also calculated before adding escape bytes.
type CRC byte

func Checksum(packet []byte) (bool, CRC, error) {
	if packet[0] != START {
		return false, 0x00, ChecksumInvalidPacketSTART
	}
	var crc byte
	for _, b := range packet {
		crc = crc ^ b
	}

	return false, CRC(crc), nil
}

var ChecksumInvalidPacketSTART = errors.New("invalid packet missing start")

func checksum() {}

// Data escaping
// Escaping means that if the escape byte occurs in data,
// the next byte is not <Start>, <End> or <Esc>, but intended byte
// with same value as flags.
// Example: 0x7D + 0x10 + 0x7F + 0x7E + 0x04 + 0xFF + 0x7E
// Here the byte 0x7E in the middle is intended, and should not be read as a flag.
// The 0x7E byte is prepended with the
// escape byte 0x7F. After parsing for escape bytes, the data becomes:
// 0x7D + 0x10 + 0x7E + 0x04 + 0xFF + 0x7E

func parseEscapeChar(packet []byte) []byte { return nil }
