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
)

// type Parser interface {
// 	Parse([]byte) error
// }
//
// type RespirationApp struct {
// 	n int
// }
//
// func Mangle(p Parser) {}

//
// type SleepApp struct {
// 	n int
// }
//
// type BaseBandAPApp struct {
// 	n int
// }
//
// type BaseBandIQApp struct {
// 	n int
// }
//
// func (r RespirationApp) Parse(b []byte) error {
// 	return nil
// }

// func (r SleepApp) Parse(b []byte) error {
// 	return nil
// }
// func (r BaseBandAPApp) Parse(b []byte) error {
// 	return nil
// }
// func (r BaseBandIQApp) Parse(b []byte) error {
// 	return nil
// }

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
			return nil, errParseNotImplemented
		}
	default:
		return nil, errParseNotImplemented
	}
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
