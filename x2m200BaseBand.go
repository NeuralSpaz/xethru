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

// Package xethru a open source implementation driver for xethru sensor modules.
// The current state of this library is still unstable and under active development.
// Contributions are welcome.
// To use with the X2M200 module you will first need to create a
// serial io.ReadWriter (there is an examples in the example dir)
// then you can use Open to create a new x2m200 device that
// will handle all the start, end, crc and escaping for you.

package xethru

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

// BaseBandIQ is the struct
type BaseBandIQ struct {
	Time         int64     `json:"time"`
	Counter      uint32    `json:"counter"`
	Bins         uint32    `json:"bins"`
	BinLength    float64   `json:"binlength"`
	SamplingFreq float64   `json:"samplingfreq"`
	CarrierFreq  float64   `json:"carrierfreq"`
	RangeOffset  float64   `json:"rangeoffset"`
	SigI         []float64 `json:"sigi"`
	SigQ         []float64 `json:"sigq"`
}

// BaseBandAmpPhase is the struct
// type BaseBandAmpPhase struct {
// 	Time         int64     `json:"time"`
// 	Counter      uint32    `json:"counter"`
// 	Bins         uint32    `json:"bins"`
// 	BinLength    float64   `json:"binlength"`
// 	SamplingFreq float64   `json:"samplingfreq"`
// 	CarrierFreq  float64   `json:"carrierfreq"`
// 	RangeOffset  float64   `json:"rangeoffset"`
// 	Amplitude    []float64 `json:"amplitude"`
// 	Phase        []float64 `json:"phase"`
// }

const (
	x2m200AppData    = 0x50
	x2m200BaseBandIQ = 0x0C
	x2m200BaseBandAP = 0x0D
	iqheadersize     = 29
	apheadersize     = 29
)

// Example: <x2m200AppData> + [x2m200BaseBandIQ] + [Counter(i)] + [SamplingFrequency(f)]
// + [CarrierFrequency(f)] + [RangeOffset(f)] + [NumOfBins(i)] + [SigI(f)] + ... + [SigQ(f)] + ... +
func parseBaseBandIQ(b []byte) (BaseBandIQ, error) {
	if len(b) < 1 {
		return BaseBandIQ{}, errParseBaseBandIQNoData
	}
	if b[0] != x2m200AppData {
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

// Example: <XTS_SPR_APPDATA> + [XTS_ID_BASEBAND_AMPLITUDE_PHASE(i)] + [Counter(i)] + [NumOfBins(i)]
// + [BinLength(f)] + [SamplingFrequency(f)] + [CarrierFrequency(f)] + [RangeOffset(f)] + [Amplitude(f)]
// + ... + [Phase(f)] + ... + <CRC> + <End>

// func parseBaseBandAP(b []byte) (BaseBandAmpPhase, error) {
// 	if len(b) < 1 {
// 		return BaseBandAmpPhase{}, errParseBaseBandAPNoData
// 	}
// 	if b[0] != x2m200AppData {
// 		return BaseBandAmpPhase{}, errParseBaseBandAPNotBaseBand
// 	}
// 	if len(b) < apheadersize {
// 		return BaseBandAmpPhase{}, errParseBaseBandAPNotEnoughBytes
// 	}
// 	x2m200basebandap := binary.LittleEndian.Uint32(b[1:5])
// 	if x2m200basebandap != x2m200BaseBandAP {
// 		log.Println(x2m200basebandap, x2m200BaseBandAP)
// 		return BaseBandAmpPhase{}, errParseBaseBandAPDataHeader
// 	}
//
// 	var ap BaseBandAmpPhase
// 	ap.Time = time.Now().UnixNano()
// 	ap.Counter = binary.LittleEndian.Uint32(b[5:9])
// 	ap.Bins = binary.LittleEndian.Uint32(b[9:13])
// 	ap.BinLength = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[13:17])))
// 	ap.SamplingFreq = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[17:21])))
// 	ap.CarrierFreq = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[21:25])))
// 	ap.RangeOffset = float64(math.Float32frombits(binary.LittleEndian.Uint32(b[25:29])))
//
// 	if len(b) < int(iqheadersize+uint32(ap.Bins)) {
// 		return ap, errParseBaseBandAPIncompletePacket
// 	}
//
// 	for i := apheadersize; i < int((ap.Bins*4)+apheadersize); i += 4 {
// 		amplitude := float64(math.Float32frombits(binary.LittleEndian.Uint32(b[i : i+4])))
// 		ap.Amplitude = append(ap.Amplitude, amplitude)
// 	}
//
// 	for i := int(apheadersize + 4*ap.Bins); i < int((ap.Bins*8)+apheadersize); i += 4 {
// 		phase := float64(math.Float32frombits(binary.LittleEndian.Uint32(b[i : i+4])))
// 		ap.Phase = append(ap.Phase, phase)
// 	}
//
// 	return ap, nil
// }
//
// var (
// 	errParseBaseBandAPNoData           = errors.New("baseband data is zero length")
// 	errParseBaseBandAPNotBaseBand      = errors.New("baseband data does not start with x2m200AppData")
// 	errParseBaseBandAPNotEnoughBytes   = errors.New("baseband data does contain enough bytes")
// 	errParseBaseBandAPDataHeader       = errors.New("baseband data does contain ap baseband header")
// 	errParseBaseBandAPIncompletePacket = errors.New("baseband data does contain a full packet of data")
// )
