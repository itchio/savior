package checker

import (
	"fmt"
	"io"
)

type Checker struct {
	reference []byte
	offset    int64
}

var _ io.WriteSeeker = (*Checker)(nil)

func New(reference []byte) *Checker {
	return &Checker{
		reference: reference,
	}
}

func (c *Checker) Write(buf []byte) (int, error) {
	n := 0
	for i := 0; i < len(buf); i++ {
		if c.offset >= int64(len(c.reference)) {
			return n, fmt.Errorf("out of bounds write: %d but max length is %d", c.offset, len(c.reference))
		}

		expected := c.reference[c.offset]
		actual := buf[i]
		if expected != actual {
			return n, fmt.Errorf("at byte %d, expected %x but got %x", c.offset, expected, actual)
		}
		c.offset++
		n++
	}
	return n, nil
}

func (c *Checker) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart {
		return c.offset, fmt.Errorf("unsupported whence value %d", whence)
	}

	if offset > int64(len(c.reference)) {
		return c.offset, fmt.Errorf("out of bounds seek: %d but max length is %d", c.offset, len(c.reference))
	}
	if offset < 0 {
		return c.offset, fmt.Errorf("out of bounds seek: %d which is < 0", c.offset)
	}

	c.offset = offset
	return c.offset, nil
}
