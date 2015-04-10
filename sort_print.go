// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package main

import (
	"fmt"
	"os"
	"sort"
)

// Type lessFunc is used for sorting interface date with multiple keys.
type lessFunc func(e1, e2 *interfaceData) bool

// byKey is an helper function that returns a lessFunc type.
// k arguments can be any key declared in makeValueMap.
func byKey(k string) lessFunc {
	return func(e1, e2 *interfaceData) bool {
		return e1.rates[k] < e2.rates[k]
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
		less := ms.lessFunctions[i]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
		// here if p == q, lets try next lessFunc
	}
	// returns result of last lessFunc
	// FIXME this should be always false so can it be avoided?
	//return false
	return ms.lessFunctions[i](p, q)
}

func printHead() {
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
}

func displayResults(results []interfaceData, noHead bool) {
	for i, r := range results {
		if i%40 == 0 && !noHead {
			printHead()
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
			fmt.Printf("%12d", r.rates["rx-eps"])
			fmt.Printf("%12d", r.rates["tx-eps"])
			fmt.Println("")
		}
	}
}
