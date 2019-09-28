package kvm

// IOCTL calls.
//const int IOctlGetVCPUEvents = KVM_GET_VCPU_EVENTS;
//const int IOctlSetVCPUEvents = KVM_SET_VCPU_EVENTS;

import (
	"syscall"
	"unsafe"
)

//
// Our event state.
type ExceptionEvent struct {
	Number    uint8   `json:"number"`
	ErrorCode *uint32 `json:"error-code"`
}

type InterruptEvent struct {
	Number uint8 `json:"number"`
	Soft   bool  `json:"soft"`
	Shadow bool  `json:"shadow"`
}

type Events struct {
	Exception *ExceptionEvent `json:"exception"`
	Interrupt *InterruptEvent `json:"interrupt"`

	NMIPending bool `json:"nmi-pending"`
	NMIMasked  bool `json:"nmi-masked"`

	SIPIVector uint32 `json:"sipi-vector"`
	Flags      uint32 `json:"flags"`
}

func (vcpu *VCPU) GetEvents() (Events, error) {

	// Execute the ioctl.
	var kvmEvents C.struct_kvm_vcpu_events
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vcpu.Fd), uintptr(C.IOctlGetVCPUEvents), uintptr(unsafe.Pointer(&kvmEvents)))
	if e != 0 {
		return Events{}, e
	}

	// Prepare our state.
	events := Events{
		NMIPending: kvmEvents.NMI.pending != C.__u8(0),
		NMIMasked:  kvmEvents.NMI.masked != C.__u8(0),
		SIPIVector: uint32(kvmEvents.sipiVector),
		Flags:      uint32(kvmEvents.flags),
	}
	if kvmEvents.exception.injected != C.__u8(0) {
		events.Exception = &ExceptionEvent{
			Number: uint8(kvmEvents.exception.nr),
		}
		if kvmEvents.exception.has_error_code != C.__u8(0) {
			error_code := uint32(kvmEvents.exception.error_code)
			events.Exception.ErrorCode = &error_code
		}
	}
	if kvmEvents.interrupt.injected != C.__u8(0) {
		events.Interrupt = &InterruptEvent{
			Number: uint8(kvmEvents.interrupt.nr),
			Soft:   kvmEvents.interrupt.soft != C.__u8(0),
			Shadow: kvmEvents.interrupt.shadow != C.__u8(0),
		}
	}

	return events, nil
}

func (vcpu *VCPU) SetEvents(events Events) error {

	// Prepare our state.
	var kvmEvents C.struct_kvm_vcpu_events

	if events.NMIPending {
		kvmEvents.NMI.pending = C.__u8(1)
	} else {
		kvmEvents.NMI.pending = C.__u8(0)
	}
	if events.NMIMasked {
		kvmEvents.NMI.masked = C.__u8(1)
	} else {
		kvmEvents.NMI.masked = C.__u8(0)
	}

	kvmEvents.sipiVector = C.__u32(events.SIPIVector)
	kvmEvents.flags = C.__u32(events.Flags)

	if events.Exception != nil {
		kvmEvents.exception.injected = C.__u8(1)
		kvmEvents.exception.nr = C.__u8(events.Exception.Number)
		if events.Exception.ErrorCode != nil {
			kvmEvents.exception.has_error_code = C.__u8(1)
			kvmEvents.exception.error_code = C.__u32(*events.Exception.ErrorCode)
		} else {
			kvmEvents.exception.has_error_code = C.__u8(0)
			kvmEvents.exception.error_code = C.__u32(0)
		}
	} else {
		kvmEvents.exception.injected = C.__u8(0)
		kvmEvents.exception.nr = C.__u8(0)
		kvmEvents.exception.has_error_code = C.__u8(0)
		kvmEvents.exception.error_code = C.__u32(0)
	}
	if events.Interrupt != nil {
		kvmEvents.interrupt.injected = C.__u8(1)
		kvmEvents.interrupt.nr = C.__u8(events.Interrupt.Number)
		if events.Interrupt.Soft {
			kvmEvents.interrupt.soft = C.__u8(1)
		} else {
			kvmEvents.interrupt.soft = C.__u8(0)
		}
		if events.Interrupt.Shadow {
			kvmEvents.interrupt.shadow = C.__u8(1)
		} else {
			kvmEvents.interrupt.shadow = C.__u8(0)
		}
	} else {
		kvmEvents.interrupt.injected = C.__u8(0)
		kvmEvents.interrupt.nr = C.__u8(0)
		kvmEvents.interrupt.soft = C.__u8(0)
		kvmEvents.interrupt.shadow = C.__u8(0)
	}
	// Execute the ioctl.
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vcpu.fd), uintptr(C.IOctlSetVCPUEvents), uintptr(unsafe.Pointer(&kvmEvents)))
	if e != 0 {
		return e
	}

	return nil
}
