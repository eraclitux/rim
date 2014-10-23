// RIM - Remote Interfaces Monitor

/* Copyright (c) 2014 Andrea Masi
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE. */

package main

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var workers = runtime.NumCPU()

const (
	RemoteCommand = `cat /proc/net/dev; echo ZZZ; sleep 1; cat /proc/net/dev;`
	Separator     = "ZZZ\n"
)

// interfaceData models single interface' data for a given host
type interfaceData struct {
	host  string
	name  string
	rates map[string]uint64
	err   error
}

type jobResult struct {
	host string
	err  error
	// es: map[string]map[string]uint64{"eth0": map[string]uint64{"tx-Bps":12000, "rx-Bps":12000}}
	data map[string]map[string]uint64
}

type job struct {
	host            string
	sshClientConfig ssh.ClientConfig
	result          chan<- jobResult
}

// unpackJobResult reads data from a jobResult and unpack it into a slice of interfaceData, one for each interface.
func unpackJobResult(jr *jobResult) []interfaceData {
	data := []interfaceData{}
	if jr.err != nil {
		return []interfaceData{interfaceData{jr.host, "", nil, jr.err}}
	}
	for kInterface, vMap := range jr.data {
		i := interfaceData{jr.host, kInterface, vMap, nil}
		data = append(data, i)
	}
	return data
}

func sanitizeHost(s *string) {
	if !strings.Contains(*s, ":") {
		*s = fmt.Sprintf("%s:%d", *s, 22)
	}
}

func splitOnSpaces(s string) []string {
	// Remove leading a trailing white spaces that gives wrong results with regexp below
	trimmedS := strings.Trim(s, " ")
	return regexp.MustCompile(`\s+`).Split(trimmedS, -1)
}

// makeValueMap creates map with t2 or t1 values
func makeValueMap(data []string) map[string]uint64 {
	dataMap := make(map[string]uint64)
	for i, s := range data {
		// FIXME parse error
		converted, _ := strconv.ParseUint(s, 10, 64)
		switch i {
		case 0:
			// here its not bytes per second but absolute bytes at t2
			// however we'll store Bps later
			dataMap["rx-Bps"] = converted
		case 1:
			// here its not packets per second but absolute bytes at t2
			// however we'll store pps later
			dataMap["rx-pps"] = converted
		case 2:
			// here its not errors per second but absolute bytes at t2
			// however we'll store eps later
			dataMap["rx-eps"] = converted
		case 3:
			// here its not drop per second but absolute bytes at t2
			// however we'll store dps later
			dataMap["rx-dps"] = converted
		case 8:
			// here its not bytes per second but absolute bytes at t2
			// however we'll store Bps later
			dataMap["tx-Bps"] = converted
		case 9:
			// here its not bytes per second but absolute bytes at t2
			// however we'll store Bps later
			dataMap["tx-pps"] = converted
		case 10:
			// here its not errors per second but absolute bytes at t2
			// however we'll store eps later
			dataMap["tx-eps"] = converted
		case 11:
			// here its not drops per second but absolute bytes at t2
			// however we'll store dps later
			dataMap["tx-dps"] = converted
		}
	}
	return dataMap
}

// calculateRates uses t2 values stored in dataAtT2 by makeValueMap to calculate rates
func calculateRates(dataAtT2, dataAtT1 map[string]uint64) {
	for k, v := range dataAtT2 {
		// assuming that ΔT is always 1 second (sleep 1)
		dataAtT2[k] = v - dataAtT1[k]
	}
}

// parseOutput arranges RemoteCommand output and calculates rates
func parseOutput(out *bytes.Buffer, data map[string]map[string]uint64) error {
	outBytes := bytes.Split(out.Bytes(), []byte(Separator))
	// Contains interfaces' value at t1
	outOne := outBytes[0]
	// Contains interfaces' value at t2
	outTwo := outBytes[1]
	scanner := bufio.NewScanner(bytes.NewBuffer(outTwo))
	for scanner.Scan() {
		s := scanner.Text()
		// Excludes titles
		if strings.Contains(s, "|") {
			continue
		}
		splittedRow := strings.Split(s, ":")
		// remove white spaces
		iface := strings.Replace(splittedRow[0], " ", "", -1)
		countersData := splitOnSpaces(splittedRow[1])
		debugPrintln("parsed data @ t2:", iface, countersData)
		data[iface] = makeValueMap(countersData)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	scanner = bufio.NewScanner(bytes.NewBuffer(outOne))
	for scanner.Scan() {
		s := scanner.Text()
		// Excludes titles
		if strings.Contains(s, "|") {
			continue
		}
		splittedRow := strings.Split(s, ":")
		// remove white spaces
		iface := strings.Replace(splittedRow[0], " ", "", -1)
		countersData := splitOnSpaces(splittedRow[1])
		dataAtT1 := makeValueMap(countersData)
		debugPrintln("parsed data @ t1:", iface, countersData)
		calculateRates(data[iface], dataAtT1)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (j *job) getRemoteData() {
	var output bytes.Buffer
	destination := j.host
	sanitizeHost(&destination)
	conn, err := ssh.Dial("tcp", destination, &j.sshClientConfig)
	if err != nil {
		j.result <- jobResult{j.host, err, nil}
		return
	}
	session, err := conn.NewSession()
	if err != nil {
		j.result <- jobResult{j.host, err, nil}
		return
	}
	defer session.Close()
	session.Stdout = &output
	if err := session.Run(RemoteCommand); err != nil {
		j.result <- jobResult{j.host, err, nil}
		return
	}
	debugPrintln(output.String())
	data := make(map[string]map[string]uint64)
	if err := parseOutput(&output, data); err != nil {
		j.result <- jobResult{j.host, err, nil}
	}
	j.result <- jobResult{j.host, nil, data}
}

func createSshConfig(user, passwd string) ssh.ClientConfig {
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	authMethods := []ssh.AuthMethod{ssh.Password(passwd)}
	if sshAuthSock != "" {
		debugPrintln("ssh-agent socket:", sshAuthSock)
		socket, err := net.Dial("unix", sshAuthSock)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			agentClient := agent.NewClient(socket)
			authMethod := ssh.PublicKeysCallback(agentClient.Signers)
			authMethods = append(authMethods, authMethod)
			// FIXME works even without calling agent.ForwardToAgent()?
			debugPrintln("ssh-agent forwarding configured")
		}
	}
	return ssh.ClientConfig{
		User: user,
		Auth: authMethods,
	}
}

func populateQueue(jobs chan<- job, results chan<- jobResult, hosts []string, sshConfig ssh.ClientConfig) {
	for _, host := range hosts {
		jobs <- job{host, sshConfig, results}
	}
	close(jobs)
}

func evaluateQueue(jobs <-chan job) {
	for j := range jobs {
		j.getRemoteData()
	}
}

func parallelizeWorkers(jQueue chan job) {
	for i := 0; i < workers; i++ {
		go evaluateQueue(jQueue)
	}
}

// sortAndPresent gets results from workes, sorts and displays them.
// Deprecated
func sortAndPresent(jobDone, displayDone chan<- struct{}, results <-chan jobResult) {
	iData := []interfaceData{}
	for res := range results {
		iData = append(iData, unpackJobResult(&res)...)
		//presentResults(iData)
		jobDone <- struct{}{}
	}
	presentResults(iData)
	displayDone <- struct{}{}
}

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

func presentResults(results []interfaceData) {
	fmt.Printf(
		"%20s%12s%9s%9s%12s%12s%12s%12s%12s%12s\n",
		"Host",
		"Interface",
		"Rx-KB/s",
		"Tx-KB/s",
		"Rx-Pckts/s",
		"Tx-Pckts/s",
		"Rx-Drp/s",
		"Tx-Drp/s",
		"Rx-Err/s",
		"Tx-Err/s",
	)
	for _, r := range results {
		if r.err != nil {
			fmt.Println("[ERROR]", r.host, r.err)
		} else {
			fmt.Printf("%20s", r.host)
			fmt.Printf("%12s", r.name)
			fmt.Printf("%9d", uint64(r.rates["rx-Bps"]/1024))
			fmt.Printf("%9d", uint64(r.rates["tx-Bps"]/1024))
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

func getHostsFromFile(path string) []string {
	// FIXME sanityze user input
	if path == "[FILE]" {
		fmt.Fprintln(os.Stderr, "-f is mandatory")
		os.Exit(2)
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file")
		os.Exit(2)
	}
	hosts := strings.Split(string(bytes), "\n")
	// remove last empty element
	hosts = hosts[:len(hosts)-1]
	debugPrintln("Parsed hosts from file:", hosts)
	return hosts
}

func main() {
	jobsQueue := make(chan job, workers)
	resultQueue := make(chan jobResult, workers)
	hostsFileFlag := flag.String("f", "[FILE]", "File cointaining target hosts, one per line.")
	userFlag := flag.String("u", "[USER]", "Ssh username.")
	passwdFlag := flag.String("p", "[PASSWORD]", "Ssh password for remote hosts. Automatically use ssh-agent as fallback.")
	flag.Parse()
	hosts := getHostsFromFile(*hostsFileFlag)
	sshConfig := createSshConfig(*userFlag, *passwdFlag)
	resultCounts := 0
	interfacesData := make([]interfaceData, 0, len(hosts))
	runtime.GOMAXPROCS(workers)
	go populateQueue(jobsQueue, resultQueue, hosts, sshConfig)
	go parallelizeWorkers(jobsQueue)
	for {
		// FIXME make a case for timeout
		select {
		case jobResult := <-resultQueue:
			resultCounts++
			interfacesData = append(interfacesData, unpackJobResult(&jobResult)...)
			if resultCounts == len(hosts) {
				presentResults(interfacesData)
				return
			}
		}
	}
}
