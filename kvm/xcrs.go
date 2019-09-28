package kvm

// IOCTL calls.
//const int IoctlGetXcrs = KVM_GET_XCRS;
//const int IoctlSetXcrs = KVM_SET_XCRS;

import (
	"syscall"
	"unsafe"
)

// A single XCR.
type XCR struct {
	ID    uint32 `json:"xcr"`
	Value uint64 `json:"value"`
}

func (vcpu *VCPU) GetXCRs() ([]XCR, error) {

	// Execute the ioctl.
	var kvm_xcrs C.struct_kvm_xcrs
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vcpu.Fd),
		uintptr(C.IoctlGetXCRs),
		uintptr(unsafe.Pointer(&kvm_xcrs)))
	if e != 0 {
		return nil, e
	}

	// Build our list.
	xcrs := make([]XCR, 0, kvm_xcrs.nr_xcrs)
	for i := 0; i < int(kvm_xcrs.nr_xcrs); i += 1 {
		xcrs = append(xcrs, XCR{
			Id:    uint32(kvm_xcrs.xcrs[i].xcr),
			Value: uint64(kvm_xcrs.xcrs[i].value),
		})
	}

	return xcrs, nil
}

func (vcpu *VCPU) SetXCRs(xcrs []XCR) error {

	// Build our parameter.
	var kvm_xcrs C.struct_kvm_xcrs
	kvm_xcrs.nr_xcrs = C.__u32(len(xcrs))
	for i, xcr := range xcrs {
		kvm_xcrs.xcrs[i].xcr = C.__u32(xcr.Id)
		kvm_xcrs.xcrs[i].value = C.__u64(xcr.Value)
	}

	// Execute the ioctl.
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vcpu.fd),
		uintptr(C.IoctlSetXcrs),
		uintptr(unsafe.Pointer(&kvm_xcrs)))
	if e != 0 {
		return e
	}

	return nil
}
