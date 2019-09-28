package machine

import (
	"errors"
)

func DriverUnknown(name string) error {
	return errors.New(fmt.Sprintf("Unknown driver: %s", name))
}

var (
	ExitWithoutReasonErr = errors.New("Exit without reason?")
	NovCPUsErr           = errors.New("No vcpus?")
	// Basic Errors
	DeviceAlreadyPausedErr = errors.New("Device already paused!")
	DeviceNotPausedErr     = errors.New("Device not paused!")
	// Memory Allocation / Layout Errors
	MemoryConflictErr     = errors.New("Memory regions conflict!")
	MemoryNotFoundErr     = errors.New("Memory region not found!")
	MemoryBusyErr         = errors.New("Memory could not be allocated!")
	MemoryUnalignedErr    = errors.New("Memory not aligned!")
	UserMemoryNotFoundErr = errors.New("No user memory found?")
	// Interrupt allocation Errors
	InterruptConflictErr    = errors.New("Device interrupt conflict!")
	InterruptUnavailableErr = errors.New("No interrupt available!")
	// PCI Errors
	PciInvalidAddressErr     = errors.New("Invalid PCI address!")
	PciBusNotFoundErr        = errors.New("Requested PCI devices, but no bus found?")
	PciMSIErrorErr           = errors.New("MSI internal error?")
	PciCapabilityMismatchErr = errors.New("Capability mismatch!")
	// UART errors.
	UartUnknownErr = errors.New("Unknown COM port.")
	// Virtio errors.
	VirtioInvalidQueueSizeErr      = errors.New("Invalid VirtIO queue size!")
	VirtioUnsupportedVnetHeaderErr = errors.New("Unsupported vnet header size.")
	// I/O memoize errors.
	// This is an internal-only error which is returned from
	// a write handler. When this is returned (and the cache
	// has had a significant number of hits at that address)
	// we will create an eventfd for that particular address
	// and value. This will reduce the number of kernel-user
	// switches necessary to handle that particular address.
	SaveIOErr = errors.New("Save I/O request (internal error).")
)
