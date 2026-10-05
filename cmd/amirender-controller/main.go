package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Job struct {
	Version int                    `json:"version"`
	ID string                      `json:"id"`
	Engine string                  `json:"engine"`
	Scene string                   `json:"scene"`
	Frame int                      `json:"frame"`
	Width int                      `json:"width"`
	Height int                     `json:"height"`
	Output string                  `json:"output"`
	Priority int                   `json:"priority"`
	EngineOptions map[string]any   `json:"engine_options"`
}

func main() {
	addr := "127.0.0.1:6800"
	if len(os.Args) > 1 { addr = os.Args[1] }
	conn, err := net.Dial("tcp", addr)
	if err != nil { panic(err) }
	defer conn.Close()

	job := Job{Version:1, ID:"m0-null-frame-0001", Engine:"null", Scene:"tests/scenes/minimal", Frame:0, Width:320, Height:256, Output:"frame0000.iff", Priority:50, EngineOptions:map[string]any{}}
	if err := json.NewEncoder(conn).Encode(map[string]any{"type":"SUBMIT","job":job}); err != nil { panic(err) }
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil { panic(err) }
	fmt.Print(line)
}
