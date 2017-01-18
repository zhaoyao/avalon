package main

import (
	//"github.com/apex/log"
	"github.com/zhaoyao/avalon/hprose"
	"github.com/zhaoyao/serverkit"
	"net"
)

type tcpConn struct {
	nc     net.Conn
	router *Router
}

func newTCPConn(r *Router) func(net.Conn) (serverkit.TCPConn, error) {
	return func(c net.Conn) (serverkit.TCPConn, error) {
		return newTCPConn0(c, r)
	}
}
func newTCPConn0(c net.Conn, r *Router) (serverkit.TCPConn, error) {
	//log.WithField("addr", c.RemoteAddr()).Info("new connection")

	return &tcpConn{nc: c, router: r}, nil
}

func (c *tcpConn) Serve() error {
	for {
		req, err := hprose.DecodeRequest(c.nc)
		if err != nil {
			//log.WithError(err).Info("decode request failed")
			return err
		}

		//log.Debugf("PROXY request: %+v", string(req.CallList[0].Func.S))

		// find backend
		// TODO full duplex
		req.CallList[0].AddWaiter()
		if err := c.router.Handle(req); err != nil {
			return err
		}

		req.CallList[0].Wait()

		if err := hprose.EncodeResponse(&hprose.Response{
			ResultList: []*hprose.Result{req.CallList[0].R},
		}, c.nc); err != nil {
			return err
		}
	}

}
