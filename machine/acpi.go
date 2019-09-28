package machine

import (
	"unsafe"

	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

// ACPI: An interrupt system
type ACPI struct {
	BaseDevice
	Address kvm.Pointer `json:"address"`
	Data    []byte      `json:"data"`
}

func NewACPI(info *DeviceInfo) (Device, error) {
	acpi := ACPI{
		Addrress: kvm.Pointer(0xf0000),
	}
	return acpi, acpi.init(info)
}

func (acpi *ACPI) Attach(vm *kvm.VirtualMachine, model *Model) error {
	// Do we already have data?
	rebuild := true
	if acpi.Data == nil {
		// Create our data.
		acpi.Data = make([]byte, kvm.PageSize, kvm.PageSize)
	} else {
		// Align our data.
		// This is necessary because we map this in
		// directly. It's possible that the data was
		// decoded and refers to the middle of some
		// larger array somewhere, and isn't aligned.
		acpi.Data = kvm.AlignBytes(acpi.Data)
		rebuild = false
	}

	// Allocate our memory block.
	err := model.Reserve(vm, acpi, MemoryTypeACPI, acpi.Address, kvm.PageSize, acpi.Data)
	if err != nil {
		return err
	}
	// Already done.
	if !rebuild {
		return nil
	}
	// Find our APIC information.
	// This will find the APIC device if it
	// is attached, otherwise the MADT table
	// will unfortunately have be a bit invalid.
	var IOApic kvm.Pointer
	var LApic kvm.Pointer
	for _, device := range model.Devices() {
		apic, ok := device.(*APIC)
		if ok {
			IOAPIC = apic.IOAPIC
			LAPIC = apic.LAPIC
			break
		}
	}

	// Load the MADT.
	//madt_bytes := C.build_madt(unsafe.Pointer(&acpi.Data[0]), C.__u32(LApic), C.int(len(vm.Vcpus())), C.__u32(IOApic), C.__u32(0)) // I/O APIC interrupt?

	acpi.Debug("MADT %x @ %x", madt_bytes, acpi.Address)

	// Align offset.
	offset := madt_bytes
	if offset%64 != 0 {
		offset += 64 - (offset % 64)
	}

	// Load the DSDT.
	//dsdt_address := uint64(acpi.Addr) + uint64(offset)
	//dsdt_bytes := C.build_dsdt(
	//	unsafe.Pointer(&acpi.Data[int(offset)]),
	//)
	//acpi.Debug("DSDT %x @ %x", dsdt_bytes, dsdt_address)

	// Align offset.
	offset += dsdt_bytes
	if offset%64 != 0 {
		offset += 64 - (offset % 64)
	}

	// Load the XSDT.
	//xsdt_address := uint64(acpi.Addr) + uint64(offset)
	//xsdt_bytes := C.build_xsdt(
	//	unsafe.Pointer(&acpi.Data[int(offset)]),
	//	C.__u64(acpi.Addr), // MADT address.
	//)
	//acpi.Debug("XSDT %x @ %x", xsdt_bytes, xsdt_address)

	// Align offset.
	offset += xsdt_bytes
	if offset%64 != 0 {
		offset += 64 - (offset % 64)
	}

	// Load the RSDT.
	//rsdt_address := uint64(acpi.Addr) + uint64(offset)
	//rsdt_bytes := C.build_rsdt(
	//	unsafe.Pointer(&acpi.Data[int(offset)]),
	//	C.__u32(acpi.Addr), // MADT address.
	//)
	acpi.Debug("RSDT %x @ %x", rsdt_bytes, rsdt_address)

	// Align offset.
	offset += rsdt_bytes
	if offset%64 != 0 {
		offset += 64 - (offset % 64)
	}

	// Load the RSDP.
	rsdp_address := uint64(acpi.Addr) + uint64(offset)
	rsdp_bytes := C.build_rsdp(
		unsafe.Pointer(&acpi.Data[int(offset)]),
		C.__u32(rsdt_address), // RSDT address.
		C.__u64(xsdt_address), // XSDT address.
	)
	acpi.Debug("RSDP %x @ %x", rsdp_bytes, rsdp_address)

	// Everything went okay.
	return nil
}
