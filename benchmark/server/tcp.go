package main

import "github.com/hprose/hprose-golang/rpc"
import "time"

func hello(name string) string {
	//println("request", name)
	time.Sleep(1 * time.Millisecond)
	return "Hello " + name + "!"
}

func main() {
	server := rpc.NewTCPServer("tcp4://0.0.0.0:4321/")
	server.AddFunction("hello", hello)
	server.Start()
}
