package xethru

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"time"
)

type status uint32

//go:generate jsonenums -type=status
//go:generate stringer -type=status
const (
	respApp    status = 594935334
	sleepApp   status = 594911596
	basebandAP status = 0x0d
	basebandIQ status = 0x0c
)

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
	someotherState respirationState = 7
)

// NewModule creates
func NewModule(f Framer, mode string) *Module {
	var appID [4]byte
	// parser := parse
	switch mode {
	case "respiration":
		appID = [4]byte{0xd6, 0xa2, 0x23, 0x14}
		// parser = parse
	case "sleep":
		log.Println("Loading Sleep Module")
		appID = [4]byte{0x17, 0x7b, 0xf1, 0x00}
		// parser = parse
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
	}

	return module
}

// Reset is
// func (r *Module) Reset() (bool, error) {
// 	log.Println("Called Reset")
// 	return r.f.Reset()
// }

type ledMode byte

// XM200 LED Modes
//
//go:generate jsonenums -type=ledMode
//go:generate stringer -type=ledMode
const (
	LEDOff        ledMode = 0
	LEDSimple     ledMode = 1
	LEDFull       ledMode = 2
	LEDInhalation ledMode = 3
)

const x2m200SetLEDControl = 0x24

// SetLEDMode is
// Example: <Start> + <XTS_SPC_MOD_SETLEDCONTROL> + <Mode> + <Reserved> + <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r *Module) SetLEDMode() error {
	// if r.LEDMode == nil {
	// 	r.LEDMode == LEDOff
	// }
	log.Println("Setting LED MODE")
	n, err := r.f.Write([]byte{x2m200SetLEDControl, byte(r.LEDMode), 0x00})
	if err != nil {
		return err
	}

	attempts := 0
reRead:
	b := make([]byte, 2048)
	n, err = r.f.Read(b)
	if err != nil {
		log.Printf("Reset read Error %v, number of bytes %d\n", err, n)
		return err
	}
	state, err := parse(b[:n])
	if err != nil {
		log.Printf("Parse read Error %v, state %#+v \n", err, state)
		return err
	}
	// log.Printf("Debug state %#+v \n", state)
	switch state.(type) {
	case SystemMessage:
		s := state.(SystemMessage)
		if s.Message == "Command Ack'ed" {
			return nil
		}
	default:
		if attempts < 20 {
			attempts++
			goto reRead
		}

	}
	return fmt.Errorf("failed to set led mode\n")

	// b := make([]byte, 1024)
	// n, err = r.f.Read(b)
	// if err != nil {
	// 	log.Println(err, n)
	// }
	// if b[0] != x2m200Ack {
	// 	log.Println("Not Ack")
	// }
}

const (
	x2m200AppCommand = 0x10
	x2m200Set        = 0x10
)

var x2m200DetectionZone = [4]byte{0x96, 0xa1, 0x0a, 0x1c}

// var x2m200DetectionZone = [4]byte{0x1c, 0x0a, 0xa1, 0x96}

// SetDetectionZone is
// Example: <Start> + <XTS_SPC_APPCOMMAND> + <XTS_SPCA_SET> + [XTS_ID_DETECTION_ZONE(i)] + [Start(f)] + [End(f)] + <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r Module) SetDetectionZone(start, end float64) error {
	log.Printf("Setting Detection zone starting at %2.2fm ending at %2.2fm\n", start, end)

	r.DetectionZoneStart = float32(start)
	r.DetectionZoneEnd = float32(end)

	startbytes := make([]byte, 4)
	endbytes := make([]byte, 4)

	binary.LittleEndian.PutUint32(startbytes, math.Float32bits(r.DetectionZoneStart))
	binary.LittleEndian.PutUint32(endbytes, math.Float32bits(r.DetectionZoneEnd))

	// n, err := r.f.Write([]byte{x2m200AppCommand, x2m200Set, x2m200DetectionZone[0], x2m200DetectionZone[1], x2m200DetectionZone[2], x2m200DetectionZone[3], startbytes[0], startbytes[1], startbytes[2], startbytes[3], endbytes[0], endbytes[1], endbytes[2], endbytes[3]})

	n, err := r.f.Write([]byte{0x10, 0x10, 0x1c, 0x0a, 0xa1, 0x96, startbytes[0], startbytes[1], startbytes[2], startbytes[3], endbytes[0], endbytes[1], endbytes[2], endbytes[3]})
	if err != nil {
		return err
		// log.Println(err, n)
	}

	attempts := 0
reRead:
	b := make([]byte, 2048)
	n, err = r.f.Read(b)
	if err != nil {
		log.Printf("Reset read Error %v, number of bytes %d\n", err, n)
		return err
	}
	state, err := parse(b[:n])
	if err != nil {
		log.Printf("Parse read Error %v, state %#+v \n", err, state)
		return err
	}
	// log.Printf("Debug state %#+v \n", state)
	switch state.(type) {
	case SystemMessage:
		s := state.(SystemMessage)
		if s.Message == "Command Ack'ed" {
			return nil
		}
	default:
		if attempts < 20 {
			attempts++
			goto reRead
		}

	}
	return fmt.Errorf("failed to set detection zone %2.2f %2.2f\n", start, end)
}

// b := make([]byte, 1024)
// n, err = r.f.Read(b)
// if err != nil {
// 	log.Println(err, n)
// }
// if b[0] != x2m200Ack {
// 	log.Printf("%#02x\n", b[0:n])
// 	log.Println("Not Ack")
// }
// }

// var x2m200Sensitivity = [4]byte{0x10, 0xa5, 0x11, 0x2b}
var x2m200Sensitivity = [4]byte{0x2b, 0x11, 0xa5, 0x10}

// SetSensitivity is
// Example: <Start> + <XTS_SPC_APPCOMMAND> + <XTS_SPCA_SET> + [XTS_ID_SENSITIVITY(i)] + [Sensitivity(i)]+ <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r Module) SetSensitivity(sensitivity int) error {

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
	attempts := 0
reRead:
	b := make([]byte, 2048)
	n, err = r.f.Read(b)
	if err != nil {
		log.Printf("Reset read Error %v, number of bytes %d\n", err, n)
		return err
	}
	state, err := parse(b[:n])
	if err != nil {
		log.Printf("Parse read Error %v, state %#+v \n", err, state)
		return err
	}
	// log.Printf("Debug state %#+v \n", state)
	switch state.(type) {
	case SystemMessage:
		s := state.(SystemMessage)
		if s.Message == "Command Ack'ed" {
			return nil
		}
	default:
		if attempts < 20 {
			attempts++
			goto reRead
		}

	}
	return fmt.Errorf("failed to set sensitivity %d\n", sensitivity)
}

const (
	x2m200LoadModule = 0x21
	x2m200Ack        = 0x10
)

// Load is
// Example: <Start> + <XTS_SPC_MOD_LOADAPP> + [AppID(i)] + <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r Module) Load() error {
load:
	n, err := r.f.Write([]byte{x2m200LoadModule, r.AppID[0], r.AppID[1], r.AppID[2], r.AppID[3]})
	if err != nil {
		log.Println(err, n)
		return err
	}
	attempts := 0
reRead:
	b := make([]byte, 2048)
	n, err = r.f.Read(b)
	if err != nil {
		log.Printf("Reset read Error %v, number of bytes %d\n", err, n)
		return err
	}
	state, err := parse(b[:n])
	if err != nil {
		log.Printf("Parse read Error %v, state %#+v \n", err, state)
		return err
	}
	// log.Printf("Debug state %#+v \n", state)
	switch state.(type) {
	case SystemMessage:
		s := state.(SystemMessage)
		if s.Message == "Command Ack'ed" {
			return nil
		}
		if s.Message == "System Still booting" {
			if attempts < 20 {
				goto reRead
			}
		}
		if s.Message == "System Ready" {
			if attempts < 20 {
				goto load
			}
		}
		// return x.Reset()
	default:
		if attempts < 20 {
			goto reRead
		}

	}

	return fmt.Errorf("did not recive ack for load module\n")
}

// <Start> + <XTS_SPC_DIR_COMMAND> + <XTS_SDC_APP_SETINT> + [XTS_SACR_OUTPUTBASEBAND(i)] + [Length(i)] + [EnableCode(i)] + <CRC> + <End> Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r Module) Enable(mode string) error {
	switch mode {
	case "phase":
		log.Println("Enable Phase Amp Baseband")
		n, err := r.f.Write([]byte{0x90, 0x71, 0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00})

		if err != nil {
			log.Println(err, n)
			return err
		}
	case "iq":
		log.Println("Enable IQ Baseband")

		n, err := r.f.Write([]byte{0x90, 0x71, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00})

		if err != nil {
			log.Println(err, n)
			return err
		}
	default:
		log.Println("Disable Baseband")

		n, err := r.f.Write([]byte{0x90, 0x71, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		if err != nil {
			log.Println(err, n)
			return err
		}
	}

	attempts := 0
reRead:
	b := make([]byte, 2048)
	n, err := r.f.Read(b)
	if err != nil {
		log.Printf("Reset read Error %v, number of bytes %d\n", err, n)
		return err
	}
	state, err := parse(b[:n])
	if err != nil {
		log.Printf("Parse read Error %v, state %#+v \n", err, state)
		return err
	}
	// log.Printf("Debug state %#+v \n", state)
	switch state.(type) {
	case SystemMessage:
		s := state.(SystemMessage)
		if s.Message == "Command Ack'ed" {
			return nil
		}
	default:
		if attempts < 20 {
			attempts++
			goto reRead
		}

	}
	return fmt.Errorf("failed to set Enable %s mode", mode)
}

// 	b := make([]byte, 1024)
// 	n, err := r.f.Read(b)
// 	if err != nil {
// 		log.Println(err, n)
// 		return err
// 	}
// 	if b[0] != x2m200Ack {
// 		log.Printf("%#02x\n", b[0:n])
// 		log.Println("Not Ack")
// 		return errors.New("error enable Phase Amp Baseband was not ack'ed")
// 	}
// 	return nil
// }

// Run start app
func (r Module) Run(stream chan interface{}) {
	defer r.f.Write([]byte{0x20, 0x11})

	n, err := r.f.Write([]byte{0x20, 0x01})
	if err != nil {
		log.Println(err, n)
	}

	output := make(chan []byte, 1000)

	go func(out chan []byte) {
		for {
			b := make([]byte, 2048)
			n, err := r.f.Read(b)
			if err != nil {
				log.Println(err)
			}
			out <- b[:n]
		}
	}(output)

	for {
		select {
		case out := <-output:
			data, err := parse(out)
			if err != nil {
				log.Println(err)
			}
			stream <- data
		}
	}
}
