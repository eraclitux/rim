// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package main

import (
	"fmt"
	"os"
)

func presentSingleResult(r *jobResult) {
	fmt.Printf("%20.20s%12.12s%8.8s%8.8s\n", "Host", "Interface", "RX-KBps", "TX-KBps")
	if r.err != nil {
		fmt.Println(r.host, r.err)
	} else {
		// k is remote interface
		// v is a map with rates
		for k, v := range r.data {
			fmt.Printf("%20.20s", r.host)
			fmt.Printf("%12.12s", k)
			fmt.Printf("%8d", uint64(v["rx-Bps"]/1024))
			fmt.Printf("%8d", uint64(v["tx-Bps"]/1024))
			fmt.Println("")
		}
	}
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
