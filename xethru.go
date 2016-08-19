package xethru

import (
	"errors"
	"fmt"
	"io"
	"log"
)

// Flow Control bytes
// startByte + [data] + CRC + endByte
const (
	startByte = 0x7D
	endByte   = 0x7E
	escByte   = 0x7F
)

// Sensor config
// type Sensor struct {
// 	conn net.Conn
// 	mode string
// 	ip   string
// 	port string
// 	baud int
// 	// serial *os.File
// 	// buf    *io.Reader
// }

// Sensorer implements io.Writer and io.Reader
type Sensorer interface {
	// NewXethruWriter(w io.Writer) io.Writer
	// Write(p []byte) (n int, err error)
	// io.Reader
	// Read(p []byte) (n int, err error)
	// Checksum(p []byte) (crc byte, err error)
}

// Write slice of []byes sensor without start, esc, crc or end bytes
// implements io.Writer
// func (s Sensor) Write(b []byte) (n int, err error) {
// 	// log.Println("returning my writer")
// 	return xethruPacketWriter()
// 	// return s.Conn.Write(b)
// }

type xethruFrameWriter struct {
	w io.Writer
}

// NewXethruWriter builds data frame
func NewXethruWriter(w io.Writer) io.Writer {
	return &xethruFrameWriter{w}
}

func (x *xethruFrameWriter) Write(p []byte) (n int, err error) {

	p = append(p[:0], append([]byte{startByte}, p[0:]...)...)
	crc, _ := checksum(p)
	for k := 0; k < len(p); k++ {
		if p[k] == endByte {
			p = append(p[:k], append([]byte{escByte}, p[k:]...)...)
			k++
		}
	}
	p = append(p, crc)
	p = append(p, endByte)

	n, err = x.w.Write(p)
	return
}

type xethruFrameReader struct {
	r      io.Reader
	err    error
	toRead []byte
	// step      func(*xethruFrameReader)
	// stepState int
}

func NewXethruReader(r io.Reader) io.Reader {
	// respByte, err := ioutil.ReadAll(r)
	return &xethruFrameReader{
		r:      r,
		err:    nil,
		toRead: make([]byte, 512),
	}
}

func (x *xethruFrameReader) Read(b []byte) (n int, err error) {
	// x.toRead, x.err = ioutil.ReadAll(x.r)
	// fmt.Println("re reeading")
	for {
		// time.Sleep(time.Second * 1)
		n, x.err = x.r.Read(x.toRead)
		log.Println("looping")
		log.Println(len(x.toRead))
		log.Println(n)
		if n > 0 {
			log.Printf("bytes before processing %x\n", x.toRead)
			//
			// if x.toRead[0] == startByte {
			// 	x.toRead = x.toRead[1:n]
			// 	n--
			// 	log.Printf("after cutting start byte out and trailing %x\n", x.toRead)
			// }
			x.toRead, n, err = xethuProtocol(x.toRead, n)
			if err != nil {
				log.Println(err)
			}

			copy(b, x.toRead)
			log.Println("return size bytes b ", b)
			log.Println(len(b))
			if len(x.toRead) == 0 {
				log.Println("returning length n EOF")
				return n, io.EOF
			}
			log.Printf("returning length %d and nil\n", n)
			return n, nil
		}
		if x.err != nil {
			log.Println("error length zero length ", x.err)
			return 0, x.err
		}
	}
	return 0, io.EOF
}

func xethuProtocol(b []byte, n int) ([]byte, int, error) {
	if len(b) == 0 {
		return nil, 0, errors.New("zero size")
	}
	if b[0] != startByte {
		return nil, 0, errors.New("no startbyte")
	}
	var last byte
	last, b = b[n-1], b[:n-1]
	n--
	if last != endByte {
		return nil, 0, errors.New("does not end with endbyte")
	}
	// for i := 0; i < n; i++ {
	// 	log.Println("b position", i)
	// 	if b[i] == endByte && b[i-1] != escByte {
	// 		log.Println("endbyte")
	// 		b = b[:i]
	// 		n--
	// 		log.Println("removed endByte")
	// 		break
	// 	}
	// }

	for i := 0; i < n; i++ {
		log.Println("b position", i)
		if b[i] == escByte {
			log.Println("escByte")
			b = b[:i+copy(b[i:], b[i+1:])]

			// b = b[:i]
			n--
			// break
		}
	}

	// pop off crc byte
	var crcByte byte
	crcByte, b = b[len(b)-1], b[:len(b)-1]
	n--

	crc, err := checksum(b)
	if err != nil {
		log.Println(err)
	}

	//
	if crcByte != crc {
		return nil, 0, errors.New("failed checksum")
	}

	// cut startByte off
	b = b[1:]
	n--
	// else {
	// 	b = b[1:]
	// 	b = b[:len(b)-1]
	// }

	return b, n, nil
}

// func (x *xethruFrameReader) Read(p []byte) (n int, err error) {
// n, err = x.r.Read(p)
//
// if err != nil {
// 	return 0, err
// }
// if n > 0 {
// 	fmt.Println(p[0:n])
// 	p = p[1:n]
// 	n = n - 1
// }
// err = nil
// // fmt.Println("bytes read ", n)
// // if n > 0 {
// // 	fmt.Println(p[0:n])
// // 	p = p[1:2]
// // 	n = n - 1
// // }
// // fmt.Println(p[0:n])
// // if err != nil {
// // 	log.Printf("%x: %v", p, err)
// // 	// if p[0] != startByte {
// // 	// 	log.Println(p[0:n])
// // 	// 	return n, errors.New("no start byte")
// // 	// }
// //
// // }
//
// // if err != nil {
// // 	log.Printf("%x: %v", p[0:n], err)
// // } else {
// // 	log.Printf("%x", p[0:n])
// // }
//
// if n > 0 {
// 	return n, nil
// }
//
// return x.Read(p)
// }
//
//
//
//
//

// func (f *decompressor) Read(b []byte) (int, error) {
// 	for {
// 		if len(f.toRead) > 0 {
// 			n := copy(b, f.toRead)
// 			f.toRead = f.toRead[n:]
// 			if len(f.toRead) == 0 {
// 				return n, f.err
// 			}
// 			return n, nil
// 		}
// 		if f.err != nil {
// 			return 0, f.err
// 		}
// 		f.step(f)
// 	}
// }

// func (s Sensor) Read(b []byte) (n int, err error) {
// 	// log.Println("returning my writer")
// 	return s.senorReader(b)
// 	// return s.Conn.Write(b)
// }

// func xethruPacketReader(io.Writer) io.Writer {

// }

// Config Errors
// var errUnsupportedTransport = errors.New("unsupported transport type")
// var errWindowsSerialConnection = errors.New("invalid serial port use COM")
// var errLinuxSerialConnection = errors.New("invalid serial port use /dev/tty")
// var errBaudRateNotanInterger = errors.New("baudrate is not an interger")
// var errBaudRateNotgiven = errors.New("baudrate must be set")

// NewSensor use this to build configuration for X2M200
// func NewSensor(transport string, connection string) (Sensor, error) {
// 	if transport != "tcp" && transport != "serial" && transport != "udp" {
// 		return Sensor{}, errUnsupportedTransport
// 	}
// 	s := Sensor{}
// 	var host, port string
// 	var baud int
// 	var err error
// 	switch transport {
// 	case "tcp":
// 		host, port, err = net.SplitHostPort(connection)
// 		if err != nil {
// 			return Sensor{}, err
// 		}

// 		connection = net.JoinHostPort(host, port)
// 		// fmt.Println(connection)
// 		s.conn, err = net.Dial("tcp", connection)
// 		if err != nil {
// 			log.Println(err)
// 			return s, err
// 		}
// 		s.mode = "network"
// 	case "udp":
// 		host, port, err = net.SplitHostPort(connection)
// 		if err != nil {
// 			return Sensor{}, err
// 		}
// 		connection = net.JoinHostPort(host, port)
// 		srvaddr, err := net.ResolveUDPAddr("udp", connection)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		locaddr, err := net.ResolveUDPAddr("udp", "localhost:0")
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		s.conn, err = net.DialUDP("udp", locaddr, srvaddr)
// 		if err != nil {
// 			log.Println(err)
// 			return s, err
// 		}
// 		s.mode = "network"
// 	case "serial":
// 		if runtime.GOOS == "windows" {
// 			if !strings.Contains(connection, "COM") {
// 				return Sensor{}, errWindowsSerialConnection
// 			}
// 		}
// 		if runtime.GOOS == "linux" {
// 			if !strings.Contains(connection, "/dev/tty") {
// 				return Sensor{}, errLinuxSerialConnection
// 			}
// 		}
// 		portbaud := strings.Split(connection, ":")
// 		if len(portbaud) < 2 {
// 			// fmt.Printf("Port Baud %+#v\n", portbaud)
// 			return Sensor{}, errBaudRateNotgiven
// 		}
// 		port = portbaud[0]
// 		baud, err = strconv.Atoi(portbaud[1])
// 		if err != nil {
// 			return Sensor{}, errBaudRateNotanInterger
// 		}
// 		fmt.Println(baud, port)
// 	}

// 	return s, nil
// }

// Read Data from Sesor
// func (s *Sensor) Read() {}
//

//
//
//
//
//

// Actual Impeemtation of Write
// func (s *Sensor) senorWriter(b []byte) (n int, err error) {
// 	b = append(b[:0], append([]byte{startByte}, b[0:]...)...)
// 	crc, _ := checksum(&b)
// 	for k := 0; k < len(b); k++ {
// 		if b[k] == endByte {
// 			b = append(b[:k], append([]byte{escByte}, b[k:]...)...)
// 			k++
// 		}
// 	}
// 	b = append(b, byte(crc))
// 	b = append(b, endByte)
// 	if s.mode == "network" {
// 		return s.conn.Write(b)
// 	}
// 	return len(b), nil
// }

//
//
//
//
//

// // Close closes the connections to sensor
// func (s Sensor) Close() error {
// 	return s.close()
// }

// func (s Sensor) close() error {

// 	if s.mode == "network" {
// 		return s.conn.Close()

// 	}

// 	// if s.mode == "usb" {
// 	// 	return s.serial.Close()
// 	// }
// 	return errors.New("error closing connections not set")
// }

// // func (s Sensor) Read(b []byte) (n int, err error) { return s.read(b) }
// //
// // func (s Sensor) read(b []byte) (n int, err error) {
// // 	// conitional for network
// // 	return s.conn.Write(b)
// //
// // }

// func read(b []byte) (n int, err error) {

// 	return len(b), nil
// }

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
// type CRC byte

// Packet contains all the bytes
// type Packet struct {
// 	p packet
// 	d data
// 	c CRC
// }

// Checksum

// func (s Sensor) Checksum(p []byte) (CRC, error) {
// 	return checksum(p)
// }

// Calculated by XOR’ing all bytes from <START> + [Data].
// Note that the CRC is done after escape bytes is removed. This
// means that CRC is also calculated before adding escape bytes.
func checksum(p []byte) (byte, error) {
	fmt.Printf("byte to check sum %x\n", p)
	if (p)[0] != startByte {
		return 0x00, errChecksumInvalidPacketSTART
	}
	var crc byte
	for _, b := range p {
		crc = crc ^ b
	}

	return crc, nil
}

var errChecksumInvalidPacketSTART = errors.New("invalid packet missing start")

// packet implements hash.Hash Interface
// packet implements io.Writer
// packet implements io.Reader
// type xethruComm struct {
// 	s    Sensor
// 	pack []byte
// }
//
// // func (p xethruComm) Write(b []byte) (n int, err error) {}
//
// // func (p xethruComm) Read(b []byte) (n int, err error) { return 0, nil }
// func (p xethruComm) Sum(b []byte) []byte {
// 	var crc byte
// 	for _, v := range b {
// 		crc = crc ^ v
// 	}
//
// 	return []byte{crc}
// }
// func (p xethruComm) Reset()         {}
// func (p xethruComm) Size() int      { return 1 }
// func (p xethruComm) BlockSize() int { return 1 }

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
