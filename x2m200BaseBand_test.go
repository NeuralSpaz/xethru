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
	"fmt"
	"log"
	"math"
	"testing"
)

func TestParseBaseBandIQNoData(t *testing.T) {
	b := []byte{}
	_, err := parseBaseBandIQ(b)

	if err != errParseBaseBandIQNoData {
		t.Errorf("expected %v, got %v", errParseBaseBandIQNoData, err)
	}
}

func TestParseBaseBandIQNotBaseBand(t *testing.T) {
	b := []byte{0x00}
	_, err := parseBaseBandIQ(b)

	if err != errParseBaseBandIQNotBaseBand {
		t.Errorf("expected %v, got %v", errParseBaseBandIQNotBaseBand, err)
	}
}

func TestParseBaseBandIQNotEnoughBytes(t *testing.T) {
	b := []byte{x2m200AppData}
	_, err := parseBaseBandIQ(b)

	if err != errParseBaseBandIQNotEnoughBytes {
		t.Errorf("expected %v, got %v", errParseBaseBandIQNotEnoughBytes, err)
	}
}

func TestParseBaseBandIQDataHeaderFail(t *testing.T) {
	b := []byte{x2m200AppData, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_, err := parseBaseBandIQ(b)

	if err != errParseBaseBandIQDataHeader {
		t.Errorf("expected %v, got %v", errParseBaseBandIQDataHeader, err)
	}
}

func TestParseBaseBandIQIncompletePacket(t *testing.T) {
	b := []byte{x2m200AppData, 0x0C, 0x00, 0x00, 0x00}
	_, err := parseBaseBandIQ(b)

	if err != errParseBaseBandIQNotEnoughBytes {
		t.Errorf("expected %v, got %v", errParseBaseBandIQNotEnoughBytes, err)
	}
}

func TestParseBaseBandIQParseHeader(t *testing.T) {
	b := []byte{x2m200AppData, 0x0C, 0x00, 0x00, 0x00}

	Counter := make([]byte, 4)
	counter := uint32(1000)
	binary.LittleEndian.PutUint32(Counter, counter)
	b = append(b, Counter...)
	Bins := make([]byte, 4)
	bins := uint32(5)
	binary.LittleEndian.PutUint32(Bins, bins)
	b = append(b, Bins...)
	BinLength := make([]byte, 4)
	binlength := 1.0
	binary.LittleEndian.PutUint32(BinLength, math.Float32bits(float32(binlength)))
	b = append(b, BinLength...)
	SamplingFreq := make([]byte, 4)
	samplingfreq := 10.0
	binary.LittleEndian.PutUint32(SamplingFreq, math.Float32bits(float32(samplingfreq)))
	b = append(b, SamplingFreq...)
	CarrierFreq := make([]byte, 4)
	carrierfreq := 100.0
	binary.LittleEndian.PutUint32(CarrierFreq, math.Float32bits(float32(carrierfreq)))
	b = append(b, CarrierFreq...)
	RangeOffset := make([]byte, 4)
	rangeoffset := 300.0
	binary.LittleEndian.PutUint32(RangeOffset, math.Float32bits(float32(rangeoffset)))
	b = append(b, RangeOffset...)

	fmt.Printf("%d %x\n", len(b), b)

	iq, err := parseBaseBandIQ(b)

	if err != errParseBaseBandIQIncompletePacket {
		t.Errorf("expected %v, got %v", errParseBaseBandIQIncompletePacket, err)
	}
	if iq.Counter != counter {
		t.Errorf("expected %v, got %v", counter, iq.Counter)
	}
	if iq.Bins != bins {
		t.Errorf("expected %v, got %v", bins, iq.Bins)
	}
	if iq.BinLength != binlength {
		t.Errorf("expected %v, got %v", binlength, iq.BinLength)
	}
	if iq.SamplingFreq != samplingfreq {
		t.Errorf("expected %v, got %v", samplingfreq, iq.SamplingFreq)
	}
	if iq.CarrierFreq != carrierfreq {
		t.Errorf("expected %v, got %v", carrierfreq, iq.CarrierFreq)
	}
	if iq.RangeOffset != rangeoffset {
		t.Errorf("expected %v, got %v", rangeoffset, iq.RangeOffset)
	}
}

func TestParseBaseBandIQParseData(t *testing.T) {
	b := []byte{x2m200AppData, 0x0C, 0x00, 0x00, 0x00}

	Counter := make([]byte, 4)
	counter := uint32(1000)
	binary.LittleEndian.PutUint32(Counter, counter)
	b = append(b, Counter...)
	Bins := make([]byte, 4)
	bins := uint32(5)
	binary.LittleEndian.PutUint32(Bins, bins)
	b = append(b, Bins...)
	BinLength := make([]byte, 4)
	binlength := 1.0
	binary.LittleEndian.PutUint32(BinLength, math.Float32bits(float32(binlength)))
	b = append(b, BinLength...)
	SamplingFreq := make([]byte, 4)
	samplingfreq := 10.0
	binary.LittleEndian.PutUint32(SamplingFreq, math.Float32bits(float32(samplingfreq)))
	b = append(b, SamplingFreq...)
	CarrierFreq := make([]byte, 4)
	carrierfreq := 100.0
	binary.LittleEndian.PutUint32(CarrierFreq, math.Float32bits(float32(carrierfreq)))
	b = append(b, CarrierFreq...)
	RangeOffset := make([]byte, 4)
	rangeoffset := 300.0
	binary.LittleEndian.PutUint32(RangeOffset, math.Float32bits(float32(rangeoffset)))
	b = append(b, RangeOffset...)

	fmt.Printf("%d %x\n", len(b), b)

	Sigi1 := make([]byte, 4)
	sigi1 := 400.0
	binary.LittleEndian.PutUint32(Sigi1, math.Float32bits(float32(sigi1)))
	b = append(b, Sigi1...)
	Sigi2 := make([]byte, 4)
	sigi2 := 500.0
	binary.LittleEndian.PutUint32(Sigi2, math.Float32bits(float32(sigi2)))
	b = append(b, Sigi2...)
	Sigi3 := make([]byte, 4)
	sigi3 := 600.0
	binary.LittleEndian.PutUint32(Sigi3, math.Float32bits(float32(sigi3)))
	b = append(b, Sigi3...)
	Sigi4 := make([]byte, 4)
	sigi4 := 700.0
	binary.LittleEndian.PutUint32(Sigi4, math.Float32bits(float32(sigi4)))
	b = append(b, Sigi4...)
	Sigi5 := make([]byte, 4)
	sigi5 := 800.0
	binary.LittleEndian.PutUint32(Sigi5, math.Float32bits(float32(sigi5)))
	b = append(b, Sigi5...)

	Sigq1 := make([]byte, 4)
	sigq1 := 1400.0
	binary.LittleEndian.PutUint32(Sigq1, math.Float32bits(float32(sigq1)))
	b = append(b, Sigq1...)
	Sigq2 := make([]byte, 4)
	sigq2 := 1500.0
	binary.LittleEndian.PutUint32(Sigq2, math.Float32bits(float32(sigq2)))
	b = append(b, Sigq2...)
	Sigq3 := make([]byte, 4)
	sigq3 := 1600.0
	binary.LittleEndian.PutUint32(Sigq3, math.Float32bits(float32(sigq3)))
	b = append(b, Sigq3...)
	Sigq4 := make([]byte, 4)
	sigq4 := 1700.0
	binary.LittleEndian.PutUint32(Sigq4, math.Float32bits(float32(sigq4)))
	b = append(b, Sigq4...)
	Sigq5 := make([]byte, 4)
	sigq5 := 1800.0
	binary.LittleEndian.PutUint32(Sigq5, math.Float32bits(float32(sigq5)))
	b = append(b, Sigq5...)

	fmt.Printf("%d %x\n", len(b), b)

	iq, err := parseBaseBandIQ(b)

	if err != nil {
		t.Errorf("expected %v, got %v", errParseBaseBandIQNotEnoughBytes, err)
	}
	if iq.Counter != counter {
		t.Errorf("expected %v, got %v", counter, iq.Counter)
	}
	if iq.Bins != bins {
		t.Errorf("expected %v, got %v", bins, iq.Bins)
	}
	if iq.BinLength != binlength {
		t.Errorf("expected %v, got %v", binlength, iq.BinLength)
	}
	if iq.SamplingFreq != samplingfreq {
		t.Errorf("expected %v, got %v", samplingfreq, iq.SamplingFreq)
	}
	if iq.CarrierFreq != carrierfreq {
		t.Errorf("expected %v, got %v", carrierfreq, iq.CarrierFreq)
	}
	if iq.RangeOffset != rangeoffset {
		t.Errorf("expected %v, got %v", rangeoffset, iq.RangeOffset)
	}

	log.Println(iq)
}

// var testPacket = []byte{0x50, 0x0c, 0x00, 0x00, 0x00, 0xe8, 0x03, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3f, 0x00, 0x00, 0x20, 0x41, 0x00, 0x00, 0xc8, 0x42, 0x00, 0x00, 0x96, 0x43}

func benchsetup(n int) []byte {
	p := []byte{x2m200AppData, 0x0C, 0x00, 0x00, 0x00}

	Counter := make([]byte, 4)
	counter := uint32(1000)
	binary.LittleEndian.PutUint32(Counter, counter)
	p = append(p, Counter...)
	Bins := make([]byte, 4)
	bins := uint32(n)
	binary.LittleEndian.PutUint32(Bins, bins)
	p = append(p, Bins...)
	BinLength := make([]byte, 4)
	binlength := 1.0
	binary.LittleEndian.PutUint32(BinLength, math.Float32bits(float32(binlength)))
	p = append(p, BinLength...)
	SamplingFreq := make([]byte, 4)
	samplingfreq := 10.0
	binary.LittleEndian.PutUint32(SamplingFreq, math.Float32bits(float32(samplingfreq)))
	p = append(p, SamplingFreq...)
	CarrierFreq := make([]byte, 4)
	carrierfreq := 100.0
	binary.LittleEndian.PutUint32(CarrierFreq, math.Float32bits(float32(carrierfreq)))
	p = append(p, CarrierFreq...)
	RangeOffset := make([]byte, 4)
	rangeoffset := 300.0
	binary.LittleEndian.PutUint32(RangeOffset, math.Float32bits(float32(rangeoffset)))
	p = append(p, RangeOffset...)

	for i := 0; i < int(2*bins); i++ {
		Sig := make([]byte, 4)
		sig := 1.0 * float64(i)
		binary.LittleEndian.PutUint32(Sig, math.Float32bits(float32(sig)))
		p = append(p, Sig...)
	}
	log.Println(len(p))
	return p
}

func benchmarkParseBaseBand(p []byte, b *testing.B) {
	var iq BaseBandIQ
	for i := 0; i < b.N; i++ {
		iq, _ = parseBaseBandIQ(p)
	}
	log.Println(iq)
}

func BenchmarkParseBaseBand10(b *testing.B)  { benchmarkParseBaseBand(benchsetup(10), b) }
func BenchmarkParseBaseBand20(b *testing.B)  { benchmarkParseBaseBand(benchsetup(20), b) }
func BenchmarkParseBaseBand30(b *testing.B)  { benchmarkParseBaseBand(benchsetup(40), b) }
func BenchmarkParseBaseBand100(b *testing.B) { benchmarkParseBaseBand(benchsetup(100), b) }
func BenchmarkParseBaseBand200(b *testing.B) { benchmarkParseBaseBand(benchsetup(200), b) }
func BenchmarkParseBaseBand400(b *testing.B) { benchmarkParseBaseBand(benchsetup(400), b) }
