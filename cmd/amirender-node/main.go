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
	"github.com/Ploos-AS/AmiRender/internal/farm"
)

type Message struct {
	Type string          `json:"type"`
	Job  json.RawMessage `json:"job"`
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
		var j farm.RenderJob
		if json.Unmarshal(m.Job, &j) != nil {
			fmt.Fprintln(c, `{"type":"FAILED","error":"bad job"}`)
			continue
		}

		if j.Engine == "null" {
			writeResult(c, farm.WorkerResult{
				Type: "COMPLETE", JobID: j.ID, Engine: "null", Output: j.Output,
			})
			continue
		}
		if j.Engine == "povray" {
			e := povray.New("")
			result, err := e.Render(context.Background(), engine.Job{
				ID: j.ID, Scene: j.Scene, Output: j.Output, Frame: j.Frame, Width: j.Width, Height: j.Height,
			})
			if err != nil {
				writeResult(c, farm.WorkerResult{
					Type: "FAILED", JobID: j.ID, Engine: "povray", Error: err.Error(),
				})
				continue
			}
			writeResult(c, farm.WorkerResult{
				Type: "COMPLETE", JobID: j.ID, Engine: "povray", Output: result.Output,
			})
			continue
		}
		writeResult(c, farm.WorkerResult{
			Type: "FAILED", JobID: j.ID, Engine: j.Engine, Error: "unsupported engine",
		})
	}
}

func writeResult(c net.Conn, result farm.WorkerResult) {
	_ = json.NewEncoder(c).Encode(result)
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
