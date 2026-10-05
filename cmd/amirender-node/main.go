package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

type Message struct {
	Type string          `json:"type"`
	Job  json.RawMessage `json:"job"`
}

type Job struct {
	ID     string `json:"id"`
	Engine string `json:"engine"`
	Output string `json:"output"`
}

func handle(c net.Conn) {
	defer c.Close()
	s := bufio.NewScanner(c)
	for s.Scan() {
		var m Message
		if json.Unmarshal(s.Bytes(), &m) != nil || m.Type != "SUBMIT" {
			fmt.Fprintln(c, `{"type":"FAILED","error":"bad request"}`)
			continue
		}
		var j Job
		if json.Unmarshal(m.Job, &j) != nil || j.Engine != "null" {
			fmt.Fprintln(c, `{"type":"FAILED","error":"unsupported engine"}`)
			continue
		}
		out, _ := json.Marshal(map[string]any{"type": "COMPLETE", "job_id": j.ID, "engine": "null", "output": j.Output})
		fmt.Fprintln(c, string(out))
	}
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:6800")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("AmiRender M0 node listening on 127.0.0.1:6800")
	for {
		c, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go handle(c)
	}
}
