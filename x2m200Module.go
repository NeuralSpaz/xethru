package xethru

import "time"

// Module Config
type Module struct {
	f                  Framer
	AppID              [4]byte
	LEDMode            ledMode
	DetectionZoneStart float32
	DetectionZoneEnd   float32
	Sensitivity        uint32
	Timeout            time.Duration
	Data               chan interface{}
	parser             func(b []byte) (interface{}, error)
}
