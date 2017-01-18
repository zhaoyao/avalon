package hprose

import (
	"fmt"
	hpio "github.com/hprose/hprose-golang/io"
	"io"
)

type reader struct {
	b   []byte
	pos int
}

func newReader(b []byte) *reader {
	return &reader{b: b, pos: 0}
}

func (r *reader) Read() (byte, error) {
	if r.pos >= len(r.b) {
		return 0, io.EOF
	}

	b := r.b[r.pos]
	r.pos++
	return b, nil
}

func (r *reader) Peek() (byte, error) {
	if r.pos >= len(r.b) {
		return 0, io.EOF
	}

	return r.b[r.pos], nil
}

func (r *reader) ReadStr() (*HStr, error) {
	if _, err := r.expect(hpio.TagString); err != nil {
		return nil, err
	}
	len, err := r.readInt64(hpio.TagQuote)
	if err != nil {
		return nil, err
	}
	start := r.pos
	if err := r.Skip(int(len) + 1); err != nil { // strip ending "
		return nil, err
	}
	return &HStr{N: len, S: r.b[start : r.pos-1]}, nil
}

func (r *reader) Skip(n int) error {
	if r.pos+n >= len(r.b) {
		return io.EOF
	}
	r.pos += n
	return nil
}

func (r *reader) SkipUtil(tags ...byte) (byte, error) {
	for r.pos < len(r.b) {
		b := r.b[r.pos]
		for _, t := range tags {
			if b == t {
				return b, nil
			}
		}
		r.pos++
	}
	return 0, io.EOF
}

func (r *reader) expect(tags ...byte) (byte, error) {
	b, err := r.Read()
	if err != nil {
		return 0, err
	}

	for _, t := range tags {
		if b == t {
			return b, nil
		}
	}
	return 0, fmt.Errorf("hprose: unexpected tag %d, %v expected", b, tags)
}

func (r *reader) readInt64(endTag byte) (int64, error) {
	b, err := r.Read()
	if err != nil {
		return 0, err
	}
	if b == endTag {
		return 0, nil
	}

	i := int64(0)
	neg := int64(1)
	if b == '-' {
		neg = -1
		b, err = r.Read()
		if err != nil {
			return 0, err
		}
	}

	for ; err == nil && b != endTag; b, err = r.Read() {
		i = i*10 + int64(b-'0')*neg
	}
	return i, err
}

/*
func (r *reader) utf8CharLen() (int, error) {
	b, err := r.Peek()
	if err != nil {
		return 0, err
	}

	switch b >> 4 {
	case 0, 1, 2, 3, 4, 5, 6, 7:
		return 1, nil
	case 12, 13:
		return 2, nil
	case 14:
		return 3, nil
	case 15:
		if b&8 == 8 {
			return 0, errors.New("bad utf-8 encoding")
		}
		return 4, nil
	default:
		return 0, errors.New("bad utf-8 encoding")
	}
}

func (r *reader) resolveTypeLen(tag byte) (len int, err error) {
	if tag >= '0' && tag <= '9' {
		return 0, nil
	}
	switch tag {
	case hpio.TagInteger:
		fallthrough
	case hpio.TagLong:
		i, err := r.readInt64(hpio.TagSemicolon)
		return int(i), err

	case hpio.TagDouble:
		b, err := r.Peek()
		if err != nil {
			return 0, err
		}
		switch b {
		case hpio.TagNaN:
			return 1, nil
		case hpio.TagInfinity:
			return 2, nil

		}
		i, err := r.readInt64(hpio.TagSemicolon)
		return int(i), err

	case hpio.TagBytes:
		i, err := r.readInt64(hpio.TagQuote)
		if err != nil {
			return 0, err
		}
		if i > 0 {
			i++ // end quotes
		}
		return int(i), err

	case hpio.TagString:
		utf8Chars, err := r.readInt64(hpio.TagQuote)
		if err != nil {
			return 0, err
		}
		for i := 0; i < utf8Chars; i++ {

		}

	case hpio.TagTrue:
		fallthrough
	case hpio.TagFalse:
		fallthrough
	case hpio.TagEmpty:
		fallthrough
	case hpio.TagNull:
		return 0, nil

	case hpio.TagDate:
	case hpio.TagTime:
	case hpio.TagUTF8Char:
		return r.utf8CharLen()

	case hpio.TagGUID:
	case hpio.TagList:
	case hpio.TagMap:
	case hpio.TagClass:
	case hpio.TagObject:
	case hpio.TagRef:
	}

	return 0, nil
}

*/
