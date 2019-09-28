package kvm

// IOCTL calls
//const int IoctlGetMPState = KVM_GET_MP_STATE;
//const int IoctlSetMPState = KVM_SET_MP_STATE;
// States
//const int MPStateRunnable = KVM_MP_STATE_RUNNABLE;
//const int MPStateUninitialized = KVM_MP_STATE_UNINITIALIZED;
//const int MPStateInitReceived = KVM_MP_STATE_INIT_RECEIVED;
//const int MPStateHalted = KVM_MP_STATE_HALTED;
//const int MPStateSipiReceived = KVM_MP_STATE_SIPI_RECEIVED;

import (
	"encoding/json"
	"syscall"
	"unsafe"
)

// Our vcpus state.
type MPState int

var Runnable = MPState(Runnable)
var Uninitialized = MPState(Uninitialized)
var InitReceived = MPState(InitReceived)
var Halted = MPState(Halted)
var SIPIReceived = MPState(SIPIReceived)

var stateMap = map[MPState]string{
	MPStateRunnable:      "runnable",
	MPStateUninitialized: "uninitialized",
	MPStateInitReceived:  "init-received",
	MPStateHalted:        "halted",
	MPStateSipiReceived:  "sipi-received",
}

var stateRevMap = map[string]MPState{
	"runnable":      Runnable,
	"uninitialized": Uninitialized,
	"init-received": InitReceived,
	"halted":        Halted,
	"sipi-received": SIPIReceived,
}

func (self *VCPU) MPState() (MPState, error) {
	// Execute the ioctl.
	var kvmState struct_kvm_mp_state
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(self.Fd), uintptr(IOctlGetMPState), uintptr(unsafe.Pointer(&kvmState)))
	if e != 0 {
		return MPState(kvmState.mpState), e
	}
	return MPState(kvmState.mpState), nil
}

func (self *VCPU) SetMPState(state MPState) error {
	// Execute the ioctl.
	var kvmState struct_kvm_mp_state
	kvmState.mp_state = uint32(state)
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(self.Fd), uintptr(IOctlSetMPState), uintptr(unsafe.Pointer(&kvmState)))
	if e != 0 {
		return e
	}
	return nil
}

func (self *MPState) MarshalJSON() ([]byte, error) {
	// Marshal as a string.
	value, ok := stateMap[*self]
	if !ok {
		return nil, UnknownStateErr
	}
	return json.Marshal(value)
}

func (self *MPState) UnmarshalJSON(data []byte) error {
	// Unmarshal as an string.
	var value string
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	// Find the state.
	newState, ok := stateRevMap[value]
	if !ok {
		return UnknownStateErr
	}
	// That's our state.
	*state = newState
	return nil
}
