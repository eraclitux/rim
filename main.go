// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

// RIM - Remote Interfaces Monitor
// Agentless network interfaces monitor for Linux firewalls/servers
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/eraclitux/goparallel"
	"github.com/eraclitux/stracer"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var workers = runtime.NumCPU()

const (
	remoteCommand = `cat /proc/net/dev; echo ZZZ; sleep 1; cat /proc/net/dev;`
	separator     = "ZZZ\n"
)

// es: map[string]map[string]uint64{"eth0": map[string]uint64{"tx-Bps":12000, "rx-Bps":12000}}
type rawData map[string]map[string]uint64

// interfaceData models single interface' data for a given host.
type interfaceData struct {
	host  string
	name  string
	rates map[string]uint64
	err   error
}

type job struct {
	host            string
	sshClientConfig ssh.ClientConfig
	result          []interfaceData
	err             error
}

func (j *job) Execute() {
	var output bytes.Buffer
	destination := j.host
	sanitizeHost(&destination)
	conn, err := ssh.Dial("tcp", destination, &j.sshClientConfig)
	if err != nil {
		j.result = packResult(j.host, err, nil)
		return
	}
	session, err := conn.NewSession()
	if err != nil {
		j.result = packResult(j.host, err, nil)
		return
	}
	defer session.Close()
	session.Stdout = &output
	if err := session.Run(remoteCommand); err != nil {
		j.result = packResult(j.host, err, nil)
		return
	}
	stracer.Traceln("Output:", output.String())
	data := make(rawData)
	if err := parseOutput(&output, data); err != nil {
		j.result = packResult(j.host, err, nil)
		return
	}
	j.result = packResult(j.host, nil, data)
}

// FIXME docs unpackJobResult reads data from a jobResult and unpack it into a slice of interfaceData, one for each interface.
func packResult(host string, err error, data rawData) []interfaceData {
	result := []interfaceData{}
	if err != nil {
		return []interfaceData{interfaceData{host, "", nil, err}}
	}
	for keyInterface, valueMap := range data {
		i := interfaceData{host, keyInterface, valueMap, nil}
		result = append(result, i)
	}
	return result
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

// makeValueMap creates map with t2 or t1 values.
func makeValueMap(data []string) (map[string]uint64, error) {
	dataMap := make(map[string]uint64)
	for i, s := range data {
		// FIXME handle convertion error
		converted, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		switch i {
		case 0:
			// here it's not bytes per second but absolute bytes at t2
			// however we'll store Bps later
			dataMap["rx-Bps"] = converted
		case 1:
			// here it's not packets per second but absolute bytes at t2
			// however we'll store pps later
			dataMap["rx-pps"] = converted
		case 2:
			// here it's not errors per second but absolute bytes at t2
			// however we'll store eps later
			dataMap["rx-eps"] = converted
		case 3:
			// here it's not drop per second but absolute bytes at t2
			// however we'll store dps later
			dataMap["rx-dps"] = converted
		case 8:
			// here it's not bytes per second but absolute bytes at t2
			// however we'll store Bps later
			dataMap["tx-Bps"] = converted
		case 9:
			// here it's not bytes per second but absolute bytes at t2
			// however we'll store Bps later
			dataMap["tx-pps"] = converted
		case 10:
			// here it's not errors per second but absolute bytes at t2
			// however we'll store eps later
			dataMap["tx-eps"] = converted
		case 11:
			// here it's not drops per second but absolute bytes at t2
			// however we'll store dps later
			dataMap["tx-dps"] = converted
		}
	}
	return dataMap, nil
}

// calculateRates uses t2 values stored in dataAtT2 by makeValueMap to calculate rates.
func calculateRates(dataAtT2, dataAtT1 map[string]uint64) {
	for k, v := range dataAtT2 {
		// assuming that ΔT is always 1 second (sleep 1)
		dataAtT2[k] = v - dataAtT1[k]
	}
}

// parseOutput arranges RemoteCommand output and calculates rates.
func parseOutput(out *bytes.Buffer, data map[string]map[string]uint64) error {
	outBytes := bytes.Split(out.Bytes(), []byte(separator))
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
		stracer.Traceln("parsed data @ t2:", iface, countersData)
		var err error
		data[iface], err = makeValueMap(countersData)
		if err != nil {
			return err
		}
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
		dataAtT1, err := makeValueMap(countersData)
		if err != nil {
			return err
		}
		stracer.Traceln("parsed data @ t1:", iface, countersData)
		calculateRates(data[iface], dataAtT1)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func createSSHConfig(user, passwd string) ssh.ClientConfig {
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	authMethods := []ssh.AuthMethod{ssh.Password(passwd)}
	if sshAuthSock != "" {
		stracer.Traceln("ssh-agent socket:", sshAuthSock)
		socket, err := net.Dial("unix", sshAuthSock)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			agentClient := agent.NewClient(socket)
			authMethod := ssh.PublicKeysCallback(agentClient.Signers)
			authMethods = append(authMethods, authMethod)
			// FIXME works even without calling agent.ForwardToAgent()?
			stracer.Traceln("ssh-agent configured")
		}
	}
	return ssh.ClientConfig{
		User: user,
		Auth: authMethods,
	}
}

func getHostsFromFile(path string) []string {
	bytes := []byte{}
	var err error
	if path == "{filename}" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(path)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input", err)
		// TODO just return an error and move this inside main
		os.Exit(2)
	}
	hosts := strings.Split(string(bytes), "\n")
	// remove last empty element
	hosts = hosts[:len(hosts)-1]
	stracer.Traceln("Parsed hosts:", hosts)
	return hosts
}
func makeTasks(hosts []string, tasks []goparallel.Tasker, sshConfig ssh.ClientConfig) []goparallel.Tasker {
	for _, h := range hosts {
		j := job{host: h, sshClientConfig: sshConfig, result: make([]interfaceData, 2)}
		tasks = append(tasks, &j)
	}
	return tasks
}

func main() {
	hostsFileFlag := flag.String("f", "{filename}", " [FILE] file containing target hosts, one per line, formatted as <hostname>[:port].")
	userFlag := flag.String("u", "root", "[USERNAME] ssh username.")
	passwdFlag := flag.String("p", "nopassword", "[PASSWORD] ssh password for remote hosts. Automatically use ssh-agent as fallback.")
	sortFlag1 := flag.String("k1", "rx-dps", "first sort key.")
	sortFlag2 := flag.String("k2", "rx-Kbps", "second sort key.")
	limitFlag := flag.Int("l", 0, "limit printed results to this number, 0 means no limits.")
	noHeadFlag := flag.Bool("n", false, "do not show titles in output.")
	extendedFlag := flag.Bool("e", false, "enable extended output.")
	versionFlag := flag.Bool("v", false, "show version and exit.")
	flag.Parse()
	if *versionFlag {
		fmt.Println("RIM - Remote Interfaces Monitor v2.0.0-alfa")
		return
	}
	sortKeys, err := sanitizeSortKeys(*sortFlag1, *sortFlag2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	hosts := getHostsFromFile(*hostsFileFlag)
	sshConfig := createSSHConfig(*userFlag, *passwdFlag)

	tasks := make([]goparallel.Tasker, 0, len(hosts))
	interfacesData := make([]interfaceData, 0, len(hosts))
	tasks = makeTasks(hosts, tasks, sshConfig)
	goparallel.RunBlocking(tasks)
	for _, t := range tasks {
		interfacesData = append(interfacesData, t.(*job).result...)
	}
	// We could use an arbitrary number of sort keys.
	orderBy(byKey(sortKeys[0]), byKey(sortKeys[1])).sort(interfacesData)
	s := []interfaceData{}
	if *limitFlag == 0 {
		s = interfacesData
	} else {
		s = interfacesData[:*limitFlag]
	}
	displayResults(s, *noHeadFlag, *extendedFlag)
}
