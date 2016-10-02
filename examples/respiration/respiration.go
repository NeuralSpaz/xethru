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

// example usage of the basic xethru protocol respiration app
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/NeuralSpaz/xethru"
	"github.com/gorilla/websocket"
	// "github.com/jacobsa/go-serial/serial"
	"github.com/tarm/serial"
)

func main() {
	log.Println("X2M200 Respiration Demo")

	commPort := flag.String("commPort", "/dev/ttyACM0", "the comm port you wish to use")
	baudrate := flag.Int("baudrate", 921600, "the baud rate for the comm port you wish to use")
	// pingTimeout := flag.Duration("pingTimeout", time.Millisecond*500, "timeout for ping command")
	flag.Parse()

	baseband := make(chan xethru.BaseBandAmpPhase)

	go openXethru(*commPort, *baudrate, baseband)
	connections = make(map[*websocket.Conn]bool)
	http.HandleFunc("/ws", wsHandler)

	// http.HandleFunc("/", indexHandler)
	http.Handle("/", http.FileServer(http.Dir("./www")))

	go func() {
		err := http.ListenAndServe("0.0.0.0:23000", nil)
		if err != nil {
			log.Println(err)
		}
	}()

	for {
		select {
		case data := <-baseband:
			b, err := json.Marshal(data)
			if err != nil {
				log.Panicln("Error Marshaling: ", err)
			}
			sendAll(b)
		}
	}

}

func sendAll(msg []byte) {
	connectMutex.Lock()
	for conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(connections, conn)
			conn.Close()
		}
	}
	connectMutex.Unlock()
}

// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	file, err := ioutil.ReadFile("index.html")
// 	// defer file.Close()
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	// ioutil.ReadFile(filename)
// 	// file.Read(b)
// 	// file, _ := Asset("index.html")
// 	w.Header().Set("Content-Type", "text/html")
// 	w.Write(file)
// }

var connectMutex sync.Mutex
var connections map[*websocket.Conn]bool

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Taken from gorilla's website
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	log.Println("Succesfully upgraded connection")
	connectMutex.Lock()
	connections[conn] = true
	connectMutex.Unlock()

	for {
		// Blocks until a message is read
		_, msg, err := conn.ReadMessage()
		if err != nil {
			connectMutex.Lock()
			log.Printf("Disconnecting %v because %v\n", conn, err)
			delete(connections, conn)
			connectMutex.Unlock()
			conn.Close()
			return
		}
		log.Println(msg)
	}
}

func openXethru(comm string, baudrate int, baseband chan xethru.BaseBandAmpPhase) {
	online, err := exists(comm)
	if err != nil {
		log.Fatal(err)
	}
	for !online {
		time.Sleep(time.Millisecond * 2000)
		online, err = exists(comm)
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(time.Millisecond * 5000)

	c := &serial.Config{Name: comm, Baud: baudrate}
	port, err := serial.OpenPort(c)

	// options := serial.OpenOptions{
	// 	PortName:        comm,
	// 	BaudRate:        baudrate,
	// 	DataBits:        8,
	// 	StopBits:        1,
	// 	MinimumReadSize: 4,
	// }
	//
	// port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	port.Flush()

	x2 := xethru.Open("x2m200", port)

	reset, err := x2.Reset()
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	log.Println(reset)
	port.Close()
	time.Sleep(time.Millisecond * 5000)

	online, err = exists(comm)
	if err != nil {
		log.Fatal(err)
	}
	for !online {
		time.Sleep(time.Millisecond * 2000)
		online, err = exists(comm)
		if err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(time.Millisecond * 10000)

	// port, err = serial.Open(options)
	port, err = serial.OpenPort(c)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	time.Sleep(time.Second * 1)
	port.Flush()

	defer port.Close()
	x2 = xethru.Open("x2m200", port)

	m := xethru.NewModule(x2, "sleep")

	log.Printf("%#+v\n", m)
	err = m.Load()
	if err != nil {
		log.Panicln(err)
	}

	log.Println("Setting LED MODE")
	m.LEDMode = xethru.LEDInhalation
	m.SetLEDMode()
	time.Sleep(time.Second * 1)

	log.Println("SetDetectionZone")
	m.SetDetectionZone(0.5, 2.1)
	time.Sleep(time.Second * 10)

	log.Println("SetSensitivity")
	m.SetSensitivity(9)
	time.Sleep(time.Second * 1)
	m.Enable("phase")
	// time.Sleep(time.Second * 5)
	stream := make(chan interface{})
	go m.Run(stream)

	// basebandfile, err := os.Create("./basebanddata.json")
	// if err != nil {
	// 	log.Panic(err)
	// }
	// defer basebandfile.Close()
	respirationfile, err := os.Create("./respiration.json")
	if err != nil {
		log.Panic(err)
	}
	defer respirationfile.Close()

	// basebandenc := json.NewEncoder(basebandfile)
	respirationenc := json.NewEncoder(respirationfile)
	// respirationscreenenc := json.NewEncoder(os.Stdout)

	// frameCounter := 0
	for {
		select {
		case s := <-stream:
			switch s.(type) {
			case xethru.Respiration:
				// if err := respirationscreenenc.Encode(&s); err != nil {
				// 	log.Println(err)
				// }
				if err := respirationenc.Encode(&s); err != nil {
					log.Println(err)
				}
			case xethru.BaseBandAmpPhase:
				baseband <- s.(xethru.BaseBandAmpPhase)
				// if err := basebandenc.Encode(&s); err != nil {
				// 	log.Println(err)
				// }

			}
		}
		// if frameCounter%100 == 0 {
		// basebandfile.Sync()
		// respirationfile.Sync()
		// }
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
