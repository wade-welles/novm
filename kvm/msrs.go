package kvm

//const int IoctlGetMsrIndexList = KVM_GET_MSR_INDEX_LIST;
//const int IoctlSetMsrs = KVM_SET_MSRS;
//const int IoctlGetMsrs = KVM_GET_MSRS;

import (
	"syscall"
	"unsafe"
)

type MSR struct {
	Index uint32 `json:"index"`
	Value uint64 `json:"value"`
}

func availableMSRs(fd int) ([]uint32, error) {

	// Find our list of MSR indices.
	// A page should be more than enough here,
	// eventually if it's not we'll end up with
	// a failed system call for some reason other
	// than E2BIG (which just says n is wrong).
	msrIndices := make([]byte, PageSize, PageSize)
	msrs := make([]uint32, 0, 0)

	for {
		_, _, e := syscall.Syscall(
			syscall.SYS_IOCTL,
			uintptr(fd),
			uintptr(C.IoctlGetMsrIndexList),
			uintptr(unsafe.Pointer(&msrIndices[0])))
		if e == syscall.E2BIG {
			// The nmsrs field will now have been
			// adjusted, and we can run it again.
			continue
		} else if e != 0 {
			return nil, e
		}

		// We're good!
		break
	}

	// Extract each msr individually.
	for i := 0; ; i += 1 {
		// Is there a valid index?
		var index C.__u32
		e := C.msr_list_index(
			unsafe.Pointer(&msrIndices[0]),
			C.int(i),
			&index)

		// Any left?
		if e != 0 {
			break
		}

		// Add this MSR.
		msrs = append(msrs, uint32(index))
	}

	return msrs, nil
}

func (vcpu *VCPU) GetMSR(index uint32) (uint64, error) {

	// Setup our structure.
	data := make([]byte, C.msr_size(), C.msr_size())

	// Set our index to retrieve.
	C.msrSet(unsafe.Pointer(&data[0]), C.__u32(index), C.__u64(0))

	// Execute our ioctl.
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vcpu.Dd),
		uintptr(C.IoctlGetMsrs),
		uintptr(unsafe.Pointer(&data[0])))
	if e != 0 {
		return 0, e
	}

	// Return our value.
	return uint64(C.msrGet(unsafe.Pointer(&data[0]))), nil
}

func (vcpu *VCPU) SetMSR(index uint32, value uint64) error {

	// Setup our structure.
	data := make([]byte, C.msrSize(), C.msrSize())

	// Set our index and value.
	C.msrSet(unsafe.Pointer(&data[0]), C.__u32(index), C.__u64(value))

	// Execute our ioctl.
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vcpu.Fd),
		uintptr(C.IoctlSetMSRs),
		uintptr(unsafe.Pointer(&data[0])))
	if e != 0 {
		return e
	}

	return nil
}

func (vcpu *VCPU) GetMSRs() ([]MSR, error) {

	// Extract each msr individually.
	msrs := make([]MSR, 0, len(vcpu.msrs))

	for _, index := range vcpu.msrs {
		// Get this MSR.
		value, err := vcpu.GetMSR(index)
		if err != nil {
			return msrs, err
		}

		msrs = append(msrs, MSR{uint32(index), uint64(value)})
	}
	// Finish it off.
	return msrs, nil
}

func (vcpu *VCPU) SetMSRs(msrs []MSR) error {
	for _, msr := range msrs {
		// Set our msrs.
		err := vcpu.SetMSR(msr.Index, msr.Value)
		if err != nil {
			return err
		}
	}
	return nil
}
