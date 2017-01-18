package hprose

import (
	"sync"
)

type Request struct {
	Raw      []byte
	ID       []byte
	FuncList bool
	CallList []*Call
}

// Call 代表一次远程调用请求
type Call struct {
	Func       *HStr
	NumArgs    int
	RawArgList []byte
	FuncList   bool
	Origin     []byte

	wait *sync.WaitGroup
	R    *Result
}

func (c *Call) Wait() {
	c.wait.Wait()
}

func (c *Call) AddWaiter() {
	if c.wait == nil {
		c.wait = &sync.WaitGroup{}
	}
	c.wait.Add(1)
}

func (c *Call) SetResult(r *Result) {
	c.R = r
	c.wait.Done()
}

type Response struct {
	Raw        []byte
	ID         []byte
	ResultList []*Result
}

type Result struct {
	Success bool
	Error   *HStr
	Body    []byte // result and arg
	Origin  []byte
}

// hprose 中的 string 类型
type HStr struct {
	N int64
	S []byte
}
