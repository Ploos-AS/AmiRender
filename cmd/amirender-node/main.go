package main

import (
	"fmt"
	"net"
	"os"

	"github.com/Ploos-AS/AmiRender/internal/node"
)

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
	if err := node.Serve(ln); err != nil {
		panic(err)
	}
}
