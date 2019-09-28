package plan9

import (
	"errors"
)

// Global errors.
var (
	BufferInsufficient = errors.New("insufficient buffer?")
	InvalidMessage     = errors.New("invalid 9pfs message?")
	XattrError         = errors.New("unable to fetch xattr?")

	// Internal errors.
	Eunknownfid error = &Error{"unknown fid", EINVAL}
	Enoauth     error = &Error{"no authentication required", EINVAL}
	Einuse      error = &Error{"fid already in use", EINVAL}
	Ebaduse     error = &Error{"bad use of fid", EINVAL}
	Eopen       error = &Error{"fid already opened", EINVAL}
	Enotdir     error = &Error{"not a directory", ENOTDIR}
	Eperm       error = &Error{"permission denied", EPERM}
	Etoolarge   error = &Error{"i/o count too large", EINVAL}
	Ebadoffset  error = &Error{"bad offset in directory read", EINVAL}
	Edirchange  error = &Error{"cannot convert between files and directories", EINVAL}
	Enouser     error = &Error{"unknown user", EINVAL}
	Enotimpl    error = &Error{"not implemented", EINVAL}
	Eexist      error = &Error{"file already exists", EEXIST}
	Enoent      error = &Error{"file not found", ENOENT}
	Enotempty   error = &Error{"directory not empty", EPERM}
)
