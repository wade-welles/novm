package machine

import (
	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

type PioEvent struct {
	*kvm.ExitPIO
}

func (pio PioEvent) Size() uint         { return pio.ExitPio.Size() }
func (pio PioEvent) GetData() uint64    { return *pio.ExitPio.Data() }
func (pio PioEvent) SetData(val uint64) { *pio.ExitPio.Data() = val }
func (pio PioEvent) IsWrite() bool      { return pio.ExitPio.IsOut() }

type PioDevice struct {
	BaseDevice
	// A map of our available ports.
	IOMap      `json:"-"`
	IOHandlers `json:"-"`
	// Our address in memory.
	Offset kvm.Pointer `json:"base"`
}

func (pio *PioDevice) PioHandlers() IOHandlers { return pio.IOHandlers }

func (pio *PioDevice) Attach(vm *kvm.VirtualMachine, model *Model) error {
	// Build our IO Handlers.
	pio.IOHandlers = make(IOHandlers)
	for region, ops := range pio.IOMap {
		new_region := MemoryRegion{region.Start + pio.Offset, region.Size}
		pio.IOHandlers[new_region] = NewIOHandler(pio, new_region.Start, ops)
	}
	// NOTE: Unlike pio devices, we don't reserve
	// memory regions for our ports. Whichever device
	// gets there first wins.
	return pio.BaseDevice.Attach(vm, model)
}
