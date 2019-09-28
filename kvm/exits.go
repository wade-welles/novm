// +build linux
package kvm

import (
	"unsafe"
)

//export KVMExitMMIO
func KVMExitMMIO(addr uint64, data uint64, length uint32, write int) unsafe.Pointer {
	return unsafe.Pointer(&ExitMMIO{
		addr:   Pointer(addr),
		data:   (*uint64)(data),
		length: uint32(length),
		write:  write != int(0),
	})
}

//export KVMExitPio
func KVMExitPointer(port uint16, size uint8, data unsafe.Pointer, length uint32, out int) unsafe.Pointer {
	return unsafe.Pointer(&ExitPio{
		port: Paddr(port),
		size: uint8(size),
		data: (*uint64)(data),
		out:  out != int(0),
	})
}

//export KVMExitInternalError
func KVMExitInternalError(code uint32) unsafe.Pointer {
	return unsafe.Pointer(&ExitInternalError{
		code: uint32(code),
	})
}

//export KVMExitException
func KVMExitException(exception uint32, error_code uint32) unsafe.Pointer {
	return unsafe.Pointer(&ExitException{
		exception: uint32(exception),
		errorCode: uint32(error_code),
	})
}

//export KVMExitUnknown
func KVMExitUnknown(code uint32) unsafe.Pointer {
	return unsafe.Pointer(&ExitUnknown{
		code: uint32(code),
	})
}

func (vcpu *VCPU) GetExitError() error {
	// Handle the error.
	switch int(vcpu.KVM.ExitReason) {
	case ExitReasonMMIO:
		return (*ExitMMIO)(handle_exit_mmio(vcpu.KVM))
	case ExitReasonIO:
		return (*ExitPio)(handle_exit_io(vcpu.KVM))
	case ExitReasonInternalError:
		return (*ExitInternalError)(handle_exit_internal_error(vcpu.KVM))
	case ExitReasonException:
		return (*ExitException)(handle_exit_exception(vcpu.KVM))
	case ExitReasonDebug:
		return &ExitDebug{}
	case ExitReasonShutdown:
		return &ExitShutdown{}
	default:
		return (*ExitUnknown)(handle_exit_unknown(vcpu.KVM))
	}

	// Unreachable.
	return nil
}
