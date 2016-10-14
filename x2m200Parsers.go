package xethru

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

const (
	appDataByte                    = 0x50
	respirationStartByte           = 0x26
	sleepStartByte                 = 0x6c
	basebandPhaseAmpltudeStartByte = 0x0d
	basebandIQStartByte            = 0x0c
	systemMesg                     = 0x30
	systemBooting                  = 0x10
	systemReady                    = 0x11
	ack                            = 0x10
)

// Respiration is the struct
type Respiration struct {
	Time          int64            `json:"time"`
	Status        status           `json:"status"`
	Counter       uint32           `json:"counter"`
	State         respirationState `json:"state"`
	RPM           uint32           `json:"rpm"`
	Distance      float64          `json:"distance"`
	SignalQuality float64          `json:"signalquality"`
	Movement      float64          `json:"movement"`
}

// Sleep is the struct
type Sleep struct {
	Time          int64            `json:"time"`
	Status        status           `json:"type"`
	Counter       uint32           `json:"counter"`
	State         respirationState `json:"state"`
	RPM           float64          `json:"rpm"`
	Distance      float64          `json:"distance"`
	SignalQuality float64          `json:"signalquality"`
	MovementSlow  float64          `json:"movementslow"`
	MovementFast  float64          `json:"movementfast"`
}

// BaseBandAmpPhase is the struct
type BaseBandAmpPhase struct {
	Time         int64     `json:"time"`
	Status       status    `json:"type"`
	Counter      uint32    `json:"counter"`
	Bins         uint32    `json:"bins"`
	BinLength    float64   `json:"binlength"`
	SamplingFreq float64   `json:"samplingfreq"`
	CarrierFreq  float64   `json:"carrier"`
	RangeOffset  float64   `json:"offset"`
	Amplitude    []float64 `json:"amplitude"`
	Phase        []float64 `json:"phase"`
}

// BaseBandIQ is the struct
type BaseBandIQ struct {
	Time         int64     `json:"time"`
	Status       status    `json:"type"`
	Counter      uint32    `json:"counter"`
	Bins         uint32    `json:"bins"`
	BinLength    float64   `json:"binlength"`
	SamplingFreq float64   `json:"samplingfreq"`
	CarrierFreq  float64   `json:"carrier"`
	RangeOffset  float64   `json:"offset"`
	SigI         []float64 `json:"i"`
	SigQ         []float64 `json:"q"`
}

// SystemMessage is the struct
type SystemMessage struct {
	Message string
}

func parse(b []byte) (interface{}, error) {
	// log.Printf("%02x\n", b)
	if len(b) == 0 {
		return nil, errNoData
	}
	switch b[0] {
	case appDataByte:
		switch b[1] {
		case respirationStartByte:
			resp, err := parseRespiration(b)
			return resp, err
		case sleepStartByte:
			return parseSleep(b)
		case basebandPhaseAmpltudeStartByte:
			return parseBaseBandAP(b)
		case basebandIQStartByte:
			return parseBaseBandIQ(b)
		default:
			return b, errParseNotImplemented
		}
	case systemMesg:
		switch b[1] {
		case systemBooting:
			return SystemMessage{Message: "System Still booting"}, nil
		case systemReady:
			return SystemMessage{Message: "System Ready"}, nil
		default:
			return b, errParseNotImplemented
		}
	case ack:
		return SystemMessage{Message: "Command Ack'ed"}, nil

	default:
		return b, errParseNotImplemented
	}
	// return nil, fmt.Errorf("something went wrong: %#02x\n", b)
}

var (
	errParseNotImplemented = errors.New("Parser not implemented")
	errNoData              = errors.New("no data to parse")
)

const respsize = 29

func parseRespiration(b []byte) (Respiration, error) {
	// Check to make sure respiration data is long enough
	if len(b) != respsize {
		return Respiration{}, errParseRespDataNotEnoughBytes
	}
	data := Respiration{}
	data.Time = time.Now().UnixNano()
	data.Status = status(binary.LittleEndian.Uint32(b[1:5]))
	data.Counter = binary.LittleEndian.Uint32(b[5:9])
	data.State = respirationState(binary.LittleEndian.Uint32(b[9:13]))
	data.RPM = binary.LittleEndian.Uint32(b[13:17])
	data.Distance = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[17:21])))
	data.Movement = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[21:25])))
	data.SignalQuality = float64(binary.LittleEndian.Uint32(b[25:29]))

	// log.Println(data)
	return data, nil
}

var (
	errParseRespDataNotEnoughBytes = errors.New("response does not contain enough bytes")
)

const sleepsize = 33

func parseSleep(b []byte) (Sleep, error) {
	// Make sure we have enough bytes to parse packet without panic
	if len(b) != sleepsize {
		return Sleep{}, errParseSleepDataNotEnoughBytes
	}
	data := Sleep{}
	data.Time = time.Now().UnixNano()
	data.Status = status(binary.LittleEndian.Uint32(b[1:5]))
	data.Counter = binary.LittleEndian.Uint32(b[5:9])
	data.State = respirationState(binary.LittleEndian.Uint32(b[9:13]))
	data.RPM = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[13:17])))
	data.Distance = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[17:21])))
	data.SignalQuality = float64(binary.LittleEndian.Uint32(b[21:25]))
	data.MovementSlow = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[25:29])))
	data.MovementFast = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[29:33])))

	return data, nil
}

var (
	errParseSleepDataNotEnoughBytes = errors.New("response does not contain enough bytes")
)

const apheadersize = 29

func parseBaseBandAP(b []byte) (BaseBandAmpPhase, error) {
	// Make sure we have enough bytes to parse header without panic
	if len(b) < apheadersize {
		return BaseBandAmpPhase{}, errParseBaseBandAPNotEnoughBytes
	}
	var ap BaseBandAmpPhase
	ap.Time = time.Now().UnixNano()
	ap.Status = status(binary.LittleEndian.Uint32(b[1:5]))
	ap.Counter = binary.LittleEndian.Uint32(b[5:9])
	ap.Bins = binary.LittleEndian.Uint32(b[9:13])
	ap.BinLength = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[13:17])))
	ap.SamplingFreq = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[17:21])))
	ap.CarrierFreq = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[21:25])))
	ap.RangeOffset = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[25:29])))

	if len(b) < int(iqheadersize+uint32(ap.Bins)) {
		return ap, errParseBaseBandAPIncompletePacket
	}

	for i := apheadersize; i < int((ap.Bins*4)+apheadersize); i += 4 {
		amplitude := float64(math.Float32frombits(binary.LittleEndian.Uint32(b[i : i+4])))
		ap.Amplitude = append(ap.Amplitude, amplitude)
	}

	for i := int(apheadersize + 4*ap.Bins); i < int((ap.Bins*8)+apheadersize); i += 4 {
		phase := float64(math.Float32frombits(binary.LittleEndian.Uint32(b[i : i+4])))
		ap.Phase = append(ap.Phase, phase)
	}
	return ap, nil
}

var (
	errParseBaseBandAPNotEnoughBytes   = errors.New("baseband data does contain enough bytes")
	errParseBaseBandAPIncompletePacket = errors.New("baseband data does contain a full packet of data")
)

const iqheadersize = 29

func parseBaseBandIQ(b []byte) (BaseBandIQ, error) {
	// Make sure we have enough bytes to parse header without panic
	if len(b) < iqheadersize {
		return BaseBandIQ{}, errParseBaseBandIQNotEnoughBytes
	}

	var iq BaseBandIQ

	iq.Time = time.Now().UnixNano()
	iq.Status = status(binary.LittleEndian.Uint32(b[1:5]))
	iq.Counter = binary.LittleEndian.Uint32(b[5:9])
	iq.Bins = binary.LittleEndian.Uint32(b[9:13])
	iq.BinLength = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[13:17])))
	iq.SamplingFreq = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[17:21])))
	iq.CarrierFreq = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[21:25])))
	iq.RangeOffset = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[25:29])))

	if len(b) < int(iqheadersize+uint32(iq.Bins)) {
		return iq, errParseBaseBandIQIncompletePacket
	}

	for i := iqheadersize; i < int((iq.Bins*4)+iqheadersize); i += 4 {
		sigi := float64(math.Float32frombits(binary.LittleEndian.Uint32(b[i : i+4])))
		iq.SigI = append(iq.SigI, sigi)
	}

	for i := int(iqheadersize + 4*iq.Bins); i < int((iq.Bins*8)+iqheadersize); i += 4 {
		sigq := float64(math.Float32frombits(binary.LittleEndian.Uint32(b[i : i+4])))
		iq.SigQ = append(iq.SigQ, sigq)
	}

	return iq, nil

}

var (
	errParseBaseBandIQNotEnoughBytes   = errors.New("baseband data does contain enough bytes")
	errParseBaseBandIQIncompletePacket = errors.New("baseband data does contain a full packet of data")
)
