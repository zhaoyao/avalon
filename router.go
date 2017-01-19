package main

import (
	"errors"
	"github.com/apex/log"
	"github.com/zhaoyao/avalon/hprose"
	"strings"
	"sync"
	"time"
)

type Router struct {
	backends map[string]*backend
	fmap     map[string]backends
	rl       sync.RWMutex
	rt       *time.Ticker
}

func newRouter() *Router {
	return &Router{backends: make(map[string]*backend)}
}

func (r *Router) AddBackend(name string, b *backend) error {
	if _, ok := r.backends[name]; ok {
		return errors.New("router: duplicated backend")
	}
	r.backends[name] = b
	log.WithFields(log.Fields{
		"name": name,
		"addr": b.addr,
	}).Info("[router] Backend added")
	return nil
}

func (r *Router) Start() error {
	for _, b := range r.backends {
		if err := b.Start(); err != nil {
			return err
		}
	}

	// TODO: 抽取 registery
	if err := r.refreshRoutes(); err != nil {
		return err
	}

	r.rt = time.NewTicker(5 * time.Minute)
	go func() {
		for _ = range r.rt.C {
			_ = r.refreshRoutes()
		}
	}()
	return nil
}

func (r *Router) Stop() {
	r.rt.Stop()
	for _, b := range r.backends {
		b.Stop()
	}
}

func (r *Router) Handle(req *hprose.Request) error {
	call := req.CallList[0]
	fname := strings.ToLower(string(call.Func.S))

	if req.FuncList {
		// call func list
		call.SetResult(&hprose.Result{
			Success: false,
			Error:   &hprose.HStr{N: 5, S: []byte("method not found")},
		})
		return nil
	}

	r.rl.RLock()
	defer r.rl.RUnlock()

	b, ok := r.fmap[fname]
	if !ok {
		// no mapping
		call.SetResult(&hprose.Result{
			Success: false,
			Error:   &hprose.HStr{N: 5, S: []byte("method not found")},
		})
		return nil
	} else {

		// TODO: load-balancer
		return b[0].Forward(call)

	}
}

func (r *Router) refreshRoutes() (err error) {
	defer log.Trace("[router] Refresh route table").Stop(&err)
	m := make(map[string]backends)

	for _, b := range r.backends {
		funcs, err := fetchFunc(b)
		if err != nil {
			log.WithError(err).WithField("backend", b.addr).Info("Failed to fetch funcs")
			continue
		}

		for _, f := range funcs {
			if f == "#" {
				continue
			}
			fl := m[f]
			m[f] = append(fl, b)
		}
	}

	r.rl.Lock()
	r.fmap = m
	r.rl.Unlock()

	for f, bcs := range r.fmap {
		log.WithFields(log.Fields{
			"func":     f,
			"backends": bcs.Addrs(),
		}).Debugf("[router] Route item")

	}
	return nil
}

func fetchFunc(b *backend) (funcs []string, err error) {
	defer log.WithField("backend", b.addr).Trace("[router] Fetch function list").Stop(&err)
	funcs, err = b.ListFunction()
	if err != nil {
		return nil, err
	}
	return funcs, err
}
