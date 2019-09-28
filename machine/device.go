package machine

import (
	"log"
	"sync"
)

type IoMap map[MemoryRegion]IoOperations
type IoHandlers map[MemoryRegion]*IoHandler

type BaseDevice struct {
	// Pointer to original device info.
	// This is reference in serialization.
	// (But is explicitly not exported, as
	// the device info will have a reference
	// back to this new device object).
	info *DeviceInfo
	// Have we been paused manually?
	is_paused bool
	// Internal pause count.
	paused int
	// Our internal lock for pause/resume.
	// This is significantly simpler than the
	// VCPU case, so we can get away with using
	// just a straight-forward RWMUtex.
	pause_lock sync.Mutex
	run_lock   sync.RWMutex
}

type Device interface {
	Name() string
	Driver() string

	PioHandlers() IoHandlers
	MmioHandlers() IoHandlers

	Attach(vm *kvm.VirtualMachine, model *Model) error
	Load(vm *kvm.VirtualMachine) error
	Save(vm *kvm.VirtualMachine) error

	Pause(manual bool) error
	Unpause(manual bool) error

	Acquire()
	Release()

	Interrupt() error

	Debug(format string, v ...interface{})
	IsDebugging() bool
	SetDebugging(debug bool)
}

func (device *BaseDevice) init(info *DeviceInfo) error {
	// Save our original device info.
	// This isn't structural (hence no export).
	device.info = info
	return nil
}

func (device *BaseDevice) Name() string {
	return device.info.Name
}

func (device *BaseDevice) Driver() string {
	return device.info.Driver
}

func (device *BaseDevice) PioHandlers() IoHandlers {
	return IoHandlers{}
}

func (device *BaseDevice) MmioHandlers() IoHandlers {
	return IoHandlers{}
}

func (device *BaseDevice) Attach(vm *kvm.VirtualMachine, model *Model) error {
	return nil
}

func (device *BaseDevice) Load(vm *kvm.VirtualMachine) error {
	return nil
}

func (device *BaseDevice) Save(vm *kvm.VirtualMachine) error {
	return nil
}

func (device *BaseDevice) Pause(manual bool) error {
	device.pause_lock.Lock()
	defer device.pause_lock.Unlock()

	if manual {
		if device.isPaused {
			return DeviceAlreadyPaused
		}
		device.isPaused = true
		if device.paused > 0 {
			// Already paused.
			return nil
		}
	} else {
		device.paused += 1
		if device.paused > 1 || device.isPaused {
			// Already paused.
			device.paused += 1
			return nil
		}
	}

	// Acquire our runlock, preventing
	// any execution from continuing.
	device.run_lock.Lock()
	return nil
}

func (device *BaseDevice) Unpause(manual bool) error {
	device.pause_lock.Lock()
	defer device.pause_lock.Unlock()

	if manual {
		if !device.is_paused {
			return DeviceNotPaused
		}
		device.is_paused = false
		if device.paused > 0 {
			// Please don't unpause.
			return nil
		}
	} else {
		device.paused -= 1
		if device.paused > 0 || device.isPaused {
			// Please don't unpause.
			return nil
		}
	}

	// Release our runlock, allow
	// execution to continue normally.
	device.run_lock.Unlock()
	return nil
}

func (device *BaseDevice) Acquire() {
	device.run_lock.RLock()
}

func (device *BaseDevice) Release() {
	device.run_lock.RUnlock()
}

func (device *BaseDevice) Interrupt() error {
	return nil
}

func (device *BaseDevice) Debug(format string, v ...interface{}) {
	if device.IsDebugging() {
		log.Printf(device.Name()+": "+format, v...)
	}
}

func (device *BaseDevice) IsDebugging() bool {
	return device.info.Debug
}

func (device *BaseDevice) SetDebugging(debug bool) {
	device.info.Debug = debug
}
