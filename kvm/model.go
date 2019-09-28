package kvm

type MemoryType int

const (
	MemoryReserved MemoryType = iota
	MemoryUser
	MemoryACPI
	MemorySpecial
)

type MemoryRegion struct {
	Start Pointer `json:"start"`
	Size  uint64  `json:"size"`
}

type TypedMemoryRegion struct {
	MemoryRegion
	MemoryType
	// The memory pointer (slice).
	User []byte
	// Allocated chunks.
	// These are offsets, which point
	// to the amount of memory allocated.
	Allocated map[uint64]uint64
}

//
// MemoryMap --
//
// Our collection of current memory regions.
//
type MemoryMap []*TypedMemoryRegion

//
// Model --
// Our basic machine model.
// This is very much different from a standard virtual machine.
// First, we only support a very limited selection of devices.
// We actually do not support *any* I/O-port based devices, which
// includes PCI devices (which require an I/O port at the root).
type Model struct {
	// Basic memory layout:
	// This is generally accessible from the loader,
	// and other modules that may need to tweak memory.
	MemoryMap
	// Basic interrupt layout:
	// This maps interrupts to devices.
	InterruptMap
	// All devices.
	Devices []Device
	// Our device lookup cache.
	PIOCache  *IOCache
	MMIOCache *IOCache
}

func NewModel(vm *VirtualMachine) (*Model, error) {
	// Create our model object.
	model := new(Model)
	// Setup the memory map.
	model.MemoryMap = make(MemoryMap, 0, 0)
	// Setup the interrupt map.
	model.InterruptMap = make(InterruptMap)
	// Create our devices.
	model.devices = make([]Device, 0, 0)
	// We're set.
	return model, nil
}

func (self *Model) Flush() error {
	collectIOHandlers := func(isPIO bool) []IOHandlers {
		ioHandlers := make([]IOHandlers, 0, 0)
		for _, device := range self.devices {
			if isPIO {
				ioHandlers = append(ioHandlers, device.PIOHandlers())
			} else {
				ioHandlers = append(ioHandlers, device.MMIOHandlers())
			}
		}
		return ioHandlers
	}
	// (Re-)Create our IOCache.
	self.PIOCache = NewIOCache(collectIOHandlers(true), true)
	self.MMIOCache = NewIOCache(collectIOHandlers(false), false)
	// We're okay.
	return nil
}

func (self *Model) Pause(manual bool) error {
	for i, device := range self.devices {
		// Ensure all devices are paused.
		err := device.Pause(manual)
		if err != nil && err != DeviceAlreadyPaused {
			for i -= 1; i >= 0; i -= 1 {
				device.Unpause(manual)
			}
			return err
		}
	}
	// All good.
	return nil
}

func (self *Model) Unpause(manual bool) error {
	for i, device := range self.devices {
		// Ensure all devices are unpaused.
		err := device.Unpause(manual)
		if err != nil && err != DeviceAlreadyPaused {
			for i -= 1; i >= 0; i -= 1 {
				device.Pause(manual)
			}
			return err
		}
	}
	// All good.
	return nil
}

func (self *Model) Load(vm *VirtualMachine) error {
	for _, device := range self.devices {
		// Load our device state.
		err := device.Load(vm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Model) Save(vm *VirtualMachine) error {
	for _, device := range self.devices {
		// Synchronize our device state.
		err := device.Save(vm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Model) DeviceInfo(vm *VirtualMachine) ([]DeviceInfo, error) {
	err := self.Pause(false)
	if err != nil {
		return nil, err
	}
	defer self.Unpause(false)
	// Synchronize our state.
	err = self.Save(vm)
	if err != nil {
		return nil, err
	}

	devices := make([]DeviceInfo, 0, len(self.devices))
	for _, device := range self.devices {
		// Get the deviceinfo.
		deviceinfo, err := NewDeviceInfo(device)
		if err != nil {
			return nil, err
		}
		devices = append(devices, deviceinfo)
	}
	return devices, nil
}
