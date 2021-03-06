// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/eraclitux/stracer"
)

// Type lessFunc is used for sorting interface date with multiple keys.
type lessFunc func(e1, e2 *interfaceData) bool

// byKey is an helper function that returns a lessFunc type.
// k arguments can be any key declared in makeValueMap.
func byKey(k string) lessFunc {
	return func(e1, e2 *interfaceData) bool {
		if e1.err != nil {
			return true
		}
		// reverse order
		return e1.rates[k] > e2.rates[k]
	}
}

// multiSorter implements the sort.Interface.
type multiSorter struct {
	interfaces    []interfaceData
	lessFunctions []lessFunc
}

// sort sorts the slice passed using lessFunc passed to orderBy.
func (ms *multiSorter) sort(d []interfaceData) {
	ms.interfaces = d
	sort.Sort(ms)
}

// orderBy returns a multiSorter loaded with a slice of lessFunc.
func orderBy(lf ...lessFunc) *multiSorter {
	ms := &multiSorter{}
	ms.lessFunctions = lf
	return ms
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.interfaces)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.interfaces[i], ms.interfaces[j] = ms.interfaces[j], ms.interfaces[i]
}

// Less is part of sort.Interface.
func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.interfaces[i], &ms.interfaces[j]
	var k int
	for k = 0; k < len(ms.lessFunctions)-1; k++ {
		less := ms.lessFunctions[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
		// here if p == q, lets try next lessFunc
	}
	// returns result of last lessFunc
	return ms.lessFunctions[k](p, q)
}

// convertKeys converts user suplied sort keys to be used in
// interfaceData.rates as we store Bps not bps
// (convertion is handled in displayResults).
func convertKeys(keys []string) {
	for i, v := range keys {
		switch v {
		case "tx-Kbps":
			keys[i] = "tx-Bps"
		case "rx-Kbps":
			keys[i] = "rx-Bps"
		}
	}
}

// sanitizeSortKeys parses user supplied sorting keys.
// It returns error if supplied keys are invalid.
func sanitizeSortKeys(keys ...string) ([]string, error) {
	vKeys := [...]string{"tx-Kbps", "tx-pps", "tx-eps", "tx-dps", "rx-Kbps", "rx-pps", "rx-eps", "rx-dps"}
	found := false
	for _, k := range keys {
		for _, v := range vKeys {
			if k == v {
				stracer.Traceln("found matching key", v, k)
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("Invalid sort key: %v. Use one in %v", k, vKeys)
		}
		found = false
	}
	stracer.Traceln("Valid supplied sorting keys:", keys)
	convertKeys(keys)
	return keys, nil
}

func printHead(extOut bool) {
	if extOut {
		fmt.Printf(
			"%20s%12s%9s%9s%12s%12s%12s%12s%12s%12s\n",
			"Host",
			"Interface",
			"Rx-Kb/s",
			"Tx-Kb/s",
			"Rx-Pckts/s",
			"Tx-Pckts/s",
			"Rx-Drp/s",
			"Tx-Drp/s",
			"Rx-Err/s",
			"Tx-Err/s",
		)
	} else {
		fmt.Printf(
			"%20s%12s%9s%9s%12s%12s%12s%12s\n",
			"Host",
			"Interface",
			"Rx-Kb/s",
			"Tx-Kb/s",
			"Rx-Pckts/s",
			"Tx-Pckts/s",
			"Rx-Drp/s",
			"Tx-Drp/s",
		)
	}
}

func displayResults(results []interfaceData, noHead bool, extOut bool) {
	for i, r := range results {
		// Clean last spinner char.
		// We have to call this in displayResults
		// which is executed before of this.
		fmt.Fprintf(os.Stderr, "\r%s", "")
		if i%20 == 0 && !noHead {
			printHead(extOut)
		}
		if r.err != nil {
			fmt.Fprintln(os.Stderr, "[ERROR]", r.host, r.err)
		} else {
			fmt.Printf("%20s", r.host)
			fmt.Printf("%12s", r.name)
			fmt.Printf("%9d", uint64(r.rates["rx-Bps"]*8/1024))
			fmt.Printf("%9d", uint64(r.rates["tx-Bps"]*8/1024))
			fmt.Printf("%12d", r.rates["rx-pps"])
			fmt.Printf("%12d", r.rates["tx-pps"])
			fmt.Printf("%12d", r.rates["rx-dps"])
			fmt.Printf("%12d", r.rates["tx-dps"])
			if extOut {
				fmt.Printf("%12d", r.rates["rx-eps"])
				fmt.Printf("%12d", r.rates["tx-eps"])
			}
			fmt.Println("")
		}
	}
}

func showSpinner() chan<- struct{} {
	doneC := make(chan struct{})
	tick := time.Tick(100 * time.Millisecond)
	go func() {
		for {
			for _, r := range `-\|/` {
				select {
				case <-doneC:
					// We have to call this in displayResults
					// which may be (and usually is) executed before of this.
					//fmt.Fprintf(os.Stderr, "\r%s", "")
					return
				case <-tick:
					fmt.Fprintf(os.Stderr, "\r%c", r)
				}
			}
		}
	}()
	return doneC
}
