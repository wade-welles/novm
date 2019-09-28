package kvm

import (
	"io"
)

type Proxy interface {
	io.ReadWriteCloser
}
