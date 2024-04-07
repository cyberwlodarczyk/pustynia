package pustynia

import (
	"errors"
	"io"
	"net"
)

type empty struct{}

func isClosed(err error) bool {
	return err == io.EOF || errors.Is(err, net.ErrClosed)
}
