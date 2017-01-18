package main

import (
	"bufio"
	//"fmt"
	//"github.com/apex/log"
	"github.com/zhaoyao/avalon/hprose"
	"net"
	"sync"
	"time"
)

type tcpBackendConn struct {
	addr     string
	input    chan *hprose.Call
	stopOnce sync.Once
}

func (bc *tcpBackendConn) Run() {
	//log.Infof("backend conn [%p] to %s", bc, bc.addr)
	for k := 0; ; k++ {
		err := bc.mainLoop()
		if err == nil {
			break
		} else {
			for i := len(bc.input); i != 0; i-- {
				r := <-bc.input
				_ = bc.handleResponse(r, nil, err)
				//bc.setResponse(r, nil, err)
			}
		}
		//log.WithError(err).Warnf("backend conn [%p] to %s, restart [%d]", bc, bc.addr, k)
		time.Sleep(time.Millisecond * 50)
	}
	//log.Infof("backend conn [%p] to %s, stop and exit", bc, bc.addr)

}

func (bc *tcpBackendConn) Close() {
	bc.stopOnce.Do(func() {
		close(bc.input)
	})
}

func (bc *tcpBackendConn) mainLoop() error {
	r, ok := <-bc.input

	if ok {
		c, tasks, err := bc.connect()
		w := bufio.NewWriter(c)
		if err != nil {
			return bc.handleResponse(r, nil, err)
		}
		for ok {
			if err := hprose.EncodeCall(r, w, true); err != nil {
				return bc.handleResponse(r, nil, err)
			}
			w.Flush()
			tasks <- r
			r, ok = <-bc.input
		}
	}

	return nil
}

func (bc *tcpBackendConn) connect() (net.Conn, chan<- *hprose.Call, error) {
	println("connect backend")
	c, err := net.DialTimeout("tcp", bc.addr, 20*time.Second)
	if err != nil {
		return nil, nil, err
	}

	if tc, ok := c.(*net.TCPConn); ok {
		tc.SetNoDelay(true)
	}

	ch := make(chan *hprose.Call, 4096)
	go func() {
		defer c.Close()
		for req := range ch {
			//
			resp, err := hprose.DecodeResponse(c)
			//println("r resp")
			bc.handleResponse(req, resp, err)
		}
	}()

	return c, ch, nil
}

func (bc *tcpBackendConn) handleResponse(r *hprose.Call, resp *hprose.Response, err error) error {

	/*
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"backend": bc.addr,
				"req":     r,
				"resp":    resp,
			}).Debug("<-- resp")
		} else {
			log.WithFields(log.Fields{
				"backend": bc.addr,
				"req":     r,
				"resp":    resp,
			}).Debug("<-- resp")

		}
	*/

	/*
		result := resp.ResultList[0]
		if result.Success {
			//fmt.Println(string(result.Body))
		} else {
			//fmt.Println(string(result.Error.S))
		}
	*/

	if err != nil {
		r.SetResult(&hprose.Result{
			Success: false,
			Error:   &hprose.HStr{N: int64(len(err.Error())), S: []byte(err.Error())},
		})
	} else {
		r.SetResult(resp.ResultList[0])

	}
	return err
}
