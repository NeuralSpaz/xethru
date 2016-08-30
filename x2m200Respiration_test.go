package xethru

import "testing"

func TestResperationParser(t *testing.T) {
	cases := []struct {
		in  []byte
		err error
		out Respiration
	}{
		{[]byte{0x01, 0x02, 0x00}, errParseRespDataNoResoirationByte, Respiration{}},
		{[]byte{respirationStartByte, 0x02, 0x00}, errParseRespDataNotEnoughBytes, Respiration{}},
		{[]byte{respirationStartByte, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x40, 0xb0, 0x00, 0x00, 0x40, 0xc0, 0x00, 0x00, 0x40, 0xe0, 0x00, 0x00}, nil, Respiration{0, 1, 2, 3, 4, 5.5, 6, 7}},
	}
	for _, c := range cases {
		out, err := parseRespiration(c.in)

		if err != c.err {
			t.Errorf("expected %v, got %v", c.err, err)
		}
		if out.Status != c.out.Status {
			t.Errorf("expected %v, got %v", c.out, out)
		}
	}
}

func TestNewResperation(t *testing.T) {
	client, _, _ := newLoopBackXethru()
	App := NewRespiration(client)

	if App.AppID != [4]byte{0x14, 0x23, 0xa2, 0xd6} {
		t.Errorf("AppID not set")
	}
}

func TestRunResperation(t *testing.T) {
	repeat := 100
	receive := Respiration{0, 1, 2, 3, 4, 5.5, 6, 7}
	send := []byte{respirationStartByte, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x40, 0xb0, 0x00, 0x00, 0x40, 0xc0, 0x00, 0x00, 0x40, 0xe0, 0x00, 0x00}
	client, sensorTX, _ := newLoopBackXethru()
	App := NewRespiration(client)

	if App.AppID != [4]byte{0x14, 0x23, 0xa2, 0xd6} {
		t.Errorf("AppID not set")
	}

	go func() {
		for i := 0; i < repeat; i++ {
			sensorTX <- send
		}
	}()

	go App.Run()

	counter := 0
	for {
		select {
		case data, ok := <-App.data:
			counter++
			if !ok {
				return
			}
			if data.Status != receive.Status {
				t.Errorf("expected %v got %v", receive, data)
			}
			// json, err := json.MarshalIndent(&data, "", "\t")
			// if err != nil {
			// 	log.Println(err)
			// }
			// log.Println(string(json))
			// log.Println(data)
		}
	}

}
