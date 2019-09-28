package machine

import (
	platform "github.com/multiverse-os/portalgun/vm/kvm"
)

type Clock struct {
	BaseDevice
	// Our clock state.
	Clock platform.Clock `json:"clock"`
}

func NewClock(info *DeviceInfo) (Device, error) {
	clock := new(Clock)
	return clock, clock.init(info)
}

func (clock *Clock) Attach(vm *platform.Vm, model *Model) error { return nil }

func (clock *Clock) Save(vm *platform.Vm) error {
	var err error
	// Save our clock state.
	clock.Clock, err = vm.GetClock()
	if err != nil {
		return err
	}
	// We're good.
	return nil
}

func (clock *Clock) Load(vm *platform.Vm) error { return vm.SetClock(clock.Clock) }
