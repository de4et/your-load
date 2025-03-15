package checker

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestCheckURL(t *testing.T) {
	tests := []struct {
		uri      string
		expected CheckerResponse
		err      error
	}{
		{
			uri: "https://stream.telko.ru/Pir5a_poz6_cam3/tracks-v1/index.fmp4.m3u8",
			expected: CheckerResponse{
				ProtocolType: ProtocolHLS,
			},
			err: nil,
		},
		{
			uri: "https://bitdash-a.akamaihd.net/content/MI201109210084_1/m3u8s/f08e80da-bf1d-4e3d-8899-f0f6155f6efa.m3u8",
			expected: CheckerResponse{
				ProtocolType: ProtocolHLS,
			},
			err: nil,
		},
	}
	for _, test := range tests {
		resp, err := NewChecker().CheckURL(test.uri)
		if !reflect.DeepEqual(resp, test.expected) || !errors.Is(err, test.err) {
			fmt.Printf("Test: %+v\n", test)
			fmt.Printf("test failed: resp - %+v expected - %+v err - %v expectederr - %v\n", resp, test.expected, err, test.err)
			t.Fail()
		}
	}
}
