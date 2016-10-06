package xethru

import (
	"encoding/binary"
	"errors"
	"log"
	"math"
	"time"
)

const (
	appDataByte            = 0x50
	respirationStartByte   = 0x26
	sleepStartByte         = 0x6c
	phaseAmpltudeStartByte = 0x0d
)

func parse(b []byte) (interface{}, error) {
	// log.Printf("%02x\n", b)
	if len(b) == 0 {
		return Respiration{}, errNoData
	}
	if b[0] != appDataByte {
		return nil, errors.New("Not Apllication Data")
	}
	switch b[1] {
	case respirationStartByte:
		resp, err := parseRespiration(b)
		return resp, err
	case sleepStartByte:
		return parseSleep(b)
	case phaseAmpltudeStartByte:
		return parseBaseBandAP(b)
	default:
		return nil, errors.New("Not Implemented")
	}
}

func parseRespiration(b []byte) (Respiration, error) {
	// log.Printf("%02x\n", b)
	if len(b) == 0 {
		return Respiration{}, errNoData
	}
	if b[0] != appDataByte {
		return Respiration{}, errors.New("Not Apllication Data")
	}
	if b[1] != respirationStartByte {
		return Respiration{}, errParseRespDataNoResoirationByte
	}
	if len(b) < 29 {
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
	errParseRespDataNoResoirationByte = errors.New("does not start with respiration byte")
	errParseRespDataNotEnoughBytes    = errors.New("response does not contain enough bytes")
	errNoData                         = errors.New("no data to parse")
)

func parseSleep(b []byte) (interface{}, error) {
	// console.log
	// log.Printf("%02x %d\n", b, len(b))
	if len(b) == 0 {
		return Sleep{}, errParseSleepDataNoData
	}
	if b[0] != appDataByte {
		return nil, errors.New("Not Apllication Data")
	}
	if b[1] != sleepStartByte {
		return Sleep{}, errParseSleepDataNoResoirationByte
	}
	if len(b) < 30 {
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
	errParseSleepDataNoResoirationByte = errors.New("does not start with respiration byte")
	errParseSleepDataNotEnoughBytes    = errors.New("response does not contain enough bytes")
	errParseSleepDataNoData            = errors.New("no data to parse")
)

const (
	x2m200AppData    = 0x50
	x2m200BaseBandIQ = 0x0C
	x2m200BaseBandAP = 0x0D
	iqheadersize     = 29
	apheadersize     = 29
)

func parseBaseBandAP(b []byte) (BaseBandAmpPhase, error) {
	if len(b) < 1 {
		return BaseBandAmpPhase{}, errParseBaseBandAPNoData
	}

	if b[0] != appDataByte {
		return BaseBandAmpPhase{}, errors.New("Not Apllication Data")
	}

	if b[1] != phaseAmpltudeStartByte {
		return BaseBandAmpPhase{}, errParseBaseBandAPNotBaseBand
	}
	if len(b) < apheadersize {
		return BaseBandAmpPhase{}, errParseBaseBandAPNotEnoughBytes
	}
	x2m200basebandap := binary.LittleEndian.Uint32(b[1:5])
	if x2m200basebandap != x2m200BaseBandAP {
		log.Println(x2m200basebandap, x2m200BaseBandAP)
		return BaseBandAmpPhase{}, errParseBaseBandAPDataHeader
	}

	var ap BaseBandAmpPhase
	ap.Time = time.Now().UnixNano()
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
	// log.Println(ap)

	// b, err := json.MarshalIndent(ap, "\t", "")
	// if err != nil {
	// 	fmt.Println("error:", err)
	// }
	// os.Stdout.Write(b)

	return ap, nil
}

var (
	errParseBaseBandAPNoData           = errors.New("baseband data is zero length")
	errParseBaseBandAPNotBaseBand      = errors.New("baseband data does not start with x2m200AppData")
	errParseBaseBandAPNotEnoughBytes   = errors.New("baseband data does contain enough bytes")
	errParseBaseBandAPDataHeader       = errors.New("baseband data does contain ap baseband header")
	errParseBaseBandAPIncompletePacket = errors.New("baseband data does contain a full packet of data")
)

func parseBaseBandIQ(b []byte) (BaseBandIQ, error) {
	if len(b) < 1 {
		return BaseBandIQ{}, errParseBaseBandIQNoData
	}
	if b[0] != appDataByte {
		return BaseBandIQ{}, errParseBaseBandIQNotBaseBand
	}
	if len(b) < iqheadersize {
		return BaseBandIQ{}, errParseBaseBandIQNotEnoughBytes
	}
	x2m200basebandiq := binary.LittleEndian.Uint32(b[1:5])
	if x2m200basebandiq != x2m200BaseBandIQ {
		// log.Println(x2m200basebandiq, x2m200BaseBandIQ)
		return BaseBandIQ{}, errParseBaseBandIQDataHeader
	}

	var iq BaseBandIQ
	iq.Time = time.Now().UnixNano()
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
	errParseBaseBandIQNoData           = errors.New("baseband data is zero length")
	errParseBaseBandIQNotBaseBand      = errors.New("baseband data does not start with x2m200AppData")
	errParseBaseBandIQNotEnoughBytes   = errors.New("baseband data does contain enough bytes")
	errParseBaseBandIQDataHeader       = errors.New("baseband data does contain iq baseband header")
	errParseBaseBandIQIncompletePacket = errors.New("baseband data does contain a full packet of data")
)
