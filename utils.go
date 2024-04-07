package pustynia

import (
	"errors"
	"io"
	"net"
)

type empty struct{}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func isClosed(err error) bool {
	return err == io.EOF || errors.Is(err, net.ErrClosed)
}
