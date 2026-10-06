package node

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/Ploos-AS/AmiRender/internal/engine"
	"github.com/Ploos-AS/AmiRender/internal/engine/povray"
	"github.com/Ploos-AS/AmiRender/internal/farm"
)

type message struct {
	Type string          `json:"type"`
	Job  json.RawMessage `json:"job"`
}

func Handle(c net.Conn) {
	defer c.Close()
	s := bufio.NewScanner(c)
	for s.Scan() {
		var m message
		if json.Unmarshal(s.Bytes(), &m) != nil || m.Type != "SUBMIT" {
			writeResult(c, farm.WorkerResult{Type: "FAILED", Error: "bad request"})
			continue
		}
		var j farm.RenderJob
		if json.Unmarshal(m.Job, &j) != nil {
			writeResult(c, farm.WorkerResult{Type: "FAILED", Error: "bad job"})
			continue
		}
		switch j.Engine {
		case "null":
			writeResult(c, farm.WorkerResult{
				Type: "COMPLETE", JobID: j.ID, Engine: "null", Output: j.Output,
			})
		case "povray":
			renderPOVRay(c, j)
		default:
			writeResult(c, farm.WorkerResult{
				Type: "FAILED", JobID: j.ID, Engine: j.Engine, Error: "unsupported engine",
			})
		}
	}
}

func renderPOVRay(c net.Conn, j farm.RenderJob) {
	e := povray.New("")
	result, err := e.Render(context.Background(), engine.Job{
		ID: j.ID, Scene: j.Scene, Output: j.Output, Frame: j.Frame, Width: j.Width, Height: j.Height,
	})
	if err != nil {
		writeResult(c, farm.WorkerResult{
			Type: "FAILED", JobID: j.ID, Engine: "povray", Error: err.Error(),
		})
		return
	}
	writeResult(c, farm.WorkerResult{
		Type: "COMPLETE", JobID: j.ID, Engine: "povray", Output: result.Output,
	})
}

func writeResult(c net.Conn, result farm.WorkerResult) {
	_ = json.NewEncoder(c).Encode(result)
}

func Serve(ln net.Listener) error {
	for {
		c, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}
		go Handle(c)
	}
}
