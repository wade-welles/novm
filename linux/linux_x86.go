package linux

import (
	"log"
	"unsafe"

	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

var Linux32Convention = Convention{
	instruction: kvm.RIP,
	arguments: []kvm.Register{
		kvm.RCX,
		kvm.RDX,
	},
	rvalue: kvm.RAX,
	stack:  kvm.RSI,
}

var Linux64Convention = Convention{
	instruction: kvm.RIP,
	arguments: []kvm.Register{
		kvm.RDI,
		kvm.RSI,
		kvm.RDX,
		kvm.RCX,
	},
	rvalue: kvm.RAX,
	stack:  kvm.RBP,
}

func SetupLinux(vcpu *kvm.Vcpu, model *Model, orig_boot_data []byte, entry_point uint64, is_64bit bool, initrd_addr kvm.Paddr, initrd_len uint64, cmdline_addr kvm.Paddr) error {
	// Copy in the GDT table.
	// These match the segments below.
	gdt_addr, gdt, err := model.Allocate(MemoryTypeUser, 0, model.Max(), kvm.PageSize, false)
	if err != nil {
		return err
	}
	if is_64bit {
		C.build_64bit_gdt(unsafe.Pointer(&gdt[0]))
	} else {
		C.build_32bit_gdt(unsafe.Pointer(&gdt[0]))
	}
	BootGdt := kvm.DescriptorValue{
		Base:  uint64(gdt_addr),
		Limit: uint16(kvm.PageSize),
	}
	err = vcpu.SetDescriptor(kvm.GDT, BootGdt, true)
	if err != nil {
		return err
	}
	// Set a null IDT.
	BootIdt := kvm.DescriptorValue{
		Base:  0,
		Limit: 0,
	}
	err = vcpu.SetDescriptor(kvm.IDT, BootIdt, true)
	if err != nil {
		return err
	}
	// Enable protected-mode.
	// This does not set any flags (e.g. paging) beyond the
	// protected mode flag. This is according to Linux entry
	// protocol for 32-bit protected mode.
	cr0, err := vcpu.GetControlRegister(kvm.CR0)
	if err != nil {
		return err
	}
	cr0 = cr0 | (1 << 0) // Protected mode.
	err = vcpu.SetControlRegister(kvm.CR0, cr0, true)
	if err != nil {
		return err
	}
	// Always have the PAE bit set.
	cr4, err := vcpu.GetControlRegister(kvm.CR4)
	if err != nil {
		return err
	}
	cr4 = cr4 | (1 << 5) // PAE enabled.
	err = vcpu.SetControlRegister(kvm.CR4, cr4, true)
	if err != nil {
		return err
	}
	// For 64-bit kernels, we need to enable long mode,
	// and load an identity page table. This will require
	// only a page of pages, as we use huge page sizes.
	if is_64bit {
		// Create our page tables.
		pde_addr, pde, err := model.Allocate(MemoryTypeUser, 0, model.Max(), kvm.PageSize, false)
		if err != nil {
			return err
		}
		pgd_addr, pgd, err := model.Allocate(MemoryTypeUser, 0, model.Max(), kvm.PageSize, false)
		if err != nil {
			return err
		}
		pml4_addr, pml4, err := model.Allocate(MemoryTypeUser, 0, model.Max(), kvm.PageSize, false)
		if err != nil {
			return err
		}
		C.build_pde(unsafe.Pointer(&pde[0]), kvm.PageSize)
		C.build_pgd(unsafe.Pointer(&pgd[0]), C.__u64(pde_addr), kvm.PageSize)
		C.build_pml4(unsafe.Pointer(&pml4[0]), C.__u64(pgd_addr), kvm.PageSize)

		log.Printf("loader: Created PDE @ %08x.", pde_addr)
		log.Printf("loader: Created PGD @ %08x.", pgd_addr)
		log.Printf("loader: Created PML4 @ %08x.", pml4_addr)

		// Set our newly build page table.
		err = vcpu.SetControlRegister(kvm.CR3, kvm.ControlRegisterValue(pml4_addr), true)
		if err != nil {
			return err
		}
		// Enable long mode.
		efer, err := vcpu.GetControlRegister(kvm.EFER)
		if err != nil {
			return err
		}
		efer = efer | (1 << 8) // Long-mode enable.
		err = vcpu.SetControlRegister(kvm.EFER, efer, true)
		if err != nil {
			return err
		}
		// Enable paging.
		cr0, err = vcpu.GetControlRegister(kvm.CR0)
		if err != nil {
			return err
		}
		cr0 = cr0 | (1 << 31) // Paging enable.
		err = vcpu.SetControlRegister(kvm.CR0, cr0, true)
		if err != nil {
			return err
		}
	}
	// NOTE: For 64-bit kernels, we need to enable
	// real 64-bit mode. This means that the L bit in
	// the segments must be one, the Db bit must be
	// zero, and we set the LME bit in EFER (above).
	var lVal uint8
	var dVal uint8
	if is_64bit {
		lVal = 1
		dVal = 0
	} else {
		lVal = 0
		dVal = 1
	}

	// Load the VMCS segments.
	//
	// NOTE: These values are loaded into the VMCS
	// registers and are expected to match the descriptors
	// we've used above. Unfortunately the API format doesn't
	// match, so we need to duplicate some work here. Ah, well
	// at least the below serves as an explanation for what
	// the magic numbers in GDT_ENTRY() above mean.
	BootCs := kvm.SegmentValue{
		Base:     0,
		Limit:    0xffffffff,
		Selector: uint16(C.BootCsSelector), // @ 0x10
		Dpl:      0,                        // Privilege level (kernel).
		Db:       dVal,                     // 32-bit segment?
		G:        1,                        // Granularity (page).
		S:        1,                        // As per BOOT_CS (code/data).
		L:        lVal,                     // 64-bit extension.
		Type:     0xb,                      // As per BOOT_CS (access must be set).
		Present:  1,
	}
	BootDs := kvm.SegmentValue{
		Base:     0,
		Limit:    0xffffffff,
		Selector: uint16(C.BootDsSelector), // @ 0x18
		Dpl:      0,                        // Privilege level (kernel).
		Db:       1,                        // 32-bit segment?
		G:        1,                        // Granularity (page).
		S:        1,                        // As per BOOT_DS (code/data).
		L:        0,                        // 64-bit extension.
		Type:     0x3,                      // As per BOOT_DS (access must be set).
		Present:  1,
	}
	BootTr := kvm.SegmentValue{
		Base:     0,
		Limit:    0xffffffff,
		Selector: uint16(C.BootTrSelector), // @ 0x20
		Dpl:      0,                        // Privilege level (kernel).
		Db:       1,                        // 32-bit segment?
		G:        1,                        // Granularity (page).
		S:        0,                        // As per BOOT_TR (system).
		L:        0,                        // 64-bit extension.
		Type:     0xb,                      // As per BOOT_TR.
		Present:  1,
	}
	err = vcpu.SetSegment(kvm.CS, BootCs, true)
	if err != nil {
		return err
	}
	err = vcpu.SetSegment(kvm.DS, BootDs, true)
	if err != nil {
		return err
	}
	err = vcpu.SetSegment(kvm.ES, BootDs, true)
	if err != nil {
		return err
	}
	err = vcpu.SetSegment(kvm.FS, BootDs, true)
	if err != nil {
		return err
	}
	err = vcpu.SetSegment(kvm.GS, BootDs, true)
	if err != nil {
		return err
	}
	err = vcpu.SetSegment(kvm.SS, BootDs, true)
	if err != nil {
		return err
	}
	err = vcpu.SetSegment(kvm.TR, BootTr, true)
	if err != nil {
		return err
	}
	// Create our boot parameters.
	boot_addr, boot_data, err := model.Allocate(MemoryTypeUser, 0, model.Max(), kvm.PageSize, false)
	if err != nil {
		return err
	}
	err = SetupLinuxBootParams(model, boot_data, orig_boot_data, cmdline_addr, initrd_addr, initrd_len)
	if err != nil {
		return err
	}
	// Set our registers.
	// This is according to the Linux 32-bit boot protocol.
	log.Printf("loader: boot_params @ %08x.", boot_addr)
	err = vcpu.SetRegister(kvm.RSI, kvm.RegisterValue(boot_addr))
	if err != nil {
		return err
	}
	err = vcpu.SetRegister(kvm.RBP, 0)
	if err != nil {
		return err
	}
	err = vcpu.SetRegister(kvm.RDI, 0)
	if err != nil {
		return err
	}
	err = vcpu.SetRegister(kvm.RBX, 0)
	if err != nil {
		return err
	}
	// Jump to our entry point.
	err = vcpu.SetRegister(kvm.RIP, kvm.RegisterValue(entry_point))
	if err != nil {
		return err
	}

	// Make sure interrupts are disabled.
	// This will actually clear out all other flags.
	rflags, err := vcpu.GetRegister(kvm.RFLAGS)
	if err != nil {
		return err
	}
	rflags = rflags &^ (1 << 9) // Interrupts off.
	rflags = rflags | (1 << 1)  // Reserved 1.
	err = vcpu.SetRegister(kvm.RFLAGS, kvm.RegisterValue(rflags))
	if err != nil {
		return err
	}

	// We're done.
	return nil
}
