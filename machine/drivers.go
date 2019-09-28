package machine

// A driver load function.
type Driver func(info *DeviceInfo) (Device, error)

// All available device drivers.
var drivers = map[string]Driver{
	"bios":                NewBIOS,
	"apic":                NewAPIC,
	"pit":                 NewPIT,
	"acpi":                NewACPI,
	"rtc":                 NewRTC,
	"clock":               NewClock,
	"uart":                NewUART,
	"pci-bus":             NewPCIBus,
	"pci-hostbridge":      NewPCIHostBridge,
	"user-memory":         NewUserMemory,
	"virtio-pci-block":    NewVirtioPCIBlock,
	"virtio-mmio-block":   NewVirtioMMIOBlock,
	"virtio-pci-console":  NewVirtioPCIConsole,
	"virtio-mmio-console": NewVirtioMMIOConsole,
	"virtio-pci-net":      NewVirtioPCINet,
	"virtio-mmio-net":     NewVirtioMMIONet,
	"virtio-pci-fs":       NewVirtioPCIFs,
	"virtio-mmio-fs":      NewVirtioMMIOFs,
}
