package hprose

import (
	hpio "github.com/hprose/hprose-golang/io"
)

func ParseList(b []byte, f func(byte, *Reader) error) error {
	r := NewReader(b)
	if _, err := r.expect(hpio.TagList); err != nil {
		return err
	}

	n, err := r.readInt64(hpio.TagOpenbrace)
	if err != nil {
		return err
	}

	for i := 0; i < int(n); i++ {
		tag, err := r.Peek()
		if err != nil {
			return err
		}

		if err := f(tag, r); err != nil {
			return err
		}
	}
	if _, err := r.expect(hpio.TagClosebrace); err != nil {
		return err
	}

	return nil
}
