package main

import (
	"github.com/zhaoyao/avalon/hprose"
	"sync/atomic"
)

type backendMode string

const (
	bcModeTCP  = backendMode("tcp")
	bcModeHTTP = backendMode("http") // TODO
)

type backend struct {
	mode     backendMode
	addr     string
	maxConns int
	conns    []*tcpBackendConn
	i        uint32
}

func newBackend(mode backendMode, addr string, maxConns int) (*backend, error) {
	bc := &backend{mode: mode, addr: addr, conns: make([]*tcpBackendConn, maxConns), maxConns: maxConns}

	for i := 0; i < maxConns; i++ {
		bc.conns[i] = &tcpBackendConn{addr: bc.addr, input: make(chan *hprose.Call, 1024)}
	}
	return bc, nil
}

func (bc *backend) Forward(c *hprose.Call) error {
	// TODO: load-balance
	v := int(atomic.AddUint32(&bc.i, 1)) % bc.maxConns
	bc.conns[v].input <- c
	return nil
}

func (bc *backend) Start() error {
	for _, c := range bc.conns {
		go c.Run()
	}
	return nil
}

func (bc *backend) Stop() {
	bc.conns[0].Close()
	for _, c := range bc.conns {
		c.Close()
	}
}
