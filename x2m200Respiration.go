package xethru

import (
	"encoding/binary"
	"errors"
	"log"
	"math"
	"time"
)

// Respiration is the struct
type Respiration struct {
	Time          int64
	Status        uint32
	Counter       uint32
	State         respirationState
	RPM           uint32
	Distance      float64
	SignalQuality float64
	Movement      float64
}

// Sleep is the struct
type Sleep struct {
	Time          int64
	Status        uint32
	Counter       uint32
	State         respirationState
	RPM           uint32
	Distance      float64
	SignalQuality float64
	MovementSlow  float64
	MovementFast  float64
}

type respirationState uint32

//go:generate jsonenums -type=respirationState
//go:generate stringer -type=respirationState
const (
	breathing      respirationState = 0
	movement       respirationState = 1
	tracking       respirationState = 2
	noMovement     respirationState = 3
	initializing   respirationState = 4
	stateReserved  respirationState = 5
	stateUnknown   respirationState = 6
	SomeotherState respirationState = 7
)

// NewModule creates
func NewModule(f Framer, mode string) *Module {
	var appID [4]byte
	parser := parseRespiration
	switch mode {
	case "respiration":
		appID = [4]byte{0xd6, 0xa2, 0x23, 0x14}
		parser = parseRespiration
	case "sleep":
		appID = [4]byte{0x00, 0xf1, 0x7b, 0x17}
	case "basebandiq":
		appID = [4]byte{0x14, 0x23, 0xa2, 0xd6}
	case "basebandampphase":
		appID = [4]byte{0x14, 0x23, 0xa2, 0xd6}
	}
	module := &Module{
		f:       f,
		AppID:   appID,
		Timeout: 500 * time.Millisecond,
		Data:    make(chan interface{}),
		parser:  parser,
	}
	module.LEDMode = LEDSimple
	module.SetLEDMode()
	return module
}

// Reset is
func (r *Module) Reset() (bool, error) {
	log.Println("Called Reset")
	return r.f.Reset(2000 * time.Millisecond)
}

type ledMode byte

//go:generate jsonenums -type=ledMode
//go:generate stringer -type=ledMode
const (
	LEDOff    ledMode = 0
	LEDSimple ledMode = 1
	LEDFull   ledMode = 2
)

const x2m200SetLEDControl = 0x24

// SetLEDMode is
// Example: <Start> + <XTS_SPC_MOD_SETLEDCONTROL> + <Mode> + <Reserved> + <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r *Module) SetLEDMode() {
	// if r.LEDMode == nil {
	// 	r.LEDMode == LEDOff
	// }
	log.Println("Setting LED MODE")
	n, err := r.f.Write([]byte{x2m200SetLEDControl, byte(r.LEDMode), 0x00})
	if err != nil {
		log.Println(err, n)
	}
	b := make([]byte, 1024)
	n, err = r.f.Read(b)
	if err != nil {
		log.Println(err, n)
	}
	if b[0] != x2m200Ack {
		log.Println("Not Ack")
	}
}

const x2m200AppCommand = 0x10
const x2m200Set = 0x10

// var x2m200DetectionZone = [4]byte{0x96, 0xa1, 0x0a, 0x1c}
var x2m200DetectionZone = [4]byte{0x1c, 0x0a, 0xa1, 0x96}

// SetDetectionZone is
// Example: <Start> + <XTS_SPC_APPCOMMAND> + <XTS_SPCA_SET> + [XTS_ID_DETECTION_ZONE(i)] + [Start(f)] + [End(f)] + <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r *Module) SetDetectionZone(start, end float64) {
	log.Printf("Setting Detection zone starting at %2.2fm ending at %2.2fm\n", start, end)

	r.DetectionZoneStart = float32(start)
	r.DetectionZoneEnd = float32(end)

	startbytes := make([]byte, 4)
	endbytes := make([]byte, 4)

	binary.LittleEndian.PutUint32(startbytes, math.Float32bits(r.DetectionZoneStart))
	binary.LittleEndian.PutUint32(endbytes, math.Float32bits(r.DetectionZoneEnd))

	n, err := r.f.Write([]byte{x2m200AppCommand, x2m200Set, x2m200DetectionZone[0], x2m200DetectionZone[1], x2m200DetectionZone[2], x2m200DetectionZone[3], startbytes[0], startbytes[1], startbytes[2], startbytes[3], endbytes[0], endbytes[1], endbytes[2], endbytes[3]})
	if err != nil {
		log.Println(err, n)
	}
	b := make([]byte, 1024)
	n, err = r.f.Read(b)
	if err != nil {
		log.Println(err, n)
	}
	if b[0] != x2m200Ack {
		log.Printf("%#02x\n", b[0:n])
		log.Println("Not Ack")
	}
}

// var x2m200Sensitivity = [4]byte{0x10, 0xa5, 0x11, 0x2b}
var x2m200Sensitivity = [4]byte{0x2b, 0x11, 0xa5, 0x10}

// SetSensitivity is
// Example: <Start> + <XTS_SPC_APPCOMMAND> + <XTS_SPCA_SET> + [XTS_ID_SENSITIVITY(i)] + [Sensitivity(i)]+ <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r *Module) SetSensitivity(sensitivity int) {

	if sensitivity > 9 {
		sensitivity = 9
	}
	if sensitivity < 0 {
		sensitivity = 0
	}

	r.Sensitivity = uint32(sensitivity)
	sensitivitybytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sensitivitybytes, r.Sensitivity)

	n, err := r.f.Write([]byte{x2m200AppCommand, x2m200Set, x2m200Sensitivity[0], x2m200Sensitivity[1], x2m200Sensitivity[2], x2m200Sensitivity[3], sensitivitybytes[0], sensitivitybytes[1], sensitivitybytes[2], sensitivitybytes[3]})
	if err != nil {
		log.Println(err, n)
	}
	b := make([]byte, 1024)
	n, err = r.f.Read(b)
	if err != nil {
		log.Println(err, n)
	}
	if b[0] != x2m200Ack {
		log.Printf("%#02x\n", b[0:n])
		log.Println("Not Ack")
	}
}

const (
	x2m200LoadModule = 0x21
	x2m200Ack        = 0x10
)

// Load is
// Example: <Start> + <XTS_SPC_MOD_LOADAPP> + [AppID(i)] + <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r *Module) Load() {
	n, err := r.f.Write([]byte{x2m200LoadModule, r.AppID[0], r.AppID[1], r.AppID[2], r.AppID[3]})
	if err != nil {
		log.Println(err, n)
	}
	b := make([]byte, 1024)
	n, err = r.f.Read(b)
	if err != nil {
		log.Println(err, n)
	}
	if b[0] != x2m200Ack {
		log.Printf("%#02x\n", b[0:n])
		log.Println("Not Ack")
	}
}

// Run start app
func (r *Module) Run() {
	defer r.f.Write([]byte{0x20, 0x11})

	n, err := r.f.Write([]byte{0x20, 0x01})
	if err != nil {
		log.Println(err, n)
	}

	for {
		b := make([]byte, 128, 256)
		n, err := r.f.Read(b)
		if err != nil {
			log.Println(err)
		}
		log.Println(b[:n], n)
		data, err := parseRespiration(b[:n])
		if err != nil {
			log.Println(err)
		}
		d := data.(Respiration)

		log.Printf("%#+v\n", d)
	}
	// defer close(r.Data)
	//
	// raw := make(chan []byte)
	// done := make(chan error)
	// defer close(raw)
	//
	// go func() {
	// 	defer close(done)
	// 	for {
	// 		b := make([]byte, 128, 256)
	// 		n, err := r.f.Read(b)
	// 		if err != nil {
	// 			done <- err
	// 			return
	// 		}
	// 		if n > 0 {
	// 			log.Printf("RAW: %#02x", b[:n])
	// 			raw <- b[:n]
	// 		}
	//
	// 	}
	// }()
	//
	// for {
	// 	select {
	// 	case b := <-raw:
	// 		log.Printf("B: %#02x", b)
	// 		d, err := r.parser(b)
	// 		if err != nil {
	// 			log.Println(err)
	// 		} else {
	// 			r.Data <- d
	// 		}
	//
	// 	case err := <-done:
	// 		log.Println(err)
	// 		return
	// 	case <-time.After(r.Timeout):
	// 		// TODO on timeout do somthing smarter
	// 		return
	// 	}
	// }

}

const (
	respirationStartByte = 0x50
)

func parseRespiration(b []byte) (interface{}, error) {
	log.Println(b)
	if b[0] != respirationStartByte {
		return Respiration{}, errParseRespDataNoResoirationByte
	}
	if len(b) < 29 {
		return Respiration{}, errParseRespDataNotEnoughBytes
	}
	data := Respiration{}
	data.Time = time.Now().UnixNano()
	data.Status = binary.LittleEndian.Uint32(b[1:5])
	data.Counter = binary.LittleEndian.Uint32(b[5:9])
	data.State = respirationState(binary.LittleEndian.Uint32(b[9:13]))
	data.RPM = binary.LittleEndian.Uint32(b[13:17])
	data.Distance = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[17:21])))
	data.Movement = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[21:25])))
	data.SignalQuality = float64(binary.LittleEndian.Uint32(b[25:29]))
	return data, nil
}

var (
	errParseRespDataNoResoirationByte = errors.New("does not start with respiration byte")
	errParseRespDataNotEnoughBytes    = errors.New("response does not contain enough bytes")
)
