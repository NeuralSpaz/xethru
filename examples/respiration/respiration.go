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
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/NeuralSpaz/xethru"
	"github.com/gorilla/websocket"
	"github.com/jacobsa/go-serial/serial"
	"github.com/mjibson/go-dsp/fft"
	// "github.com/tarm/serial"
)

func main() {
	log.Println("X2M200 Respiration Demo")

	commPort := flag.String("commPort", "/dev/ttyACM0", "the comm port you wish to use")
	baudrate := flag.Uint("baudrate", 115200, "the baud rate for the comm port you wish to use")
	// pingTimeout := flag.Duration("pingTimeout", time.Millisecond*500, "timeout for ping command")
	flag.Parse()

	baseband := make(chan xethru.BaseBandAmpPhase)
	resp := make(chan xethru.Respiration)
	// time.Sleep(time.Second * 5)
	go openXethru(*commPort, *baudrate, baseband, resp)
	baseBandconnections = make(map[*websocket.Conn]bool)
	respirationconnections = make(map[*websocket.Conn]bool)
	http.HandleFunc("/ws/bb", baseBandwsHandler)
	http.HandleFunc("/ws/r", respirationwsHandler)

	http.Handle("/", http.FileServer(http.Dir("./www")))
	// http.HandleFunc("/", indexHandler)
	// http.HandleFunc("/js/reconnecting-websocket.min.js", websocketReconnectHandler)

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
			sendBaseBand(b)
			// go addtoamp(data)

		case data := <-resp:
			b, err := json.Marshal(data)
			if err != nil {
				log.Panicln("Error Marshaling: ", err)
			}
			sendrespiration(b)
		}
	}
}

var amp []float64

func addtoamp(a xethru.BaseBandAmpPhase) {
	if len(amp) > 3120 {
		amp = amp[1:]
	}
	amp = append(amp, a.Amplitude[1])
	x := fft.FFTReal(amp)
	log.Println(len(amp), x)
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	file, _ := Asset("www/index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(file)
}
func websocketReconnectHandler(w http.ResponseWriter, r *http.Request) {
	file, _ := Asset("www/js/reconnecting-websocket.min.js")
	w.Header().Set("Content-Type", "text/javascript")
	w.Write(file)
}

var respirationconnectionsMutex sync.Mutex
var respirationconnections map[*websocket.Conn]bool

func respirationwsHandler(w http.ResponseWriter, r *http.Request) {
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
	respirationconnectionsMutex.Lock()
	respirationconnections[conn] = true
	respirationconnectionsMutex.Unlock()

	for {
		// Blocks until a message is read
		_, msg, err := conn.ReadMessage()
		if err != nil {
			respirationconnectionsMutex.Lock()
			// log.Printf("Disconnecting %v because %v\n", conn, err)
			delete(respirationconnections, conn)
			respirationconnectionsMutex.Unlock()
			conn.Close()
			return
		}
		log.Println(msg)
	}
}

func sendrespiration(msg []byte) {
	respirationconnectionsMutex.Lock()
	for conn := range respirationconnections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(respirationconnections, conn)
			conn.Close()
		}
	}
	respirationconnectionsMutex.Unlock()
}

var baseBandconnectionsMutex sync.Mutex
var baseBandconnections map[*websocket.Conn]bool

func baseBandwsHandler(w http.ResponseWriter, r *http.Request) {
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
	baseBandconnectionsMutex.Lock()
	baseBandconnections[conn] = true
	baseBandconnectionsMutex.Unlock()

	for {
		// Blocks until a message is read
		_, msg, err := conn.ReadMessage()
		if err != nil {
			baseBandconnectionsMutex.Lock()
			// log.Printf("Disconnecting %v because %v\n", conn, err)
			delete(baseBandconnections, conn)
			baseBandconnectionsMutex.Unlock()
			conn.Close()
			return
		}
		log.Println(msg)
	}
}

func sendBaseBand(msg []byte) {
	baseBandconnectionsMutex.Lock()
	for conn := range baseBandconnections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(baseBandconnections, conn)
			conn.Close()
		}
	}
	baseBandconnectionsMutex.Unlock()
}

func openXethru(comm string, baudrate uint, baseband chan xethru.BaseBandAmpPhase, resp chan xethru.Respiration) {

	time.Sleep(time.Millisecond * 2000)

	options := serial.OpenOptions{
		PortName:        comm,
		BaudRate:        baudrate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Printf("serial.Open: %v\n", err)
	}

	x2 := xethru.Open("x2m200", port)

	reset, err := x2.Reset()
	if err != nil {
		log.Printf("serial.Reset: %v\n", err)
	}
	log.Println(reset)
	port.Close()
	// time.Sleep(time.Millisecond * 5000)

	count := 20
	for {
		select {
		case <-time.After(time.Second):
			count--
			log.Println(count)
		}
		if count == 0 {
			break
		}
	}
	// time.Sleep(time.Millisecond * 20000)

	port, err = serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	time.Sleep(time.Second * 1)

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

	log.Println("SetDetectionZone")
	m.SetDetectionZone(0.5, 2.1)

	log.Println("SetSensitivity")
	m.SetSensitivity(7)
	m.Enable("phase")
	stream := make(chan interface{})
	go m.Run(stream)

	// open default browser
	open("http://localhost:23000/")

	respirationfile, err := os.Create("./respiration.json")
	if err != nil {
		log.Panic(err)
	}
	defer respirationfile.Close()

	sleepfile, err := os.Create("./sleep.json")
	if err != nil {
		log.Panic(err)
	}
	defer sleepfile.Close()

	sleepenc := json.NewEncoder(sleepfile)
	respirationenc := json.NewEncoder(respirationfile)
	for {
		select {
		case s := <-stream:
			switch s.(type) {
			case xethru.Respiration:
				resp <- s.(xethru.Respiration)
				if err := respirationenc.Encode(&s); err != nil {
					log.Println(err)
				}
			case xethru.BaseBandAmpPhase:
				baseband <- s.(xethru.BaseBandAmpPhase)
			case xethru.Sleep:
				s = s.(xethru.Sleep)
				if err := sleepenc.Encode(&s); err != nil {
					log.Println(err)
				}
			default:
				log.Printf("%#v", s)
			}

		}
	}
}

type kalman struct {
	g  float64 // gain
	em float64 // error in measument
	ee float64 // error in estimate
	e  float64 // estimate
	le float64 // lastestimate
}

func (k *kalman) setMeasumentError(em float64) {
	k.em = em
}

func (k *kalman) kalmanFilter(m float64) float64 {
	k.g = k.ee / (k.ee + k.em)
	k.e = k.le + k.g*(m-k.le)
	k.ee = (1 - k.g) * k.ee
	return k.e
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
