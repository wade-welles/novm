package kvm

import (
	"errors"
	"syscall"

	unix "golang.org/x/sys/unix"
)

var (
	// Serialization Errors
	incompatibleVCPUDataErr = errors.New("Incompatible VCPU data?")
	incompatiblePITErr      = errors.New("Incompatible PIT state?")
	incompatibleIRQChipErr  = errors.New("Incompatible IRQ chip state?")
	incompatibleLApicErr    = errors.New("Incompatible LApic state?")
	// vCPU State Errors
	vCPUNotPausedErr     = errors.New("Vcpu is not paused?")
	vCPUAlreadyPausedErr = errors.New("Vcpu is already paused.")
	vcPUUnknownStateErr  = errors.New("Unknown vcpu state?")
	// Register Errors
	registerUnknownErr = errors.New("Unknown Register")
)

// https://github.com/golang/sys/blob/master/unix/syscall_unix.go#L33
func syscallError(e syscall.Errno) error {
	switch e {
	case unix.EAGAIN:
		return syscall.EAGAIN
	case unix.EINVAL:
		return syscall.EINVAL
	case unix.ENOENT:
		return syscall.ENOENT
	case unix.EINTR:
		return syscall.EINTR
	default:
		return errors.New("[syscall] failed to execute syscall, error code unknown:", e)
	}
}
