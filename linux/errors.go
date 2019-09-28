package linux

import (
	"errors"
)

var InvalidSetupHeader = errors.New("Setup header past page boundary?")
