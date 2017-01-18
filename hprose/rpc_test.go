package hprose

import (
	"bytes"
	"testing"
)

func parseCall(t *testing.T, s string) *Request {
	r := newReader([]byte(s))

	ret, err := decodeRequest(r)

	if err != nil {
		t.Fatal(err)
	}

	return ret
}

func parseResponse(t *testing.T, s string) *Response {
	r := newReader([]byte(s))

	ret, err := decodeResponse(r)

	if err != nil {
		t.Fatal(err)
	}

	return ret
}

func TestFuncList(t *testing.T) {
	ret := parseCall(t, "z")

	if len(ret.CallList) != 0 {
		t.Fatal("expected 1 call")
	}

	if !ret.FuncList {
		t.Fatal("expected FuncList")
	}
}

func TestNoArgCall(t *testing.T) {
	req := parseCall(t, `Cs5"hello"z`)
	ret := req.CallList

	if len(ret) != 1 {
		t.Fatalf("expected 1 call: %d", len(ret))
	}

	if ret[0].FuncList {
		t.Fatal("expected no FuncList")
	}

	name := ret[0].Func.S
	if string(name) != "hello" {
		t.Fatalf("expected func name hello: %s", name)
	}
	if len(ret[0].RawArgList) != 0 {
		t.Fatal("expected not arg")
	}
}

func TestCallWithArg(t *testing.T) {
	req := parseCall(t, `Cs5"hello"a1{s5"world"}z`)
	ret := req.CallList

	if len(ret) != 1 {
		t.Fatalf("expected 1 call: %d", len(ret))
	}

	if ret[0].FuncList {
		t.Fatal("expected no FuncList")
	}

	name := string(ret[0].Func.S)
	if name != "hello" {
		t.Fatalf("expected func name hello: %s", name)
	}

	if ret[0].NumArgs != 1 {
		t.Fatal("num arg expect 1")
	}

	if bytes.Compare(ret[0].RawArgList, []byte(`s5"world"`)) != 0 {
		t.Fatalf("expected arg: %s", string(ret[0].RawArgList))
	}
}

func TestBatchCall(t *testing.T) {
	req := parseCall(t, `Cs5"hello"a1{s5"world"}Cs5"hello"a1{s5"world1"}z`)
	ret := req.CallList

	if len(ret) != 2 {
		t.Fatalf("expected 2 calls: %d", len(ret))
	}

	if ret[0].FuncList {
		t.Fatal("expected no FuncList")
	}

	if ret[0].Func.N != 5 {
		t.Fatalf("unexpected func len: %d", ret[0].Func.N)
	}
	name := string(ret[0].Func.S)
	if name != "hello" {
		t.Fatalf("expected func name hello: %s", name)
	}
	if string(ret[0].RawArgList) != `s5"world"` {
		t.Fatalf("expected arg: %s", string(ret[0].RawArgList))
	}

	if ret[1].FuncList {
		t.Fatal("expected no FuncList")
	}

	name = string(ret[1].Func.S)
	if name != "hello" {
		t.Fatalf("expected func name hello: %s", name)
	}
	if bytes.Compare(ret[1].RawArgList, []byte(`s5"world1"`)) != 0 {
		t.Fatalf("expected arg: %s", string(ret[1].RawArgList))
	}
}

func TestDecodeSuccess(t *testing.T) {
	resp := parseResponse(t, `Rs12"Hello world!"z`)

	if len(resp.ResultList) != 1 {
		t.Fatalf("expected 1 result: %d", len(resp.ResultList))
	}

	result := resp.ResultList[0]
	if !result.Success {
		t.Fatal("expected success result")
	}

	if bytes.Compare(result.Body, []byte(`s12"Hello world!"`)) != 0 {
		t.Fatalf("expected body: %s", string(result.Body))
	}
}

func TestDecodeError(t *testing.T) {
	resp := parseResponse(t, `Es24"This is a error example."z`)

	if len(resp.ResultList) != 1 {
		t.Fatalf("expected 1 result: %d", len(resp.ResultList))
	}

	result := resp.ResultList[0]
	if result.Success {
		t.Fatal("expected error result")
	}

	if result.Body != nil {
		t.Fatal("expected no body")
	}

	if string(result.Error.S) != "This is a error example." {
		t.Fatalf("unexpected error: '%s'", string(result.Error.S))
	}
}

func TestDecodeMulti(t *testing.T) {
	resp := parseResponse(t, `Rs12"Hello world!"Es24"This is a error example."z`)

	if len(resp.ResultList) != 2 {
		t.Fatalf("expected 2 result: %d", len(resp.ResultList))
	}

	result := resp.ResultList[0]
	if !result.Success {
		t.Fatal("expected success result")
	}

	result = resp.ResultList[1]
	if result.Success {
		t.Fatal("expected error result")
	}

	if result.Body != nil {
		t.Fatal("expected no body")
	}

	if string(result.Error.S) != "This is a error example." {

		t.Fatalf("unexpected error: '%s'", string(result.Error.S))
	}
}
