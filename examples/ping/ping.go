// Copyright © 2016 Josh Gardiner aka NeuralSpaz on github.com
//
// This file is part of Xethru-Go - A Golang library for the xethru modules
//
package main

import (
	"flag"
	"log"
	"time"

	"github.com/NeuralSpaz/xethru"
	"github.com/jacobsa/go-serial/serial"
)

func main() {
	log.Println("X2M200 Ping Demo")
	commPort := flag.String("commPort", "/dev/ttyUSB0", "the comm port you wish to use")
	baudrate := flag.Uint("baudrate", 115200, "the baud rate for the comm port you wish to use")
	pingTimeout := flag.Duration("pingTimeout", time.Millisecond*300, "timeout for ping command")
	flag.Parse()

	options := serial.OpenOptions{
		PortName:        *commPort,
		BaudRate:        *baudrate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	defer port.Close()
	x2 := xethru.Open(port)

	for i := 0; i < 10; i++ {
		ok, err := x2.Ping(*pingTimeout)
		if err != nil {
			log.Fatalf("Error Communicating with Device: %v", err)
		}
		if !ok {
			log.Fatal("Device Not Ready")
		}
		log.Println("Got Pong")

		time.Sleep(*pingTimeout)
	}

	//
	// appconfig := xetheu.AppConfig{
	// 	Name:        "Resp",
	// 	ZoneStart:   0.5,
	// 	ZoneEnd:     1.5,
	// 	LEDMode:     Full,
	// 	Sensitivity: 10,
	// 	Output:      os.Stdout,
	// }
	//
	// app, err := x2.LoadApp(appconfig)
	// if err != nil {
	// 	log.Fatalf("Error Loading App: %v", err)
	// }
	// app.Start()
}
