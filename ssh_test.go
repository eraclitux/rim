package main

import (
	"reflect"
	"strings"
	"testing"
)

var outPut string = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
 bnep0:  391914     760    0    0    0     0          0         0   142925     941    0    0    0     0       0          0
 wlan0:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
    lo:  895637    1038    0    0    0     0          0         0   895637    1038    0    0    0     0       0          0
ZZZ
Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
 bnep0:  391914     760    0    0    0     0          0         0   143112     943    0    0    0     0       0          0
 wlan0:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
    lo:  895737    1040    0    0    0     0          0         0   895737    1040    0    0    0     0       0          0`

func TestParseOutput(t *testing.T) {
	testData := make(rawData)
	var expectedData rawData = map[string]map[string]uint64{
		"bnep0": map[string]uint64{"rx-pps": 0, "rx-eps": 0, "rx-dps": 0, "tx-Bps": 187, "tx-pps": 2, "tx-eps": 0, "tx-dps": 0, "rx-Bps": 0},
		"wlan0": map[string]uint64{"rx-eps": 0, "rx-dps": 0, "tx-Bps": 0, "tx-pps": 0, "tx-eps": 0, "tx-dps": 0, "rx-Bps": 0, "rx-pps": 0},
		"lo":    map[string]uint64{"tx-Bps": 100, "tx-pps": 2, "tx-eps": 0, "tx-dps": 0, "rx-Bps": 100, "rx-pps": 2, "rx-eps": 0, "rx-dps": 0},
	}
	r := strings.NewReader(outPut)
	parseOutput(r, testData)
	if !reflect.DeepEqual(testData, expectedData) {
		t.Log(testData)
		t.Log(expectedData)
		t.Fail()
	}
}
