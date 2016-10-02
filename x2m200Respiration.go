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
	Status        status
	Counter       uint32
	State         respirationState
	RPM           uint32
	Distance      float64
	SignalQuality float64
	Movement      float64
}

type status uint32

//go:generate jsonenums -type=status
//go:generate stringer -type=status
const (
	respApp status = 594935334
)

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

type BaseBandAmpPhase struct {
	Time         int64     `json:"time"`
	Counter      uint32    `json:"counter"`
	Bins         uint32    `json:"bins"`
	BinLength    float64   `json:"binlength"`
	SamplingFreq float64   `json:"samplingfreq"`
	CarrierFreq  float64   `json:"carrierfreq"`
	RangeOffset  float64   `json:"rangeoffset"`
	Amplitude    []float64 `json:"amplitude"`
	Phase        []float64 `json:"phase"`
}

//go:generate jsonenums -type=respirationState
//go:generate stringer -type=respirationState174
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
	parser := parse
	switch mode {
	case "respiration":
		appID = [4]byte{0xd6, 0xa2, 0x23, 0x14}
		parser = parse
	case "sleep":
		log.Println("Loading Sleep Module")
		appID = [4]byte{0x17, 0x7b, 0xf1, 0x00}
		parser = parse
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

	return module
}

// Reset is
// func (r *Module) Reset() (bool, error) {
// 	log.Println("Called Reset")
// 	return r.f.Reset()
// }

type ledMode byte

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

var x2m200DetectionZone = [4]byte{0x96, 0xa1, 0x0a, 0x1c}

// var x2m200DetectionZone = [4]byte{0x1c, 0x0a, 0xa1, 0x96}

// SetDetectionZone is
// Example: <Start> + <XTS_SPC_APPCOMMAND> + <XTS_SPCA_SET> + [XTS_ID_DETECTION_ZONE(i)] + [Start(f)] + [End(f)] + <CRC> + <End>
// Response: <Start> + <XTS_SPR_ACK> + <CRC> + <End>
func (r Module) SetDetectionZone(start, end float64) {
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
func (r Module) SetSensitivity(sensitivity int) {

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
func (r Module) Load() error {
	n, err := r.f.Write([]byte{x2m200LoadModule, r.AppID[0], r.AppID[1], r.AppID[2], r.AppID[3]})
	if err != nil {
		log.Println(err, n)
		return err
	}
	b := make([]byte, 2048)
	n, err = r.f.Read(b)
	if err != nil {
		log.Println(err, n)
		return err
	}
	if b[0] != x2m200Ack {
		log.Printf("%#02x\n", b[0:n])
		log.Println("Not Ack")
		return errors.New("load module error, was not ack'ed")
	}
	return nil
}

// const unsigned long xts_id_baseband_iq = 0x0000000c;
// const unsigned long xts_id_baseband_amplitude_phase = 0x0000000d;
//
// const unsigned long xts_sacr_outputbaseband = 0x00000010;
// const unsigned long xts_sacr_id_baseband_output_off = 0x00000000;
// const unsigned long xts_sacr_id_baseband_output_amplitude_phase = 0x00000002;

// const unsigned char xts_spc_dir_command = 0x90;
// const unsigned char xts_sdc_app_setint = 0x71;

// void enable_raw_data()
// {
//   long data_length = 1;
//
//   //Fill send buffer
//   send_buf[0] = xts_spc_dir_command;
//   send_buf[1] = xts_sdc_app_setint;
//
//   memcpy(send_buf+2, &xts_sacr_outputbaseband, 4);
//   memcpy(send_buf+6, &data_length, 4);
//   memcpy(send_buf+10, &xts_sacr_id_baseband_output_amplitude_phase, 4);
//
//   //Send the command
//   send_command(send_buf, 14);
//
//   //Get response
//   receive_data(true);
//
//   // Check if acknowledge was received
//   check_ack();
// }

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

	b := make([]byte, 1024)
	n, err := r.f.Read(b)
	if err != nil {
		log.Println(err, n)
		return err
	}
	if b[0] != x2m200Ack {
		log.Printf("%#02x\n", b[0:n])
		log.Println("Not Ack")
		return errors.New("error enable Phase Amp Baseband was not ack'ed")
	}
	return nil
}

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
			data, err := r.parser(out)
			if err != nil {
				log.Println(err)
			}
			stream <- data
			// switch v := data.(type) {
			// case Respiration:
			// 	respiration <- data.(Respiration)
			// }
			// d := data.(Respiration)

			// data, err := r.parser(out)
			// if err != nil {
			// 	log.Println(err)
			// }

			// switch v := data.(type) {
			// case Respiration:
			// 	data.(Respiration)
			// }
			// d := data.(Respiration)
			// log.Printf("%#+v\n", data)
		}

	}
	//
	// for {
	// 	b := make([]byte, 32, 64)
	// 	n, err := r.f.Read(b)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	// log.Println(b[:n], n)
	// 	data, err := parseRespiration(b[:n])
	// 	for err != nil {
	// 		log.Println(err)
	// 		if err == errParseRespDataNoResoirationByte {
	// 			b = b[:0+copy(b[0:], b[1:])]
	// 			n--
	// 			data, err = parseRespiration(b[:n])
	// 		}
	// 		if err == errParseRespDataNotEnoughBytes || err == errNoData {
	// 			break
	// 		}
	// 	}
	// 	d := data.(Respiration)
	//
	// 	log.Printf("%#+v\n", d)
	// }
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
	data.Status = binary.LittleEndian.Uint32(b[1:5])
	data.Counter = binary.LittleEndian.Uint32(b[5:9])
	data.State = respirationState(binary.LittleEndian.Uint32(b[9:13]))
	data.RPM = binary.LittleEndian.Uint32(b[13:17])
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
