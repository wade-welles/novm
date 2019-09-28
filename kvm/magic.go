package kvm

///// Magic addresses externally used to lay out x86_64 VMs.
//
///// Initial stack for the boot CPU.
//pub const BOOT_STACK_POINTER: usize = 0x8ff0;
//
///// Kernel command line start address.
//pub const CMDLINE_START: usize = 0x20000;
///// Kernel command line start address maximum size.
//pub const CMDLINE_MAX_SIZE: usize = 0x10000;
//
///// Start of the high memory.
//pub const HIMEM_START: usize = 0x0010_0000; //1 MB.
//
//// Typically, on x86 systems 16 IRQs are used (0-15).
///// First usable IRQ ID for virtio device interrupts on x86_64.
//pub const IRQ_BASE: u32 = 5;
///// Last usable IRQ ID for virtio device interrupts on x86_64.
//pub const IRQ_MAX: u32 = 15;
//
///// Address for the TSS setup.
//pub const KVM_TSS_ADDRESS: usize = 0xfffb_d000;
//
///// The 'zero page', a.k.a linux kernel bootparams.
//pub const ZERO_PAGE_START: usize = 0x7000;

/////////////////////////////////////////////////
// [has vcpu code]
//https://github.com/firecracker-microvm/firecracker/blob/5ef557c80d2f103a66ee9de615f38a21cb8480d6/vmm/src/vstate.rs
// [has apic code]
// https://github.com/firecracker-microvm/firecracker/blob/4f42a7ef02907a7f2ec5b35232d22b5666110922/arch/src/x86_64/interrupts.rs
// [has cpuid code]
// https://github.com/firecracker-microvm/firecracker/blob/7cf1828daa82e576c612b7fd82972058b6d0b983/cpuid/src/transformer/amd.rsa
// [build kvm segment from gdt]
// https://github.com/firecracker-microvm/firecracker/blob/bef292499e2ad15ffd3229f7369564605a336b2c/arch/src/x86_64/gdt.rs
// Initial pagetables.
//const PML4_START: usize = 0x9000;
//const PDPTE_START: usize = 0xa000;
//const PDE_START: usize = 0xb000;
//////////////////////////////////////////////////////////////////////////////////////////////////////////
//// https://github.com/firecracker-microvm/firecracker/blob/bef292499e2ad15ffd3229f7369564605a336b2c/arch/src/x86_64/mptable.rs
// Most of these variables are sourced from the Intel MP Spec 1.4.
//const SMP_MAGIC_IDENT: [c_char; 4] = char_array!(c_char; '_', 'M', 'P', '_');
//const MPC_SIGNATURE: [c_char; 4] = char_array!(c_char; 'P', 'C', 'M', 'P');
//const MPC_SPEC: i8 = 4;
//const MPC_OEM: [c_char; 8] = char_array!(c_char; 'F', 'C', ' ', ' ', ' ', ' ', ' ', ' ');
//const MPC_PRODUCT_ID: [c_char; 12] = ['0' as c_char; 12];
//const BUS_TYPE_ISA: [u8; 6] = char_array!(u8; 'I', 'S', 'A', ' ', ' ', ' ');
//const IO_APIC_DEFAULT_PHYS_BASE: u32 = 0xfec0_0000; // source: linux/arch/x86/include/asm/apicdef.h
//const APIC_DEFAULT_PHYS_BASE: u32 = 0xfee0_0000; // source: linux/arch/x86/include/asm/apicdef.h
//const APIC_VERSION: u8 = 0x14;
//const CPU_STEPPING: u32 = 0x600;
//const CPU_FEATURE_APIC: u32 = 0x200;
//const CPU_FEATURE_FPU: u32 = 0x001;
///////////////////////////////////////////////////////////////////////////////////////////////////////
// IOC

//pub const _IOC_NRBITS: c_uint = 8;
//pub const _IOC_TYPEBITS: c_uint = 8;
//pub const _IOC_SIZEBITS: c_uint = 14;
//pub const _IOC_DIRBITS: c_uint = 2;
//pub const _IOC_NRMASK: c_uint = 255;
//pub const _IOC_TYPEMASK: c_uint = 255;
//pub const _IOC_SIZEMASK: c_uint = 16383;
//pub const _IOC_DIRMASK: c_uint = 3;
//pub const _IOC_NRSHIFT: c_uint = 0;
//pub const _IOC_TYPESHIFT: c_uint = 8;
//pub const _IOC_SIZESHIFT: c_uint = 16;
//pub const _IOC_DIRSHIFT: c_uint = 30;
//pub const _IOC_NONE: c_uint = 0;
//pub const _IOC_WRITE: c_uint = 1;
//pub const _IOC_READ: c_uint = 2;
//pub const IOC_IN: c_uint = 1_073_741_824;
//pub const IOC_OUT: c_uint = 2_147_483_648;
//pub const IOC_INOUT: c_uint = 3_221_225_472;
//pub const IOCSIZE_MASK: c_uint = 1_073_676_288;
//pub const IOCSIZE_SHIFT: c_uint = 16;

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

/// Configure Floating-Point Unit (FPU) registers for a given CPU.
///
/// # Arguments
///
/// * `vcpu` - Structure for the VCPU that holds the VCPU's fd.
//pub fn setup_fpu(vcpu: &VcpuFd) -> Result<()> {
//    let fpu: kvm_fpu = kvm_fpu {
//        fcw: 0x37f,
//        mxcsr: 0x1f80,
//        ..Default::default()
//    };
//
//    vcpu.set_fpu(&fpu).map_err(Error::SetFPURegisters)
//}
//
///// Configure Model Specific Registers (MSRs) for a given CPU.
/////
///// # Arguments
/////
///// * `vcpu` - Structure for the VCPU that holds the VCPU's fd.
//pub fn setup_msrs(vcpu: &VcpuFd) -> Result<()> {
//    let entry_vec = create_msr_entries();
//    let vec_size_bytes =
//        mem::size_of::<kvm_msrs>() + (entry_vec.len() * mem::size_of::<kvm_msr_entry>());
//    let vec: Vec<u8> = Vec::with_capacity(vec_size_bytes);
//    #[allow(clippy::cast_ptr_alignment)]
//    let msrs: &mut kvm_msrs = unsafe {
//        // Converting the vector's memory to a struct is unsafe.  Carefully using the read-only
//        // vector to size and set the members ensures no out-of-bounds errors below.
//        &mut *(vec.as_ptr() as *mut kvm_msrs)
//    };
//
//    unsafe {
//        // Mapping the unsized array to a slice is unsafe because the length isn't known.
//        // Providing the length used to create the struct guarantees the entire slice is valid.
//        let entries: &mut [kvm_msr_entry] = msrs.entries.as_mut_slice(entry_vec.len());
//        entries.copy_from_slice(&entry_vec);
//    }
//    msrs.nmsrs = entry_vec.len() as u32;
//
//    vcpu.set_msrs(msrs)
//        .map_err(Error::SetModelSpecificRegisters)
//}
//
///// Configure base registers for a given CPU.
/////
///// # Arguments
/////
///// * `vcpu` - Structure for the VCPU that holds the VCPU's fd.
///// * `boot_ip` - Starting instruction pointer.
//pub fn setup_regs(vcpu: &VcpuFd, boot_ip: u64) -> Result<()> {
//    let regs: kvm_regs = kvm_regs {
//        rflags: 0x0000_0000_0000_0002u64,
//        rip: boot_ip,
//        // Frame pointer. It gets a snapshot of the stack pointer (rsp) so that when adjustments are
//        // made to rsp (i.e. reserving space for local variables or pushing values on to the stack),
//        // local variables and function parameters are still accessible from a constant offset from rbp.
//        rsp: super::layout::BOOT_STACK_POINTER as u64,
//        // Starting stack pointer.
//        rbp: super::layout::BOOT_STACK_POINTER as u64,
//        // Must point to zero page address per Linux ABI. This is x86_64 specific.
//        rsi: super::layout::ZERO_PAGE_START as u64,
//        ..Default::default()
//    };
//
//    vcpu.set_regs(&regs).map_err(Error::SetBaseRegisters)
//}
//
///// Configures the segment registers and system page tables for a given CPU.
/////
///// # Arguments
/////
///// * `mem` - The memory that will be passed to the guest.
///// * `vcpu` - Structure for the VCPU that holds the VCPU's fd.
//pub fn setup_sregs(mem: &GuestMemory, vcpu: &VcpuFd) -> Result<()> {
//    let mut sregs: kvm_sregs = vcpu.get_sregs().map_err(Error::GetStatusRegisters)?;
//
//    configure_segments_and_sregs(mem, &mut sregs)?;
//    setup_page_tables(mem, &mut sregs)?; // TODO(dgreid) - Can this be done once per system instead?
//
//    vcpu.set_sregs(&sregs).map_err(Error::SetStatusRegisters)
//}
//
//const BOOT_GDT_OFFSET: usize = 0x500;
//const BOOT_IDT_OFFSET: usize = 0x520;
//
//const BOOT_GDT_MAX: usize = 4;
//
//const EFER_LMA: u64 = 0x400;
//const EFER_LME: u64 = 0x100;
//
//const X86_CR0_PE: u64 = 0x1;
//const X86_CR0_PG: u64 = 0x8000_0000;
//const X86_CR4_PAE: u64 = 0x20;
//
//fn write_gdt_table(table: &[u64], guest_mem: &GuestMemory) -> Result<()> {
//    let boot_gdt_addr = GuestAddress(BOOT_GDT_OFFSET);
//    for (index, entry) in table.iter().enumerate() {
//        let addr = guest_mem
//            .checked_offset(boot_gdt_addr, index * mem::size_of::<u64>())
//            .ok_or(Error::WriteGDT)?;
//        guest_mem
//            .write_obj_at_addr(*entry, addr)
//            .map_err(|_| Error::WriteGDT)?;
//    }
//    Ok(())
//}
//
//fn write_idt_value(val: u64, guest_mem: &GuestMemory) -> Result<()> {
//    let boot_idt_addr = GuestAddress(BOOT_IDT_OFFSET);
//    guest_mem
//        .write_obj_at_addr(val, boot_idt_addr)
//        .map_err(|_| Error::WriteIDT)
//}
//
//fn configure_segments_and_sregs(mem: &GuestMemory, sregs: &mut kvm_sregs) -> Result<()> {
//    let gdt_table: [u64; BOOT_GDT_MAX as usize] = [
//        gdt_entry(0, 0, 0),            // NULL
//        gdt_entry(0xa09b, 0, 0xfffff), // CODE
//        gdt_entry(0xc093, 0, 0xfffff), // DATA
//        gdt_entry(0x808b, 0, 0xfffff), // TSS
//    ];
//
//    let code_seg = kvm_segment_from_gdt(gdt_table[1], 1);
//    let data_seg = kvm_segment_from_gdt(gdt_table[2], 2);
//    let tss_seg = kvm_segment_from_gdt(gdt_table[3], 3);
//
//    // Write segments
//    write_gdt_table(&gdt_table[..], mem)?;
//    sregs.gdt.base = BOOT_GDT_OFFSET as u64;
//    sregs.gdt.limit = mem::size_of_val(&gdt_table) as u16 - 1;
//
//    write_idt_value(0, mem)?;
//    sregs.idt.base = BOOT_IDT_OFFSET as u64;
//    sregs.idt.limit = mem::size_of::<u64>() as u16 - 1;
//
//    sregs.cs = code_seg;
//    sregs.ds = data_seg;
//    sregs.es = data_seg;
//    sregs.fs = data_seg;
//    sregs.gs = data_seg;
//    sregs.ss = data_seg;
//    sregs.tr = tss_seg;
//
//    /* 64-bit protected mode */
//    sregs.cr0 |= X86_CR0_PE;
//    sregs.efer |= EFER_LME | EFER_LMA;
//
//    Ok(())
//}
//
//fn setup_page_tables(mem: &GuestMemory, sregs: &mut kvm_sregs) -> Result<()> {
//    // Puts PML4 right after zero page but aligned to 4k.
//    let boot_pml4_addr = GuestAddress(PML4_START);
//    let boot_pdpte_addr = GuestAddress(PDPTE_START);
//    let boot_pde_addr = GuestAddress(PDE_START);
//
//    // Entry covering VA [0..512GB)
//    mem.write_obj_at_addr(boot_pdpte_addr.offset() as u64 | 0x03, boot_pml4_addr)
//        .map_err(|_| Error::WritePML4Address)?;
//
//    // Entry covering VA [0..1GB)
//    mem.write_obj_at_addr(boot_pde_addr.offset() as u64 | 0x03, boot_pdpte_addr)
//        .map_err(|_| Error::WritePDPTEAddress)?;
//    // 512 2MB entries together covering VA [0..1GB). Note we are assuming
//    // CPU supports 2MB pages (/proc/cpuinfo has 'pse'). All modern CPUs do.
//    for i in 0..512 {
//        mem.write_obj_at_addr(
//            (i << 21) + 0x83u64,
//            boot_pde_addr.unchecked_add((i * 8) as usize),
//        )
//        .map_err(|_| Error::WritePDEAddress)?;
//    }
//
//    sregs.cr3 = boot_pml4_addr.offset() as u64;
//    sregs.cr4 |= X86_CR4_PAE;
//    sregs.cr0 |= X86_CR0_PG;
//    Ok(())
//}
//
//fn create_msr_entries() -> Vec<kvm_msr_entry> {
//    let mut entries = Vec::<kvm_msr_entry>::new();
//
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_IA32_SYSENTER_CS,
//        data: 0x0,
//        ..Default::default()
//    });
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_IA32_SYSENTER_ESP,
//        data: 0x0,
//        ..Default::default()
//    });
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_IA32_SYSENTER_EIP,
//        data: 0x0,
//        ..Default::default()
//    });
//    // x86_64 specific msrs, we only run on x86_64 not x86.
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_STAR,
//        data: 0x0,
//        ..Default::default()
//    });
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_CSTAR,
//        data: 0x0,
//        ..Default::default()
//    });
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_KERNEL_GS_BASE,
//        data: 0x0,
//        ..Default::default()
//    });
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_SYSCALL_MASK,
//        data: 0x0,
//        ..Default::default()
//    });
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_LSTAR,
//        data: 0x0,
//        ..Default::default()
//    });
//    // end of x86_64 specific code
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_IA32_TSC,
//        data: 0x0,
//        ..Default::default()
//    });
//    entries.push(kvm_msr_entry {
//        index: msr_index::MSR_IA32_MISC_ENABLE,
//        data: u64::from(msr_index::MSR_IA32_MISC_ENABLE_FAST_STRING),
//        ..Default::default()
//    });
//
//    entries
//	}
