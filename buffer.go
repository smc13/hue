package hue

import "strconv"

type buffer []byte

func (b *buffer) Write(p []byte) (n int, err error) {
	*b = append(*b, p...)
	return len(p), nil
}

func (b *buffer) WriteString(s string) (n int, err error) {
	return b.Write([]byte(s))
}

func (b *buffer) WriteQuoted(s string) (n int, err error) {
	return b.WriteString(strconv.Quote(s))
}
