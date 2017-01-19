package hprose

import (
	//"strings"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	//"github.com/apex/log"
	hpio "github.com/hprose/hprose-golang/io"
	"io"
)

/*
func decode(mode int, buf []byte) error {
	r := hpio.NewReader(buf, false)
	tag, err := r.ReadByte()
	if err != nil {
		return err
	}

	r.UnreadByte()
	switch {
	case tag == hpio.TagCall || tag == hpio.TagEnd:
		decodeCall(r)

	}

	return nil
}

*/

func DecodeRequest(r io.Reader) (*Request, error) {
	buf := make([]byte, 2048)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]

	pl := uint32(0)
	if err := binary.Read(bytes.NewBuffer(buf[:4]), binary.BigEndian, &pl); err != nil {
		return nil, err
	}

	b := buf[4:]
	dr := NewReader(b)
	req, err := decodeRequest(dr)

	if dr.pos != len(b) {
		return nil, errors.New("bad request")
	}

	//log.Debugf("--> Request: %s", string(b))
	return req, err
}

func EncodeRequest(r *Request, w io.Writer) error {
	return nil
}

func EncodeCall(c *Call, w io.Writer, withLen bool) error {
	b := bytes.NewBuffer(make([]byte, 0, 512))

	if !c.FuncList {
		b.WriteByte(hpio.TagCall)
		// func name
		b.WriteByte(hpio.TagString)
		fmt.Fprintf(b, "%d", c.Func.N)
		b.WriteByte(hpio.TagQuote)
		b.Write(c.Func.S)
		b.WriteByte(hpio.TagQuote)

		if len(c.RawArgList) > 0 {
			b.WriteByte(hpio.TagList)
			fmt.Fprintf(b, "%d", c.NumArgs)
			b.WriteByte(hpio.TagOpenbrace)
			b.Write(c.RawArgList)
			b.WriteByte(hpio.TagClosebrace)
		}
	}
	b.WriteByte(hpio.TagEnd)

	if withLen {
		if err := binary.Write(w, binary.BigEndian, int32(b.Len())); err != nil {
			return err
		}
	}

	_, err := b.WriteTo(w)
	return err
}

func DecodeResponse(r io.Reader) (*Response, error) {
	buf := make([]byte, 2048)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]
	pl := uint32(0)
	if err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &pl); err != nil {
		return nil, err
	}

	b := buf[4:]
	dr := NewReader(b)
	resp, err := decodeResponse(dr)

	if err != nil {
		return nil, err
	}

	if dr.pos != len(b) {
		return nil, fmt.Errorf("bad response: not fully consumed, total: %d pos: %d", len(b), dr.pos)
	}

	//log.Debugf("--> Response: %s", string(b))
	return resp, nil
}

func EncodeResponse(r *Response, w io.Writer) error {
	b := bytes.NewBuffer(make([]byte, 0, 512))

	for _, result := range r.ResultList {
		if result.Success {
			b.WriteByte(hpio.TagResult)
			b.Write(result.Body)
		} else {
			b.WriteByte(hpio.TagError)

			b.WriteByte(hpio.TagString)
			fmt.Fprintf(b, "%d", result.Error.N)
			b.WriteByte(hpio.TagQuote)
			b.Write(result.Error.S)
			b.WriteByte(hpio.TagQuote)
		}
	}

	b.WriteByte(hpio.TagEnd)

	if err := binary.Write(w, binary.BigEndian, int32(b.Len())); err != nil {
		return err
	}

	_, err := b.WriteTo(w)
	//log.Debugf("<-- Response: %s", string(b.Bytes()))
	return err
}

func decodeRequest(r *Reader) (*Request, error) {
	var (
		tag      byte
		err      error
		callList []*Call
	)
	for {
		tag, err = r.expect(hpio.TagCall, hpio.TagEnd)
		if err != nil {
			return nil, err
		}

		if tag == hpio.TagEnd {
			break
		}

		funcName, err := r.ReadStr()
		if err != nil {
			return nil, err
		}
		call := &Call{
			Func:   funcName,
			Origin: r.b,
		}
		callList = append(callList, call)

		tag, err = r.expect(hpio.TagCall, hpio.TagEnd, hpio.TagList)
		if err != nil {
			break
		}

		if tag == hpio.TagEnd {
			break
		}

		if tag == hpio.TagCall {
			r.pos--
			continue
		}

		// hpio.TagList
		n, err := r.readInt64(hpio.TagOpenbrace)
		if err != nil {
			break
		}
		call.NumArgs = int(n)

		argStart := r.pos
		if _, err = r.SkipUtil(hpio.TagClosebrace); err != nil {
			return nil, err
		}
		call.RawArgList = r.b[argStart:r.pos]
		if err = r.Skip(1); err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	if len(callList) == 0 {
		return &Request{Raw: r.b, FuncList: true}, nil
	}
	return &Request{Raw: r.b, CallList: callList}, nil
}

func decodeResponse(r *Reader) (*Response, error) {
	var (
		tag        byte
		err        error
		resultList []*Result
	)
	for {
		tag, err = r.expect(hpio.TagResult, hpio.TagError, hpio.TagEnd, hpio.TagFunctions)
		if err != nil {
			return nil, err
		}

		if tag == hpio.TagEnd {
			break
		}

		if tag == hpio.TagFunctions {
			body := r.b[r.pos:]
			r.pos = len(r.b)
			return &Response{
				Raw: r.b,
				ResultList: []*Result{
					&Result{
						Success: true,
						Body:    body,
						Origin:  body,
					},
				},
			}, nil
		}

		result := &Result{Origin: r.b}
		resultList = append(resultList, result)

		result.Success = tag == hpio.TagResult
		if tag == hpio.TagError {
			result.Error, err = r.ReadStr()
			if err != nil {
				return nil, err
			}
			continue
		}

		// TagResult
		bodyStart := r.pos
		tag, err = r.SkipUtil(hpio.TagResult, hpio.TagError, hpio.TagEnd)
		if err != nil {
			return nil, err
		}

		result.Body = r.b[bodyStart:r.pos]
	}

	if err != nil {
		return nil, err
	}

	if len(resultList) == 0 {
		return nil, errors.New("hprose: invalid response, empty result")
	}

	return &Response{Raw: r.b, ResultList: resultList}, nil
}
