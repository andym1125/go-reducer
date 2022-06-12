package main

import "errors"

type Reader struct {
	cursor int //data[cursor] is unconsumed
	data   []byte
	delims map[byte]bool
	len    int
}

func NewReader(data []byte) *Reader {
	return &Reader{
		cursor: 0,
		data:   data,
		delims: map[byte]bool{
			10: true,
			13: true,
		},
		len: len(data),
	}
}

func (r *Reader) Eof() bool {
	return r.len <= r.cursor
}

func (r *Reader) Readln() []byte {

	var ret []byte
	for i := r.cursor; i < r.len; i++ {

		//What if peeking causes an error somehow?
		c, err := r.PeekAt(i)
		if err != nil {
			return ret
		}

		//If we find delimiter, consume all delimiters and return
		if _, isDelim := r.delims[c]; isDelim {

			r.UnsafeRead()
			for j := r.cursor; j < r.len; j++ {

				jc, jerr := r.PeekAt(j)
				if err != nil {
					panic(jerr)
				}

				if _, isAlsoDelim := r.delims[jc]; isAlsoDelim {
					r.UnsafeRead()
				} else {
					break
				}
			}

			return ret
		}

		ret = append(ret, r.UnsafeRead())
	}
	return ret
}

func (r *Reader) Read() (byte, error) {

	if r.cursor >= len(r.data) {
		return 0, errors.New("Out of bounds")
	}

	r.cursor++
	return r.data[r.cursor-1], nil
}

func (r *Reader) UnsafeRead() byte {
	r.cursor++
	return r.data[r.cursor-1]
}

func (r *Reader) Peek() (byte, error) {

	if r.cursor >= len(r.data) {
		return 0, errors.New("Out of bounds")
	}

	return r.data[r.cursor], nil
}

func (r *Reader) PeekAt(idx int) (byte, error) {

	if idx >= len(r.data) {
		return 0, errors.New("Out of bounds")
	}

	return r.data[idx], nil
}
