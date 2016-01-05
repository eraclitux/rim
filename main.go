// Copyright (c) 2014 Andrea Masi. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE.txt file.

// RIM - Remote Interfaces Monitor
// Agentless network interfaces monitor for Linux firewalls/servers
// It uses ssh to get data from remote hosts.
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/eraclitux/cfgp"
	"github.com/eraclitux/goparallel"
	"github.com/eraclitux/stracer"
)

var Version = "unknown-build"
var BuildTime = "unknown-time"

type configuration struct {
	HostsFile string `cfgp:"f,[FILE] file containing target hosts. One per line formatted as <hostname>[:port],"`
	User      string `cfgp:"u,[USERNAME] ssh username,"`
	Passwd    string `cfgp:"p,[PASSWORD] ssh password for remote hosts. Automatically use ssh-agent as fallback,"`
	Sort1     string `cfgp:"k1,first sort key,"`
	Sort2     string `cfgp:"k2,second sort key,"`
	Limit     int    `cfgp:"l,limit printed results to this number; 0 means no limits,"`
	NoHead    bool   `cfgp:"n,do not show titles in output,"`
	Extended  bool   `cfgp:"e,enable extended output,"`
	Version   bool   `cfgp:"v,show version and exit,"`
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

func getHostsFromFile(path string) ([]string, error) {
	bytes := []byte{}
	var err error
	if path == "" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	hosts := strings.Split(string(bytes), "\n")
	// remove last empty element
	hosts = hosts[:len(hosts)-1]
	if len(hosts) == 0 {
		return nil, errors.New("zero hosts")
	}
	stracer.Traceln("Parsed hosts:", hosts)
	return hosts, nil
}

func main() {
	// Set default values.
	conf := configuration{
		User:   "root",
		Passwd: "nopassword",
		Sort1:  "rx-dps",
		Sort2:  "rx-Kbps",
	}
	cfgp.Path = os.Getenv("RIM_CONF_FILE")
	stracer.Traceln("conf file path:", os.Getenv("RIM_CONF_FILE"))
	cfgp.Parse(&conf)

	if conf.Version {
		fmt.Println("RIM - Remote Interfaces Monitor", Version, BuildTime)
		return
	}
	sortKeys, err := sanitizeSortKeys(conf.Sort1, conf.Sort2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	hosts, err := getHostsFromFile(conf.HostsFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting hosts:", err)
		os.Exit(2)
	}
	sshConfig := createSSHConfig(conf.User, conf.Passwd)
	interfacesData := make([]interfaceData, 0, len(hosts))
	tasks := makeTasks(hosts, sshConfig)
	goparallel.RunBlocking(tasks)
	for _, t := range tasks {
		interfacesData = append(interfacesData, t.(*job).result...)
	}
	// We could use an arbitrary number of sort keys.
	orderBy(byKey(sortKeys[0]), byKey(sortKeys[1])).sort(interfacesData)
	s := []interfaceData{}
	if conf.Limit == 0 {
		s = interfacesData
	} else {
		s = interfacesData[:conf.Limit]
	}
	displayResults(s, conf.NoHead, conf.Extended)
}
