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

// example usage of the basic xethru protocol
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/NeuralSpaz/xethru"
	// "github.com/jacobsa/go-serial/serial"
	"github.com/tarm/serial"
)

func main() {
	log.Println("X2M200 Respiration Demo")

	commPort := flag.String("commPort", "/dev/ttyACM0", "the comm port you wish to use")
	baudrate := flag.Int("baudrate", 115200, "the baud rate for the comm port you wish to use")
	// pingTimeout := flag.Duration("pingTimeout", time.Millisecond*500, "timeout for ping command")
	flag.Parse()
	log.Println("Waiting for Device to be ready")

	online, err := exists(*commPort)
	if err != nil {
		log.Fatal(err)
	}
	for !online {
		time.Sleep(time.Millisecond * 2000)
		online, err = exists(*commPort)
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(time.Millisecond * 5000)

	c := &serial.Config{Name: *commPort, Baud: *baudrate}
	port, err := serial.OpenPort(c)

	// options := serial.OpenOptions{
	// 	PortName:        *commPort,
	// 	BaudRate:        *baudrate,
	// 	DataBits:        8,
	// 	StopBits:        1,
	// 	MinimumReadSize: 4,
	// }

	// port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	x2 := xethru.Open("x2m200", port)

	reset, err := x2.Reset()
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	log.Println(reset)
	port.Close()
	time.Sleep(time.Millisecond * 5000)

	online, err = exists(*commPort)
	if err != nil {
		log.Fatal(err)
	}
	for !online {
		time.Sleep(time.Millisecond * 2000)
		online, err = exists(*commPort)
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

	// defer port.Close()
	x2 = xethru.Open("x2m200", port)

	log.Println("strting Sleep Mode")
	m := xethru.NewModule(x2, "sleep")
	m.AppID = [4]byte{0x17, 0x7b, 0xf1, 0x00}
	// const unsigned long xts_id_app_sleep = 0x 00 f1 7b 17;
	log.Printf("%#+v\n", m)
	err = m.Load()
	if err != nil {
		log.Fatalln("error loading Module ", err)
	}

	log.Println("Setting LED MODE")
	m.LEDMode = xethru.LEDFull
	m.SetLEDMode()
	time.Sleep(time.Second * 1)

	log.Println("SetDetectionZone")
	m.SetDetectionZone(1.8, 2.5)
	time.Sleep(time.Second * 1)

	log.Println("SetSensitivity")
	m.SetSensitivity(9)
	time.Sleep(time.Second * 5)

	m.Run()

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
