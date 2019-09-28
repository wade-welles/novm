package linux

import (
	"unsafe"

	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

func SetupLinuxBootParams(model *Model, boot_params_data, orig_boot_params_data []byte, cmdline_addr platform.Paddr, initrd_addr platform.Paddr, initrd_len uint64) error {
	// Grab a reference to our boot params struct.
	boot_params := (*C.struct_boot_params)(unsafe.Pointer(&boot_params_data[0]))
	// The setup header.
	// First step is to copy the existing setup_header
	// out of the given kernel image. We copy only the
	// header, and not the rest of the setup page.
	setup_start := 0x01f1
	setup_end := 0x0202 + int(orig_boot_params_data[0x0201])
	if setup_end > platform.PageSize {
		return InvalidSetupHeader
	}
	C.memcpy(unsafe.Pointer(&boot_params_data[setup_start]), unsafe.Pointer(&orig_boot_params_data[setup_start]), C.size_t(setup_end-setup_start))
	// Setup our BIOS memory map.
	// NOTE: We have to do this via C bindings. This is really
	// annoying, but basically because of the unaligned structures
	// in the struct_boot_params, the Go code generated here is
	// actually *incompatible* with the actual C layout.
	// First, the count.
	//C.e820_set_count(boot_params, C.int(len(model.MemoryMap)))
	//// Then, fill out the region information.
	//for index, region := range model.MemoryMap {
	//	var memtype C.int
	//	switch region.MemoryType {
	//	case MemoryTypeUser:
	//		memtype = C.E820Ram
	//	case MemoryTypeReserved:
	//		memtype = C.E820Reserved
	//	case MemoryTypeSpecial:
	//		memtype = C.E820Reserved
	//	case MemoryTypeAcpi:
	//		memtype = C.E820Acpi
	//	}
	//	//C.e820_set_region(boot_params, C.int(index), C.__u64(region.Start), C.__u64(region.Size), C.__u8(memtype))
	//}
	// Set necessary setup header bits.
	C.set_header(boot_params, C.__u64(initrd_addr), C.__u64(initrd_len), C.__u64(cmdline_addr))
	// All done!
	return nil
}
