package kvm

import (
	"log"
	"sync"
)

type BaseDevice struct {
	// Pointer to original device info.
	// This is reference in serialization.
	// (But is explicitly not exported, as
	// the device info will have a reference
	// back to this new device object).
	Info *DeviceInfo
	// Have we been paused manually?
	IsPaused bool
	// Internal pause count.
	Paused int
	// Our internal lock for pause/resume.
	// This is significantly simpler than the
	// VCPU case, so we can get away with using
	// just a straight-forward RWMUtex.
	pauseLock sync.Mutex
	runLock   sync.RWMutex
}

type Device interface {
	Name() string
	Driver() string

	PIOHandlers() IOHandlers
	MMIOHandlers() IOHandlers

	Attach(vm *VirtualMachine, model *Model) error
	Load(vm *VirtualMachine) error
	Save(vm *VirtualMachine) error

	Pause(manual bool) error
	Unpause(manual bool) error

	Acquire()
	Release()

	Interrupt() error

	Debug(format string, v ...interface{})
	IsDebugging() bool
	SetDebugging(debug bool)
}

func (self *BaseDevice) Init(info *DeviceInfo) error {
	// Save our original device info.
	// This isn't structural (hence no export).
	self.Info = info
	return nil
}

func (device *BaseDevice) Name() string {
	return device.info.Name
}

func (self *BaseDevice) Driver() string {
	return self.Info.Driver
}

func (self *BaseDevice) PIOHandlers() IOHandlers {
	return IOHandlers{}
}

func (self *BaseDevice) MMIOHandlers() IOHandlers {
	return IOHandlers{}
}

func (self *BaseDevice) Attach(vm *VirtualMachine, model *Model) error {
	return nil
}

func (self *BaseDevice) Load(vm *VirtualMachine) error {
	return nil
}

func (self *BaseDevice) Save(vm *VirtualMachine) error {
	return nil
}

func (self *BaseDevice) Pause(manual bool) error {
	self.pauseLock.Lock()
	defer self.pauseLock.Unlock()

	if manual {
		if self.IsPaused {
			return DeviceAlreadyPaused
		}
		self.IsPaused = true
		if self.Paused > 0 {
			// Already paused.
			return nil
		}
	} else {
		self.Paused += 1
		if self.Paused > 1 || self.IsPaused {
			// Already paused.
			self.Paused += 1
			return nil
		}
	}
	// Acquire our runlock, preventing
	// any execution from continuing.
	self.runLock.Lock()
	return nil
}

func (self *BaseDevice) Unpause(manual bool) error {
	self.pauseLock.Lock()
	defer self.pauseLock.Unlock()

	if manual {
		if !self.IsPaused {
			return DeviceNotPaused
		}
		self.IsPaused = false
		if self.Paused > 0 {
			// Please don't unpause.
			return nil
		}
	} else {
		self.Paused -= 1
		if self.Paused > 0 || self.IsPaused {
			// Please don't unpause.
			return nil
		}
	}
	// Release our runlock, allow
	// execution to continue normally.
	self.runLock.Unlock()
	return nil
}

func (self *BaseDevice) Acquire() {
	self.runLock.RLock()
}

func (self *BaseDevice) Release() {
	self.runLock.RUnlock()
}

func (self *BaseDevice) Interrupt() error {
	return nil
}

func (self *BaseDevice) Debug(format string, v ...interface{}) {
	if self.IsDebugging() {
		log.Printf(self.Name()+": "+format, v...)
	}
}

func (self *BaseDevice) IsDebugging() bool {
	return self.info.Debug
}

func (self *BaseDevice) SetDebugging(debug bool) {
	self.Info.Debug = debug
}
