package xethru

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"testing"
)

func TestXethruWrite(t *testing.T) {

	cases := []struct {
		b      []byte
		n      int
		err    error
		writen []byte
	}{
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
		read  []byte
		err   error
		write []byte
	}{
		{[]byte{0x01, 0x02, 0x03}, nil, []byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x00, 0x01, 0x02, 0x03}, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
		{[]byte{0x00, 0x01, 0x02, 0x7e}, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x7f, 0x7e, 0x00, 0x7e}},
		{[]byte{0x7e, 0x01, 0x02, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x01, 0x02, 0x7f, 0x7e, 0x7e, 0x7e}},
		{[]byte{0x7e, 0x7e, 0x02, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x02, 0x7f, 0x7e, 0x01, 0x7e}},
		{[]byte{0x7e, 0x7e, 0x7e, 0x7e}, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7d, 0x7e}},
	}

	for k, c := range cases {
		r := bytes.NewReader(c.write)
		x := NewXethruReader(r)

		log.Println("Running Test, ", k)
		read, err := ioutil.ReadAll(x)

		if err != nil {
			log.Println(err)
		}

		if string(read) != string(c.read) {
			t.Errorf("Expected: %x, got %x\n", c.read, read)
		}

	}

	//
	// var b bytes.Buffer
	// b.Write([]byte{0x01, 0x02, 0x03})
	// r := bufio.NewReader(&b)
	// r := bytes.NewReader([]byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7e})
	// x := NewXethruReader(r)
	// // buf := new(bytes.Buffer)
	// // _, err := buf.ReadFrom(x)
	// //
	// //
	// // longBuf := make([]byte, 20)
	// // n, err := x.Read(longBuf)
	// //
	// //
	// read, err := ioutil.ReadAll(x)
	// //
	// //
	// if err != nil {
	// 	fmt.Printf("error: %+#v", err)
	// }
	// // fmt.Println("lines read:", n)
	// fmt.Printf("bytes: %x\n", longBuf)
	// respByte, err := ioutil.ReadAll(x)
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println("Read: ", buf)

}

// func TestImlements(t *testing.T) {
// 	packetHash := types.Implements((*packet)(nil), hash.Hash)
// 	fmt.Println(packetHash)
//
// }

// func TestNewSensor(t *testing.T) {
// 	if runtime.GOOS == "linux" {
// 		l, err := net.Listen("tcp", ":30000")
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer l.Close()
// 		cases := []struct {
// 			transport, connection string
// 			err                   error
// 		}{
// 			{"tcp", "localhost:30000", nil},
// 			{"udp", "localhost:30000", nil},
// 			{"serial", "/dev/ttyUSB0:9600", nil},
// 			{"serial", "/dev/ttyUSB0", errBaudRateNotgiven},
// 			{"serial", "/dev/ttyUSB0:junk", errBaudRateNotanInterger},
// 			{"serial", "COM10:9600", errLinuxSerialConnection},
// 			{"junk", "morejunk", errUnsupportedTransport},
// 		}
// 		for _, c := range cases {
// 			_, err := NewSensor(c.transport, c.connection)
// 			if err != c.err {
// 				t.Errorf("could not create sensor transport: %v, connection: %v; got %+#v wanted %+#v\n", c.transport, c.connection, err, c.err)
// 			}
// 		}
// 		netcases := []struct {
// 			transport, connection string
// 			err                   string
// 		}{
// 			{"udp", "192.168.1.18080", "missing port in address 192.168.1.18080"},
// 			{"tcp", "192.168.1.18080", "missing port in address 192.168.1.18080"},
// 		}
// 		for _, c := range netcases {
// 			_, err := NewSensor(c.transport, c.connection)
// 			strerror := fmt.Sprintf("%s", err)
// 			if strerror != c.err {
// 				t.Errorf("could not create sensor transport: %v, connection: %v; got %+#v wanted %+#v\n", c.transport, c.connection, strerror, c.err)
// 			}
// 		}
// 	}
// }
//
// func TestChecksum(t *testing.T) {
// 	cases := []struct {
// 		p   []byte
// 		crc CRC
// 		err error
// 	}{
// 		{[]byte{0x00, 0x01, 0x02}, 0x00, errChecksumInvalidPacketSTART},
// 		{[]byte{startByte, 0x01, 0x02}, 0x7E, nil},
// 		{[]byte{startByte, 0x01, 0x02, 0x03}, 0x7D, nil},
// 		{[]byte{startByte, 0x01, 0x02, 0xFF}, 0x81, nil},
// 		{[]byte{startByte, 0x01, 0x02, 0x7F}, 0x01, nil},
// 	}
// 	for _, c := range cases {
// 		crc, err := checksum(&c.p)
// 		if err != c.err {
// 			t.Errorf("Expected: %v, got %v\n", c.err, err)
// 		}
// 		// if valid != c.valid {
// 		// 	t.Errorf("Expected: %v, got %v\n", c.valid, valid)
// 		// }
// 		if crc != c.crc {
// 			t.Errorf("Expected: %X, got %X\n", c.crc, crc)
// 		}
// 	}
// }
//
// func TestWriteTCP(t *testing.T) {
//
// 	cases := []struct {
// 		b       []byte
// 		n       int
// 		err     error
// 		network []byte
// 	}{
// 		{[]byte{0x01, 0x02, 0x03}, 6, nil, []byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
// 		{[]byte{0x00, 0x01, 0x02, 0x03}, 7, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
// 		{[]byte{0x00, 0x01, 0x02, 0x7e}, 8, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x7f, 0x7e, 0x00, 0x7e}},
// 		{[]byte{0x7e, 0x01, 0x02, 0x7e}, 9, nil, []byte{0x7d, 0x7f, 0x7e, 0x01, 0x02, 0x7f, 0x7e, 0x7e, 0x7e}},
// 		{[]byte{0x7e, 0x7e, 0x02, 0x7e}, 10, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x02, 0x7f, 0x7e, 0x01, 0x7e}},
// 		{[]byte{0x7e, 0x7e, 0x7e, 0x7e}, 11, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7d, 0x7e}},
// 	}
//
// 	for k, c := range cases {
// 		port := strconv.Itoa(k)
// 		l, err := net.Listen("tcp", ":3000"+port)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer l.Close()
//
// 		s, err := NewSensor("tcp", "localhost:3000"+port)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer s.Close()
//
// 		conn, err := l.Accept()
// 		if err != nil {
// 			return
// 		}
// 		n, err := s.Write(c.b)
// 		if err != c.err {
// 			t.Errorf("Expected: %v, got %v\n", c.err, err)
// 		}
// 		if n != c.n {
// 			t.Errorf("Expected: %d, got %d\n", c.n, n)
// 		}
//
// 		defer conn.Close()
// 		buf := bufio.NewReader(conn)
// 		var message []byte
// 		for i := 0; i < n; i++ {
// 			b, err := buf.ReadByte()
// 			if err != nil {
// 				log.Println(err)
// 			}
// 			message = append(message, b)
//
// 		}
// 		if string(message) != string(c.network) {
// 			t.Errorf("Expected: %x, got %x\n", c.network, message)
// 		}
// 	}
//
// }
//
// func TestWriteUDP(t *testing.T) {
//
// 	cases := []struct {
// 		b       []byte
// 		n       int
// 		err     error
// 		network []byte
// 	}{
// 		{[]byte{0x01, 0x02, 0x03}, 6, nil, []byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
// 		{[]byte{0x00, 0x01, 0x02, 0x03}, 7, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x7d, 0x7e}},
// 		{[]byte{0x00, 0x01, 0x02, 0x7e}, 8, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x7f, 0x7e, 0x00, 0x7e}},
// 		{[]byte{0x7e, 0x01, 0x02, 0x7e}, 9, nil, []byte{0x7d, 0x7f, 0x7e, 0x01, 0x02, 0x7f, 0x7e, 0x7e, 0x7e}},
// 		{[]byte{0x7e, 0x7e, 0x02, 0x7e}, 10, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x02, 0x7f, 0x7e, 0x01, 0x7e}},
// 		{[]byte{0x7e, 0x7e, 0x7e, 0x7e}, 11, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7d, 0x7e}},
// 	}
//
// 	for k, c := range cases {
// 		// fmt.Println(k)
// 		port := strconv.Itoa(k)
// 		udpaddr, err := net.ResolveUDPAddr("udp", "localhost:6000"+port)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		conn, err := net.ListenUDP("udp", udpaddr)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer conn.Close()
//
// 		s, err := NewSensor("udp", "localhost:6000"+port)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer s.Close()
//
// 		// nw := bufio.NewWriter(s)
//
// 		// writelogger := iotest.NewWriteLogger("write", s)
// 		// // w := bufio.NewWriter(writelogger)
// 		// n, err := writelogger.Write(c.b)
// 		// log.Println(n)
// 		n, err := s.Write(c.b)
// 		if err != c.err {
// 			t.Errorf("Expected: %v, got %v\n", c.err, err)
// 		}
// 		if n != c.n {
// 			t.Errorf("Expected: %d, got %d\n", c.n, n)
// 		}
// 		// log.Println("we Got here")
// 		// readlogger := iotest.NewReadLogger("read", conn)
// 		// buf := bufio.NewReader(readlogger)
// 		buf := bufio.NewReader(conn)
// 		var message []byte
// 		// log.Println(n)
// 		for i := 0; i < n; i++ {
// 			b, err := buf.ReadByte()
// 			if err != nil {
// 				log.Println(err)
// 			}
// 			message = append(message, b)
// 			// fmt.Println(message)
//
// 		}
// 		if string(message) != string(c.network) {
// 			t.Errorf("Expected: %x, got %x\n", c.network, message)
// 		}
// 	}
// }

// func TestReadTCP(t *testing.T) {
//
// 	cases := []struct {
// 		b       []byte
// 		n       int
// 		err     error
// 		network []byte
// 		r       *io.Reader
// 	}{
// 		{[]byte{0x01, 0x02, 0x03}, 6, nil, []byte{0x7d, 0x01, 0x02, 0x03, 0x7d, 0x7e}, nil},
// 		// {[]byte{0x00, 0x01, 0x02, 0x03}, 7, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x7d, 0x7e}, nil},
// 		// {[]byte{0x00, 0x01, 0x02, 0x7e}, 8, nil, []byte{0x7d, 0x00, 0x01, 0x02, 0x7f, 0x7e, 0x00, 0x7e}, nil},
// 		// {[]byte{0x7e, 0x01, 0x02, 0x7e}, 9, nil, []byte{0x7d, 0x7f, 0x7e, 0x01, 0x02, 0x7f, 0x7e, 0x7e, 0x7e}, nil},
// 		// {[]byte{0x7e, 0x7e, 0x02, 0x7e}, 10, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x02, 0x7f, 0x7e, 0x01, 0x7e}, nil},
// 		// {[]byte{0x7e, 0x7e, 0x7e, 0x7e}, 11, nil, []byte{0x7d, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7f, 0x7e, 0x7d, 0x7e}, nil},
// 	}
//
// 	for k, c := range cases {
// 		fmt.Println(c, k)
// 		rl := iotest.NewReadLogger("testing", c.r)
// 	}
//
// }
