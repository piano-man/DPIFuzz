package main

import (
	"os"
	"bufio"
	m "masterthesis"
	"masterthesis/scenario"
	"time"
	"os/exec"
	"encoding/json"
	"strings"
)

func GitCommit() string {
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "git"
	cmdArgs := []string{"rev-parse", "--verify", "HEAD"}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(cmdOut))
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	commit := GitCommit()

	scanner := bufio.NewScanner(file)
	results := make([]m.Trace, 0, 0)
	for scanner.Scan() {
		host := scanner.Text()
		trace := m.Trace{
			Commit:    commit,
			Host:      host,
			StartedAt: time.Now().Unix(),
			Results: make(map[string]interface{}),
		}

		conn := m.NewConnection(host, strings.Split(host, ":")[0])
		conn.ReceivedPacketHandler = func(data []byte) {
			trace.Stream = append(trace.Stream, m.TracePacket{Direction: m.ToClient, Timestamp: time.Now().Unix(), Data: data})
		}
		conn.SentPacketHandler = func(data []byte) {
			trace.Stream = append(trace.Stream, m.TracePacket{Direction: m.ToServer, Timestamp: time.Now().Unix(), Data: data})
		}

		start := time.Now()
		scenario.RunVersionNegotiationScenario(conn, &trace)
		trace.Duration = uint64(time.Now().Sub(start).Seconds() * 1000)
		trace.Ip = strings.Split(conn.ConnectedIp().String(), ":")[0]

		results = append(results, trace)
	}

	out, _ := json.Marshal(results)
	println(string(out))
}