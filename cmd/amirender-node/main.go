package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/Ploos-AS/AmiRender/internal/engine"
	"github.com/Ploos-AS/AmiRender/internal/engine/povray"
)

type Message struct {
	Type string          `json:"type"`
	Job  json.RawMessage `json:"job"`
}

type Job struct {
	ID     string `json:"id"`
	Engine string `json:"engine"`
	Scene  string `json:"scene"`
	Output string `json:"output"`
	Frame  int    `json:"frame"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
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
		if json.Unmarshal(m.Job, &j) != nil {
			fmt.Fprintln(c, `{"type":"FAILED","error":"bad job"}`)
			continue
		}

		if j.Engine == "null" {
			out, _ := json.Marshal(map[string]any{"type": "COMPLETE", "job_id": j.ID, "engine": "null", "output": j.Output})
			fmt.Fprintln(c, string(out))
			continue
		}
		if j.Engine == "povray" {
			e := povray.New("")
			result, err := e.Render(context.Background(), engine.Job{
				ID: j.ID, Scene: j.Scene, Output: j.Output, Frame: j.Frame, Width: j.Width, Height: j.Height,
			})
			if err != nil {
				out, _ := json.Marshal(map[string]any{"type": "FAILED", "job_id": j.ID, "engine": "povray", "error": err.Error()})
				fmt.Fprintln(c, string(out))
				continue
			}
			out, _ := json.Marshal(map[string]any{"type": "COMPLETE", "job_id": j.ID, "engine": "povray", "output": result.Output})
			fmt.Fprintln(c, string(out))
			continue
		}
		fmt.Fprintln(c, `{"type":"FAILED","error":"unsupported engine"}`)
	}
}

func main() {
	addr := "127.0.0.1:6800"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Printf("AmiRender node listening on %s\n", addr)
	for {
		c, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go handle(c)
	}
}
