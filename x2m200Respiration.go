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
	State         resperationState
	RPM           uint32
	Distance      float64
	SignalQuality float64
	Movement      float64
}

// RespirationModule Config
type RespirationModule struct {
	f             Framer
	AppID         [4]byte
	LEDMode       uint32
	DetectionZone uint32
	Sensitivity   uint32
	Timeout       time.Duration
	data          chan Respiration
}

type resperationState uint32

const (
	resperationStateBreathing    resperationState = 0
	resperationStateMovement     resperationState = 1
	resperationStateTracking     resperationState = 2
	resperationStateNoMovement   resperationState = 3
	resperationStateInitializing resperationState = 4
	resperationStateReserved     resperationState = 5
	resperationStateUnknown      resperationState = 6
)

func (s resperationState) String() string {
	switch s {
	case 0:
		return "Breathing"
	case 1:
		return "Movement"
	case 2:
		return "Tracking"
	case 3:
		return "No Movement"
	case 4:
		return "Initializing"
	case 5:
		return "Reserved"
	case 6:
		return "Unknown"
	default:
		return "NotValid"
	}
}

// New creates and loads app
func New(f Framer) *RespirationModule {
	module := &RespirationModule{
		f:       f,
		AppID:   [4]byte{0x14, 0x23, 0xa2, 0xd6},
		Timeout: 500 * time.Millisecond,
		data:    make(chan Respiration),
	}
	return module
}

// Run start app
func (r *RespirationModule) Run() {
	defer close(r.data)

	raw := make(chan []byte)
	done := make(chan error)
	defer close(raw)

	go func() {
		defer close(done)
		for {
			b := make([]byte, 32, 64)
			n, err := r.f.Read(b)
			if err != nil {
				done <- err
				return
			}
			raw <- b[:n]
		}
	}()

	for {
		select {
		case b := <-raw:
			d, err := parseRespiration(b)
			if err != nil {
				log.Println(err)
			}
			r.data <- d
		case err := <-done:
			log.Println(err)
			return
		case <-time.After(r.Timeout):
			// TODO on timeout do somthing smarter
			return
		}
	}

}

const (
	respirationStartByte = 0x50
)

func parseRespiration(b []byte) (Respiration, error) {
	if b[0] != respirationStartByte {
		return Respiration{}, errParseRespDataNoResoirationByte
	}
	if len(b) < 29 {
		return Respiration{}, errParseRespDataNotEnoughBytes
	}
	data := Respiration{}
	data.Time = time.Now().UnixNano()
	data.Status = binary.BigEndian.Uint32(b[1:5])
	data.Counter = binary.BigEndian.Uint32(b[5:9])
	data.State = resperationState(binary.BigEndian.Uint32(b[9:13]))
	data.RPM = binary.BigEndian.Uint32(b[13:17])
	data.Distance = float64(math.Float32frombits(binary.BigEndian.Uint32(b[17:21])))
	data.SignalQuality = float64(math.Float32frombits(binary.BigEndian.Uint32(b[21:25])))
	data.Movement = float64(math.Float32frombits(binary.BigEndian.Uint32(b[25:29])))
	return data, nil
}

var (
	errParseRespDataNoResoirationByte = errors.New("does not start with respiration byte")
	errParseRespDataNotEnoughBytes    = errors.New("response does not contain enough bytes")
)
