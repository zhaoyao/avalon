package main

import (
	"fmt"
	hpio "github.com/hprose/hprose-golang/io"
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

type backends []*backend

func (bds backends) Addrs() []string {
	var fl []string
	for _, x := range bds {
		fl = append(fl, x.addr)
	}
	return fl
}

func newBackend(mode backendMode, addr string, maxConns int) (*backend, error) {
	bc := &backend{mode: mode, addr: addr, conns: make([]*tcpBackendConn, maxConns), maxConns: maxConns}

	for i := 0; i < maxConns; i++ {
		bc.conns[i] = &tcpBackendConn{addr: bc.addr, input: make(chan *hprose.Call, 1024)}
	}
	return bc, nil
}

func (bc *backend) ListFunction() ([]string, error) {
	c := &hprose.Call{
		FuncList: true,
	}

	c.AddWaiter()
	if err := bc.Forward(c); err != nil {
		return nil, err
	}
	c.Wait()

	r := c.R
	if !r.Success {
		return nil, fmt.Errorf("error fetch function list: %s", string(r.Error.S))
	}

	var fl []string
	err := hprose.ParseList(r.Body, func(tag byte, r *hprose.Reader) error {
		switch tag {
		case hpio.TagString:
			s, err := r.ReadStr()
			if err != nil {
				return err
			}

			fl = append(fl, string(s.S))
		case hpio.TagUTF8Char:
			var ret rune
			var err error
			if ret, err = r.ReadUTF8(); err != nil {
				return err
			}
			fl = append(fl, string(ret))
		}
		return nil
	})
	return fl, err
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
