// +build linux
package kvm

// IOCTL calls.
//const int IoctlCreatePIT2 = KVM_CREATE_PIT2;
//const int IoctlGetPIT2 = KVM_GET_PIT2;
//const int IoctlSetPIT2 = KVM_SET_PIT2;

// Size of pit state.
//const int PitSize = sizeof(struct kvm_pit_state2);

import (
	"syscall"
	"unsafe"
)

//
// PITState --
// We represent the PITState as a blob.
// This representation should be relatively
// safe from a forward-compatibility perspective,
// as KVM internally will take care of reserving
// bits and ensuring compatibility, etc.
type PITState struct {
	Data []byte `json:"data"`
}

func (vm *VirtualMachine) CreatePIT() error {
	// Prepare the PIT config.
	// The only flag supported at the time of writing
	// was KVM_PIT_SPEAKER_DUMMY, which I really have no
	// interest in supporting.
	var pit C.struct_kvm_pit_config
	pit.flags = C.__u32(0)

	// Execute the ioctl.
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vm.Fd),
		uintptr(C.IoctlCreatePIT2),
		uintptr(unsafe.Pointer(&pit)))
	if e != 0 {
		return e
	}

	return nil
}

func (vm *VirtualMachine) GetPIT() (PITState, error) {
	// Prepare the pit state.
	state := PITState{make([]byte, C.PITSize, C.PITSize)}
	// Execute the ioctl.
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vm.Fd),
		uintptr(C.IoctlGetPIT2),
		uintptr(unsafe.Pointer(&state.Data[0])))
	if e != 0 {
		return state, e
	}

	return state, nil
}

func (vm *VirtualMachine) SetPIT(state PITState) error {
	// Is there any state to set?
	// We just eat this error, it's fine.
	if state.Data == nil {
		return nil
	}
	// Is this the right size?
	if len(state.Data) != int(C.PITSize) {
		return PITIncompatibleErr
	}

	// Execute the ioctl.
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vm.Fd),
		uintptr(C.IoctlSetPIT2),
		uintptr(unsafe.Pointer(&state.Data[0])))
	if e != 0 {
		return e
	}

	return nil
}
