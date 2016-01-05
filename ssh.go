// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eraclitux/stracer"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	remoteCommand = `cat /proc/net/dev; echo ZZZ; sleep 1; cat /proc/net/dev;`
	separator     = "ZZZ\n"
)

// rawData stores all parameters for a single interface
// at a given time.
// es: map[string]map[string]uint64{"eth0": map[string]uint64{"tx-Bps":12000, "rx-Bps":12000}}
type rawData map[string]map[string]uint64

// interfaceData models single interface's data for a given host.
type interfaceData struct {
	host  string
	name  string
	rates map[string]uint64
	err   error
}

// makeValueMap creates a map for a given interface
// and populates it with peculiar values (rx-Bps, rx-pps etc)
// at t2 or t1 instants.
func makeValueMap(data []string) (map[string]uint64, error) {
	dataMap := make(map[string]uint64)
	for i, s := range data {
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

// calculateRates uses values at instant t2 stored in dataAtT2
// by makeValueMap to calculate rates for a single interface.
// Calculated rates are than stored in dataAtT2.
func calculateRates(dataAtT2, dataAtT1 map[string]uint64) {
	for k, v := range dataAtT2 {
		// assuming that ΔT is always 1 second
		// (we use sleep 1 at remote host)
		dataAtT2[k] = v - dataAtT1[k]
	}
}

// parseOutput arranges remoteCommand output executed
// on a single remote host and calculates rates
// for all its network interfaces.
// It stores results in passed data var.
func parseOutput(out io.Reader, data rawData) error {
	all, err := ioutil.ReadAll(out)
	if err != nil {
		return err
	}
	outBytes := bytes.Split(all, []byte(separator))
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
		// rates are stored in passed data map
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
			// FIXME return error
			fmt.Fprintln(os.Stderr, err)
		} else {
			agentClient := agent.NewClient(socket)
			authMethod := ssh.PublicKeysCallback(agentClient.Signers)
			authMethods = append(authMethods, authMethod)
			stracer.Traceln("ssh-agent configured")
		}
	}
	return ssh.ClientConfig{
		User: user,
		Auth: authMethods,
	}
}

// assembleSSHClient helps to create an ssh client
// with a configurable timeout connection.
func assembleSSHClient(network, addr string, config *ssh.ClientConfig, timeout int) (*ssh.Client, error) {
	cNet, err := net.DialTimeout(network, addr, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}
	conn, newChan, reqChan, err := ssh.NewClientConn(cNet, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(conn, newChan, reqChan),
		nil
}
