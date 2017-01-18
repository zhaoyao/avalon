package main

import (
	"github.com/zhaoyao/avalon/hprose"
)

type Router struct {
	bc *backend
}

func newRouter() *Router {
	bc, _ := newBackend(bcModeTCP, "127.0.0.1:4321", 10)
	return &Router{bc: bc}
}

func (r *Router) Start() error {
	return r.bc.Start()
}

func (r *Router) Stop() {
	r.bc.Stop()
}

func (r *Router) Handle(req *hprose.Request) error {
	// TODO: select backend
	return r.bc.Forward(req.CallList[0])
}
