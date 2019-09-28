package protocol

import (
	"errors"
)

var (
	UnknownStatusErr  = errors.New("Unknown status?")
	UnknownCommandErr = errors.New("Unknown command?")
)
