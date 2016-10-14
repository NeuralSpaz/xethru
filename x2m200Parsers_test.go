package xethru

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		b    []byte
		err  error
		resp interface{}
	}{
		// {[]byte{0xFF}, errParseNotImplemented, nil},
		{[]byte{}, errNoData, nil},
		{[]byte{appDataByte, respirationStartByte}, errParseRespDataNotEnoughBytes, Respiration{}},
		{[]byte{appDataByte, sleepStartByte}, errParseSleepDataNotEnoughBytes, Sleep{}},
		{[]byte{appDataByte, basebandPhaseAmpltudeStartByte}, errParseBaseBandAPNotEnoughBytes, BaseBandAmpPhase{}},
		{[]byte{appDataByte, basebandIQStartByte}, errParseBaseBandIQNotEnoughBytes, BaseBandIQ{}},
		// {[]byte{appDataByte, 0x00}, errParseNotImplemented, nil},
		// {[]byte{appDataByte, sleepStartByte}, errParseSleepDataNotEnoughBytes, BaseBandIQ{}},
	}
	for n, c := range cases {
		resp, err := parse(c.b)
		// log.Printf("%#v, %#v \n", resp, err)
		if err != c.err {
			t.Errorf("test %d Expected: %v, got %v\n", n, c.err, err)
		}
		resptype := reflect.TypeOf(resp)
		expecttype := reflect.TypeOf(c.resp)

		if resptype != expecttype {
			t.Errorf("test %d Expected: %v, got %v\n", n, expecttype, resptype)
		}

	}
}

func TestParseRespiration(t *testing.T) {
	cases := []struct {
		b    []byte
		err  error
		resp Respiration
	}{
		{
			[]byte{appDataByte, respirationStartByte},
			errParseRespDataNotEnoughBytes,
			Respiration{
				Time:          0,
				Status:        0,
				Counter:       0,
				State:         0,
				RPM:           0,
				Distance:      0,
				SignalQuality: 0,
				Movement:      0,
			}}, {
			[]byte{appDataByte, 0x26, 0xfe, 0x75, 0x23, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			nil,
			Respiration{
				Time:          0,
				Status:        respApp,
				Counter:       0,
				State:         0,
				RPM:           0,
				Distance:      0,
				SignalQuality: 0,
				Movement:      0,
			}},
	}
	for n, c := range cases {
		resp, err := parseRespiration(c.b)
		// log.Printf("%#v, %#v \n", resp, err)
		if err != c.err {
			t.Errorf("test %d Expected: %v, got %v\n", n, c.err, err)
		}
		resp.Time = 0
		if resp != c.resp {
			t.Errorf("test %d Expected: %#v, got %#v\n", n, c.resp, resp)
		}
	}
}

func TestParseSleep(t *testing.T) {
	cases := []struct {
		b    []byte
		err  error
		resp Sleep
	}{
		{
			[]byte{appDataByte, respirationStartByte},
			errParseSleepDataNotEnoughBytes,
			Sleep{
				Time:          0,
				Status:        0,
				Counter:       0,
				State:         0,
				RPM:           0,
				Distance:      0,
				SignalQuality: 0,
				MovementSlow:  0,
				MovementFast:  0,
			}}, {
			[]byte{appDataByte, 0x6c, 0xa1, 0x75, 0x23, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			nil,
			Sleep{
				Time:          0,
				Status:        sleepApp,
				Counter:       0,
				State:         0,
				RPM:           0,
				Distance:      0,
				SignalQuality: 0,
				MovementSlow:  0,
				MovementFast:  0,
			}},
	}
	for n, c := range cases {
		// log.Println(len(c.b))
		resp, err := parseSleep(c.b)
		// log.Printf("%#v, %#v \n", resp, err)
		if err != c.err {
			t.Errorf("test %d Expected: %v, got %v\n", n, c.err, err)
		}
		resp.Time = 0
		if resp != c.resp {
			t.Errorf("test %d Expected: %#v, got %#v\n", n, c.resp, resp)
		}
	}
}

func TestBasebandAmpPhase(t *testing.T) {
	cases := []struct {
		b    []byte
		err  error
		resp BaseBandAmpPhase
	}{
		{
			[]byte{appDataByte, basebandPhaseAmpltudeStartByte},
			errParseBaseBandAPNotEnoughBytes,
			BaseBandAmpPhase{}}, {
			[]byte{appDataByte, 0x0d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			nil,
			BaseBandAmpPhase{}}, {
			[]byte{appDataByte, 0x0d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			errParseBaseBandAPIncompletePacket,
			BaseBandAmpPhase{}}, {
			[]byte{appDataByte, 0x0d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			nil,
			BaseBandAmpPhase{}},
	}
	for n, c := range cases {
		// log.Println(len(c.b))
		_, err := parseBaseBandAP(c.b)
		// log.Printf("%#v, %#v \n", resp, err)
		if err != c.err {
			t.Errorf("test %d Expected: %v, got %v\n", n, c.err, err)
		}
		// TODO: Validate response
	}
}

func TestBasebandIQ(t *testing.T) {
	cases := []struct {
		b    []byte
		err  error
		resp BaseBandIQ
	}{
		{
			[]byte{appDataByte, basebandPhaseAmpltudeStartByte},
			errParseBaseBandIQNotEnoughBytes,
			BaseBandIQ{}}, {
			[]byte{appDataByte, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			nil,
			BaseBandIQ{}}, {
			[]byte{appDataByte, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			errParseBaseBandIQIncompletePacket,
			BaseBandIQ{}}, {
			[]byte{appDataByte, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			nil,
			BaseBandIQ{}},
	}
	for n, c := range cases {
		// log.Println(len(c.b))
		_, err := parseBaseBandIQ(c.b)
		// log.Printf("%#v, %#v \n", resp, err)
		if err != c.err {
			t.Errorf("test %d Expected: %v, got %v\n", n, c.err, err)
		}
		// TODO: Validate response
	}
}
