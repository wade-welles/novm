package kvm

// IOCTL calls.
//const int IoctlGetXSave = KVM_GET_XSAVE;
//const int IoctlSetXSave = KVM_SET_XSAVE;

import (
	"syscall"
	"unsafe"
)

// Our xsave state.
type XSave struct {
	Region [1024]uint32 `json:"region"`
}

func (vcpu *VCPU) GetXSave() (XSave, error) {

	// Execute the ioctl.
	var kvm_xsave C.struct_kvm_xsave
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vcpu.Fd),
		uintptr(C.IoctlGetXSave),
		uintptr(unsafe.Pointer(&kvm_xsave)))
	if e != 0 {
		return XSave{}, e
	}

	state := XSave{}
	for i := 0; i < len(state.Region); i += 1 {
		state.Region[i] = uint32(kvm_xsave.region[i])
	}

	return state, nil
}

func (vcpu *VCPU) SetXSave(state XSave) error {

	// Execute the ioctl.
	var kvm_xsave C.struct_kvm_xsave
	for i := 0; i < len(state.Region); i += 1 {
		kvm_xsave.region[i] = C.__u32(state.Region[i])
	}
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vcpu.Fd),
		uintptr(C.IoctlSetXSave),
		uintptr(unsafe.Pointer(&kvm_xsave)))
	if e != 0 {
		return e
	}

	return nil
}
