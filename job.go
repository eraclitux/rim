package main

import (
	"bytes"

	"github.com/eraclitux/goparallel"
	"github.com/eraclitux/stracer"
	"golang.org/x/crypto/ssh"
)

type job struct {
	host            string
	sshClientConfig ssh.ClientConfig
	result          []interfaceData
	timeout         int
	err             error
}

func (j *job) Execute() {
	output := new(bytes.Buffer)
	destination := j.host
	sanitizeHost(&destination)

	conn, err := assembleSSHClient("tcp", destination, &j.sshClientConfig, j.timeout)

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
	session.Stdout = output
	if err := session.Run(remoteCommand); err != nil {
		j.result = packResult(j.host, err, nil)
		return
	}
	stracer.Traceln("Output:", output.String())
	data := make(rawData)
	if err := parseOutput(output, data); err != nil {
		j.result = packResult(j.host, err, nil)
		return
	}
	j.result = packResult(j.host, nil, data)
}

// packResult gets raw data for all network interfaces
// from an executed job and puts them into
// a slice of interfaceData, one for each network interface.
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

func makeTasks(hosts []string, sshConfig ssh.ClientConfig, timeout int) []goparallel.Tasker {
	tasks := make([]goparallel.Tasker, 0, len(hosts))
	for _, h := range hosts {
		// FIXME are 2 elements allocated in result really used
		// or result is just overwritten? In this case use 0 lenght
		j := job{
			host:            h,
			sshClientConfig: sshConfig,
			result:          make([]interfaceData, 2),
			timeout:         timeout,
		}
		tasks = append(tasks, &j)
	}
	return tasks
}
