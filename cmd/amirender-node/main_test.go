package main

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestNullEngineEndToEnd(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err == nil {
			handle(c)
		}
	}()

	conn, err := net.DialTimeout("tcp", ln.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	job := map[string]any{
		"version": 1,
		"id":      "m0-e2e-0001",
		"engine":  "null",
		"scene":   "minimal",
		"frame":   0,
		"width":   320,
		"height":  256,
		"output":  "frame0000.iff",
	}
	if err := json.NewEncoder(conn).Encode(map[string]any{"type": "SUBMIT", "job": job}); err != nil {
		t.Fatal(err)
	}

	var reply map[string]any
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&reply); err != nil {
		t.Fatal(err)
	}
	if reply["type"] != "COMPLETE" {
		t.Fatalf("expected COMPLETE, got %v", reply)
	}
	if reply["job_id"] != "m0-e2e-0001" {
		t.Fatalf("wrong job id: %v", reply["job_id"])
	}
	if reply["engine"] != "null" {
		t.Fatalf("wrong engine: %v", reply["engine"])
	}
}
