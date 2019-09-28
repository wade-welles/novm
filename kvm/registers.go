package kvm

import (
	"log"
	"syscall"
	"unsafe"
)

const (
	PageSize = 4096
)

func (vcpu *VCPU) RefreshRegisters(dirty bool) error {
	// Ensure that our registers are up-to-date.
	// NOTE: We don't use the sync registers capability
	// which will expose the registers via the shared page.
	// We don't really manipulate the registers often
	// beyond the bootloader, so there's really no sense
	// having the checks for dirty registers, etc.
	if !vcpu.regs_cached {
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vcpu.Fd), uintptr(C.IoctlGetRegs), uintptr(unsafe.Pointer(&vcpu.regs)))
		if e != 0 {
			return e
		}
		vcpu.regs_cached = true
	}
	if dirty {
		vcpu.regs_dirty = true
	}
	return nil
}

func (vcpu *VCPU) FlushRegisters() error {
	// Ensure that our registers are up-to-date.
	if vcpu.regs_dirty {
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vcpu.Fd), uintptr(C.IoctlSetRegs), uintptr(unsafe.Pointer(&vcpu.regs)))
		if e != 0 {
			return e
		}
		vcpu.regs_dirty = false
	}
	vcpu.regs_cached = false
	return nil
}

func (vcpu *VCPU) RefreshSpecialRegisters(dirty bool) error {
	// Ensure that our special registers are up-to-date.
	if !vcpu.sregs_cached {
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vcpu.Fd), uintptr(C.IoctlGetSRegs), uintptr(unsafe.Pointer(&vcpu.sregs)))
		if e != 0 {
			return e
		}
		vcpu.sregs_cached = true

		// We never attempt to inject an interrupt via
		// the interrupt_bitmap mechanism. We handle that
		// via other state functions and explicitly.
		C.clear_interrupt_bitmap(&vcpu.sregs)
	}

	if dirty {
		vcpu.sregs_dirty = true
	}
	return nil
}

func (vcpu *VCPU) FlushSpecialRegisters() error {
	// Ensure that our registers are up-to-date.
	if vcpu.sregs_dirty {
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vcpu.Fd), uintptr(C.IoctlSetSRegs), uintptr(unsafe.Pointer(&vcpu.sregs)))
		if e != 0 {
			return e
		}
		vcpu.sregs_dirty = false
	}

	vcpu.sregs_cached = false
	return nil
}

func (vcpu *VCPU) flushAllRegs() error {
	// Flush all registers.
	err := vcpu.flushSRegs()
	if err != nil {
		return err
	}
	return vcpu.flushRegs()
}

func (vcpu *VCPU) SetRegister(reg Register, val RegisterValue) error {
	err := vcpu.refreshRegs(true)
	if err != nil {
		return err
	}

	switch reg {
	case RAX:
		vcpu.regs.rax = C.__u64(val)
	case RBX:
		vcpu.regs.rbx = C.__u64(val)
	case RCX:
		vcpu.regs.rcx = C.__u64(val)
	case RDX:
		vcpu.regs.rdx = C.__u64(val)
	case RSI:
		vcpu.regs.rsi = C.__u64(val)
	case RDI:
		vcpu.regs.rdi = C.__u64(val)
	case RSP:
		vcpu.regs.rsp = C.__u64(val)
	case RBP:
		vcpu.regs.rbp = C.__u64(val)
	case R8:
		vcpu.regs.r8 = C.__u64(val)
	case R9:
		vcpu.regs.r9 = C.__u64(val)
	case R10:
		vcpu.regs.r10 = C.__u64(val)
	case R11:
		vcpu.regs.r11 = C.__u64(val)
	case R12:
		vcpu.regs.r12 = C.__u64(val)
	case R13:
		vcpu.regs.r13 = C.__u64(val)
	case R14:
		vcpu.regs.r14 = C.__u64(val)
	case R15:
		vcpu.regs.r15 = C.__u64(val)
	case RIP:
		vcpu.regs.rip = C.__u64(val)
	case RFLAGS:
		vcpu.regs.rflags = C.__u64(val)
	default:
		return UnknownRegister
	}

	return nil
}

func (vcpu *VCPU) GetRegister(reg Register) (RegisterValue, error) {
	err := vcpu.refreshRegs(false)
	if err != nil {
		return RegisterValue(0), err
	}

	switch reg {
	case RAX:
		return RegisterValue(vcpu.regs.rax), nil
	case RBX:
		return RegisterValue(vcpu.regs.rbx), nil
	case RCX:
		return RegisterValue(vcpu.regs.rcx), nil
	case RDX:
		return RegisterValue(vcpu.regs.rdx), nil
	case RSI:
		return RegisterValue(vcpu.regs.rsi), nil
	case RDI:
		return RegisterValue(vcpu.regs.rdi), nil
	case RSP:
		return RegisterValue(vcpu.regs.rsp), nil
	case RBP:
		return RegisterValue(vcpu.regs.rbp), nil
	case R8:
		return RegisterValue(vcpu.regs.r8), nil
	case R9:
		return RegisterValue(vcpu.regs.r9), nil
	case R10:
		return RegisterValue(vcpu.regs.r10), nil
	case R11:
		return RegisterValue(vcpu.regs.r11), nil
	case R12:
		return RegisterValue(vcpu.regs.r12), nil
	case R13:
		return RegisterValue(vcpu.regs.r13), nil
	case R14:
		return RegisterValue(vcpu.regs.r14), nil
	case R15:
		return RegisterValue(vcpu.regs.r15), nil
	case RIP:
		return RegisterValue(vcpu.regs.rip), nil
	case RFLAGS:
		return RegisterValue(vcpu.regs.rflags), nil
	}

	return RegisterValue(0), UnknownRegister
}

func (vcpu *VCPU) SetControlRegister(
	reg ControlRegister,
	val ControlRegisterValue,
	sync bool) error {

	err := vcpu.refreshSRegs(true)
	if err != nil {
		return err
	}

	switch reg {
	case CR0:
		vcpu.sregs.cr0 = C.__u64(val)
	case CR2:
		vcpu.sregs.cr2 = C.__u64(val)
	case CR3:
		vcpu.sregs.cr3 = C.__u64(val)
	case CR4:
		vcpu.sregs.cr4 = C.__u64(val)
	case CR8:
		vcpu.sregs.cr8 = C.__u64(val)
	case EFER:
		vcpu.sregs.efer = C.__u64(val)
	case APIC_BASE:
		vcpu.sregs.apic_base = C.__u64(val)
	default:
		return UnknownRegister
	}

	if sync {
		err = vcpu.flushSRegs()
		if err != nil {
			return err
		}
	}

	return nil
}

func (vcpu *VCPU) GetControlRegister(reg ControlRegister) (ControlRegisterValue, error) {
	err := vcpu.refreshSRegs(false)
	if err != nil {
		return ControlRegisterValue(0), err
	}

	switch reg {
	case CR0:
		return ControlRegisterValue(vcpu.sregs.cr0), nil
	case CR2:
		return ControlRegisterValue(vcpu.sregs.cr2), nil
	case CR3:
		return ControlRegisterValue(vcpu.sregs.cr3), nil
	case CR4:
		return ControlRegisterValue(vcpu.sregs.cr4), nil
	case CR8:
		return ControlRegisterValue(vcpu.sregs.cr8), nil
	case EFER:
		return ControlRegisterValue(vcpu.sregs.efer), nil
	case APIC_BASE:
		return ControlRegisterValue(vcpu.sregs.apic_base), nil
	}

	return ControlRegisterValue(0), UnknownRegister
}

func (vcpu *VCPU) SetSegment(
	seg Segment,
	val SegmentValue,
	sync bool) error {

	err := vcpu.refreshSRegs(true)
	if err != nil {
		return err
	}

	switch seg {
	case CS:
		vcpu.sregs.cs.base = C.__u64(val.Base)
		vcpu.sregs.cs.limit = C.__u32(val.Limit)
		vcpu.sregs.cs.selector = C.__u16(val.Selector)
		vcpu.sregs.cs._type = C.__u8(val.Type)
		vcpu.sregs.cs.present = C.__u8(val.Present)
		vcpu.sregs.cs.dpl = C.__u8(val.Dpl)
		vcpu.sregs.cs.db = C.__u8(val.Db)
		vcpu.sregs.cs.s = C.__u8(val.S)
		vcpu.sregs.cs.l = C.__u8(val.L)
		vcpu.sregs.cs.g = C.__u8(val.G)
		vcpu.sregs.cs.avl = C.__u8(val.Avl)
		vcpu.sregs.cs.unusable = C.__u8(^val.Present & 0x1)
	case DS:
		vcpu.sregs.ds.base = C.__u64(val.Base)
		vcpu.sregs.ds.limit = C.__u32(val.Limit)
		vcpu.sregs.ds.selector = C.__u16(val.Selector)
		vcpu.sregs.ds._type = C.__u8(val.Type)
		vcpu.sregs.ds.present = C.__u8(val.Present)
		vcpu.sregs.ds.dpl = C.__u8(val.Dpl)
		vcpu.sregs.ds.db = C.__u8(val.Db)
		vcpu.sregs.ds.s = C.__u8(val.S)
		vcpu.sregs.ds.l = C.__u8(val.L)
		vcpu.sregs.ds.g = C.__u8(val.G)
		vcpu.sregs.ds.avl = C.__u8(val.Avl)
		vcpu.sregs.ds.unusable = C.__u8(^val.Present & 0x1)
	case ES:
		vcpu.sregs.es.base = C.__u64(val.Base)
		vcpu.sregs.es.limit = C.__u32(val.Limit)
		vcpu.sregs.es.selector = C.__u16(val.Selector)
		vcpu.sregs.es._type = C.__u8(val.Type)
		vcpu.sregs.es.present = C.__u8(val.Present)
		vcpu.sregs.es.dpl = C.__u8(val.Dpl)
		vcpu.sregs.es.db = C.__u8(val.Db)
		vcpu.sregs.es.s = C.__u8(val.S)
		vcpu.sregs.es.l = C.__u8(val.L)
		vcpu.sregs.es.g = C.__u8(val.G)
		vcpu.sregs.es.avl = C.__u8(val.Avl)
		vcpu.sregs.es.unusable = C.__u8(^val.Present & 0x1)
	case FS:
		vcpu.sregs.fs.base = C.__u64(val.Base)
		vcpu.sregs.fs.limit = C.__u32(val.Limit)
		vcpu.sregs.fs.selector = C.__u16(val.Selector)
		vcpu.sregs.fs._type = C.__u8(val.Type)
		vcpu.sregs.fs.present = C.__u8(val.Present)
		vcpu.sregs.fs.dpl = C.__u8(val.Dpl)
		vcpu.sregs.fs.db = C.__u8(val.Db)
		vcpu.sregs.fs.s = C.__u8(val.S)
		vcpu.sregs.fs.l = C.__u8(val.L)
		vcpu.sregs.fs.g = C.__u8(val.G)
		vcpu.sregs.fs.avl = C.__u8(val.Avl)
		vcpu.sregs.fs.unusable = C.__u8(^val.Present & 0x1)
	case GS:
		vcpu.sregs.gs.base = C.__u64(val.Base)
		vcpu.sregs.gs.limit = C.__u32(val.Limit)
		vcpu.sregs.gs.selector = C.__u16(val.Selector)
		vcpu.sregs.gs._type = C.__u8(val.Type)
		vcpu.sregs.gs.present = C.__u8(val.Present)
		vcpu.sregs.gs.dpl = C.__u8(val.Dpl)
		vcpu.sregs.gs.db = C.__u8(val.Db)
		vcpu.sregs.gs.s = C.__u8(val.S)
		vcpu.sregs.gs.l = C.__u8(val.L)
		vcpu.sregs.gs.g = C.__u8(val.G)
		vcpu.sregs.gs.avl = C.__u8(val.Avl)
		vcpu.sregs.gs.unusable = C.__u8(^val.Present & 0x1)
	case SS:
		vcpu.sregs.ss.base = C.__u64(val.Base)
		vcpu.sregs.ss.limit = C.__u32(val.Limit)
		vcpu.sregs.ss.selector = C.__u16(val.Selector)
		vcpu.sregs.ss._type = C.__u8(val.Type)
		vcpu.sregs.ss.present = C.__u8(val.Present)
		vcpu.sregs.ss.dpl = C.__u8(val.Dpl)
		vcpu.sregs.ss.db = C.__u8(val.Db)
		vcpu.sregs.ss.s = C.__u8(val.S)
		vcpu.sregs.ss.l = C.__u8(val.L)
		vcpu.sregs.ss.g = C.__u8(val.G)
		vcpu.sregs.ss.avl = C.__u8(val.Avl)
		vcpu.sregs.ss.unusable = C.__u8(^val.Present & 0x1)
	case TR:
		vcpu.sregs.tr.base = C.__u64(val.Base)
		vcpu.sregs.tr.limit = C.__u32(val.Limit)
		vcpu.sregs.tr.selector = C.__u16(val.Selector)
		vcpu.sregs.tr._type = C.__u8(val.Type)
		vcpu.sregs.tr.present = C.__u8(val.Present)
		vcpu.sregs.tr.dpl = C.__u8(val.Dpl)
		vcpu.sregs.tr.db = C.__u8(val.Db)
		vcpu.sregs.tr.s = C.__u8(val.S)
		vcpu.sregs.tr.l = C.__u8(val.L)
		vcpu.sregs.tr.g = C.__u8(val.G)
		vcpu.sregs.tr.avl = C.__u8(val.Avl)
		vcpu.sregs.tr.unusable = C.__u8(^val.Present & 0x1)
	case LDT:
		vcpu.sregs.ldt.base = C.__u64(val.Base)
		vcpu.sregs.ldt.limit = C.__u32(val.Limit)
		vcpu.sregs.ldt.selector = C.__u16(val.Selector)
		vcpu.sregs.ldt._type = C.__u8(val.Type)
		vcpu.sregs.ldt.present = C.__u8(val.Present)
		vcpu.sregs.ldt.dpl = C.__u8(val.Dpl)
		vcpu.sregs.ldt.db = C.__u8(val.Db)
		vcpu.sregs.ldt.s = C.__u8(val.S)
		vcpu.sregs.ldt.l = C.__u8(val.L)
		vcpu.sregs.ldt.g = C.__u8(val.G)
		vcpu.sregs.ldt.avl = C.__u8(val.Avl)
		vcpu.sregs.ldt.unusable = C.__u8(^val.Present & 0x1)
	default:
		return UnknownRegister
	}

	if sync {
		err = vcpu.flushSRegs()
		if err != nil {
			return err
		}
	}

	return nil
}

func (vcpu *VCPU) GetSegment(seg Segment) (SegmentValue, error) {
	err := vcpu.refreshSRegs(false)
	if err != nil {
		return SegmentValue{}, err
	}

	switch seg {
	case CS:
		return SegmentValue{
				Base:     uint64(vcpu.sregs.cs.base),
				Limit:    uint32(vcpu.sregs.cs.limit),
				Selector: uint16(vcpu.sregs.cs.selector),
				Type:     uint8(vcpu.sregs.cs._type),
				Present:  uint8(vcpu.sregs.cs.present),
				Dpl:      uint8(vcpu.sregs.cs.dpl),
				Db:       uint8(vcpu.sregs.cs.db),
				S:        uint8(vcpu.sregs.cs.s),
				L:        uint8(vcpu.sregs.cs.l),
				G:        uint8(vcpu.sregs.cs.g),
				Avl:      uint8(vcpu.sregs.cs.avl)},
			nil
	case DS:
		return SegmentValue{
				Base:     uint64(vcpu.sregs.ds.base),
				Limit:    uint32(vcpu.sregs.ds.limit),
				Selector: uint16(vcpu.sregs.ds.selector),
				Type:     uint8(vcpu.sregs.ds._type),
				Present:  uint8(vcpu.sregs.ds.present),
				Dpl:      uint8(vcpu.sregs.ds.dpl),
				Db:       uint8(vcpu.sregs.ds.db),
				S:        uint8(vcpu.sregs.ds.s),
				L:        uint8(vcpu.sregs.ds.l),
				G:        uint8(vcpu.sregs.ds.g),
				Avl:      uint8(vcpu.sregs.ds.avl)},
			nil
	case ES:
		return SegmentValue{
				Base:     uint64(vcpu.sregs.es.base),
				Limit:    uint32(vcpu.sregs.es.limit),
				Selector: uint16(vcpu.sregs.es.selector),
				Type:     uint8(vcpu.sregs.es._type),
				Present:  uint8(vcpu.sregs.es.present),
				Dpl:      uint8(vcpu.sregs.es.dpl),
				Db:       uint8(vcpu.sregs.es.db),
				S:        uint8(vcpu.sregs.es.s),
				L:        uint8(vcpu.sregs.es.l),
				G:        uint8(vcpu.sregs.es.g),
				Avl:      uint8(vcpu.sregs.es.avl)},
			nil
	case FS:
		return SegmentValue{
				Base:     uint64(vcpu.sregs.fs.base),
				Limit:    uint32(vcpu.sregs.fs.limit),
				Selector: uint16(vcpu.sregs.fs.selector),
				Type:     uint8(vcpu.sregs.fs._type),
				Present:  uint8(vcpu.sregs.fs.present),
				Dpl:      uint8(vcpu.sregs.fs.dpl),
				Db:       uint8(vcpu.sregs.fs.db),
				S:        uint8(vcpu.sregs.fs.s),
				L:        uint8(vcpu.sregs.fs.l),
				G:        uint8(vcpu.sregs.fs.g),
				Avl:      uint8(vcpu.sregs.fs.avl)},
			nil
	case GS:
		return SegmentValue{
				Base:     uint64(vcpu.sregs.gs.base),
				Limit:    uint32(vcpu.sregs.gs.limit),
				Selector: uint16(vcpu.sregs.gs.selector),
				Type:     uint8(vcpu.sregs.gs._type),
				Present:  uint8(vcpu.sregs.gs.present),
				Dpl:      uint8(vcpu.sregs.gs.dpl),
				Db:       uint8(vcpu.sregs.gs.db),
				S:        uint8(vcpu.sregs.gs.s),
				L:        uint8(vcpu.sregs.gs.l),
				G:        uint8(vcpu.sregs.gs.g),
				Avl:      uint8(vcpu.sregs.gs.avl)},
			nil
	case SS:
		return SegmentValue{
				Base:     uint64(vcpu.sregs.ss.base),
				Limit:    uint32(vcpu.sregs.ss.limit),
				Selector: uint16(vcpu.sregs.ss.selector),
				Type:     uint8(vcpu.sregs.ss._type),
				Present:  uint8(vcpu.sregs.ss.present),
				Dpl:      uint8(vcpu.sregs.ss.dpl),
				Db:       uint8(vcpu.sregs.ss.db),
				S:        uint8(vcpu.sregs.ss.s),
				L:        uint8(vcpu.sregs.ss.l),
				G:        uint8(vcpu.sregs.ss.g),
				Avl:      uint8(vcpu.sregs.ss.avl)},
			nil
	case TR:
		return SegmentValue{
				Base:     uint64(vcpu.sregs.tr.base),
				Limit:    uint32(vcpu.sregs.tr.limit),
				Selector: uint16(vcpu.sregs.tr.selector),
				Type:     uint8(vcpu.sregs.tr._type),
				Present:  uint8(vcpu.sregs.tr.present),
				Dpl:      uint8(vcpu.sregs.tr.dpl),
				Db:       uint8(vcpu.sregs.tr.db),
				S:        uint8(vcpu.sregs.tr.s),
				L:        uint8(vcpu.sregs.tr.l),
				G:        uint8(vcpu.sregs.tr.g),
				Avl:      uint8(vcpu.sregs.tr.avl)},
			nil
	case LDT:
		return SegmentValue{
				Base:     uint64(vcpu.sregs.ldt.base),
				Limit:    uint32(vcpu.sregs.ldt.limit),
				Selector: uint16(vcpu.sregs.ldt.selector),
				Type:     uint8(vcpu.sregs.ldt._type),
				Present:  uint8(vcpu.sregs.ldt.present),
				Dpl:      uint8(vcpu.sregs.ldt.dpl),
				Db:       uint8(vcpu.sregs.ldt.db),
				S:        uint8(vcpu.sregs.ldt.s),
				L:        uint8(vcpu.sregs.ldt.l),
				G:        uint8(vcpu.sregs.ldt.g),
				Avl:      uint8(vcpu.sregs.ldt.avl)},
			nil
	}

	return SegmentValue{}, UnknownRegister
}

func (vcpu *VCPU) SetDescriptor(desc Descriptor, val DescriptorValue, sync bool) error {
	err := vcpu.refreshSRegs(true)
	if err != nil {
		return err
	}
	switch desc {
	case GDT:
		vcpu.sregs.gdt.base = C.__u64(val.Base)
		vcpu.sregs.gdt.limit = C.__u16(val.Limit)
	case IDT:
		vcpu.sregs.idt.base = C.__u64(val.Base)
		vcpu.sregs.idt.limit = C.__u16(val.Limit)
	default:
		return UnknownRegister
	}

	if sync {
		err = vcpu.flushSRegs()
		if err != nil {
			return err
		}
	}

	return nil
}

func (vcpu *VCPU) GetDescriptor(desc Descriptor) (DescriptorValue, error) {
	err := vcpu.refreshSRegs(false)
	if err != nil {
		return DescriptorValue{}, err
	}

	switch desc {
	case GDT:
		return DescriptorValue{
				Base:  uint64(vcpu.sregs.gdt.base),
				Limit: uint16(vcpu.sregs.gdt.limit)},
			nil
	case IDT:
		return DescriptorValue{
				Base:  uint64(vcpu.sregs.idt.base),
				Limit: uint16(vcpu.sregs.idt.limit)},
			nil
	}

	return DescriptorValue{}, UnknownRegister
}

func (vm *VirtualMachine) SizeSpecialMemory() uint64 {
	return 4 * PageSize
}

func (vm *VirtualMachine) MapSpecialMemory(addr Pointer) error {
	// We require 1 page for the identity map.
	err := vm.MapReservedMemory(addr, PageSize)
	if err != nil {
		return err
	}
	// Set the EPT identity map.
	// (This requires a single page).
	ept_identity_addr := C.__u64(addr)
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vm.fd),
		uintptr(IoctlSetIdentityMapAddr),
		uintptr(unsafe.Pointer(&ept_identity_addr)))
	if e != 0 {
		log.Printf("Unable to set identity map to %08x!", addr)
		return e
	}

	// We require 3 pages for the TSS address.
	err = vm.MapReservedMemory(addr+PageSize, 3*PageSize)
	if err != nil {
		return err
	}

	// Set the TSS address to above.
	// (This requires three pages).
	_, _, e = syscall.Syscall(syscall.SYS_IOCTL, uintptr(vm.Fd), uintptr(C.IoctlSetTssAddr), uintptr(addr+PageSize))
	if e != 0 {
		log.Printf("Unable to set TSS ADDR to %08x!", addr+PageSize)
		return e
	}

	// We're okay.
	return nil
}

//
// Our general purpose registers.
//
type Register int
type RegisterValue uint64

const (
	RAX Register = iota
	RBX
	RCX
	RDX
	RSI
	RDI
	RSP
	RBP
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
	RIP
	RFLAGS
)

//
// Special control registers.
//
type ControlRegister int
type ControlRegisterValue uint64

const (
	CR0 ControlRegister = iota
	CR2
	CR3
	CR4
	CR8
	EFER
	APIC_BASE
)

//
// Segment descriptor registers.
//
type Descriptor int
type DescriptorValue struct {
	Base  uint64 `json:"base"`
	Limit uint16 `json:"limit"`
}

const (
	GDT Descriptor = iota
	IDT
)

//
// Segment registers.
//
type Segment int
type SegmentValue struct {
	Base     uint64 `json:"base"`
	Limit    uint32 `json:"limit"`
	Selector uint16 `json:"selector"`
	Type     uint8  `json:"type"`
	Present  uint8  `json:"present"`
	Dpl      uint8  `json:"dpl"`
	Db       uint8  `json:"db"`
	L        uint8  `json:"l"`
	S        uint8  `json:"s"`
	G        uint8  `json:"g"`
	Avl      uint8  `json:"avl"`
}

const (
	CS Segment = iota
	DS
	ES
	FS
	GS
	SS
	TR
	LDT
)

//
// Utility structure containing all registers.
//
type Registers struct {
	RAX    *RegisterValue
	RBX    *RegisterValue
	RCX    *RegisterValue
	RDX    *RegisterValue
	RSI    *RegisterValue
	RDI    *RegisterValue
	RSP    *RegisterValue
	RBP    *RegisterValue
	R8     *RegisterValue
	R9     *RegisterValue
	R10    *RegisterValue
	R11    *RegisterValue
	R12    *RegisterValue
	R13    *RegisterValue
	R14    *RegisterValue
	R15    *RegisterValue
	RIP    *RegisterValue
	RFLAGS *RegisterValue

	CR0       *ControlRegisterValue
	CR2       *ControlRegisterValue
	CR3       *ControlRegisterValue
	CR4       *ControlRegisterValue
	CR8       *ControlRegisterValue
	EFER      *ControlRegisterValue
	APIC_BASE *ControlRegisterValue `json:"APIC"`

	IDT *DescriptorValue
	GDT *DescriptorValue

	CS  *SegmentValue
	DS  *SegmentValue
	ES  *SegmentValue
	FS  *SegmentValue
	GS  *SegmentValue
	SS  *SegmentValue
	TR  *SegmentValue
	LDT *SegmentValue
}

func (vcpu *VCPU) getRegister(
	name string,
	reg Register,
	errs []error) (*RegisterValue, []error) {

	value, err := vcpu.GetRegister(reg)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	return &value, errs
}

func (vcpu *VCPU) getControlRegister(
	name string,
	reg ControlRegister,
	errs []error) (*ControlRegisterValue, []error) {

	value, err := vcpu.GetControlRegister(reg)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	return &value, errs
}

func (vcpu *VCPU) getDescriptor(
	name string,
	desc Descriptor,
	errs []error) (*DescriptorValue, []error) {

	value, err := vcpu.GetDescriptor(desc)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	return &value, errs
}

func (vcpu *VCPU) getSegment(
	name string,
	seg Segment,
	errs []error) (*SegmentValue, []error) {

	value, err := vcpu.GetSegment(seg)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	return &value, errs
}

func (vcpu *VCPU) GetRegisters() (Registers, error) {
	vcpu.Pause(false)
	defer vcpu.Unpause(false)

	var regs Registers
	errs := make([]error, 0, 0)

	regs.RAX, errs = vcpu.getRegister("RAX", RAX, errs)
	regs.RBX, errs = vcpu.getRegister("RBX", RBX, errs)
	regs.RCX, errs = vcpu.getRegister("RCX", RCX, errs)
	regs.RDX, errs = vcpu.getRegister("RDX", RDX, errs)
	regs.RSI, errs = vcpu.getRegister("RSI", RSI, errs)
	regs.RDI, errs = vcpu.getRegister("RDI", RDI, errs)
	regs.RSP, errs = vcpu.getRegister("RSP", RSP, errs)
	regs.RBP, errs = vcpu.getRegister("RBP", RBP, errs)
	regs.R8, errs = vcpu.getRegister("R8", R8, errs)
	regs.R9, errs = vcpu.getRegister("R9", R9, errs)
	regs.R10, errs = vcpu.getRegister("R10", R10, errs)
	regs.R11, errs = vcpu.getRegister("R11", R11, errs)
	regs.R12, errs = vcpu.getRegister("R12", R12, errs)
	regs.R13, errs = vcpu.getRegister("R13", R13, errs)
	regs.R14, errs = vcpu.getRegister("R14", R14, errs)
	regs.R15, errs = vcpu.getRegister("R15", R15, errs)
	regs.RIP, errs = vcpu.getRegister("RIP", RIP, errs)
	regs.RFLAGS, errs = vcpu.getRegister("RFLAGS", RFLAGS, errs)

	regs.CR0, errs = vcpu.getControlRegister("CR0", CR0, errs)
	regs.CR2, errs = vcpu.getControlRegister("CR2", CR2, errs)
	regs.CR3, errs = vcpu.getControlRegister("CR3", CR3, errs)
	regs.CR4, errs = vcpu.getControlRegister("CR4", CR4, errs)
	regs.CR8, errs = vcpu.getControlRegister("CR8", CR8, errs)
	regs.EFER, errs = vcpu.getControlRegister("EFER", EFER, errs)
	regs.APIC_BASE, errs = vcpu.getControlRegister("APIC_BASE", APIC_BASE, errs)

	regs.GDT, errs = vcpu.getDescriptor("GDT", GDT, errs)
	regs.IDT, errs = vcpu.getDescriptor("IDT", IDT, errs)

	regs.CS, errs = vcpu.getSegment("CS", CS, errs)
	regs.DS, errs = vcpu.getSegment("DS", DS, errs)
	regs.ES, errs = vcpu.getSegment("ES", ES, errs)
	regs.FS, errs = vcpu.getSegment("FS", FS, errs)
	regs.GS, errs = vcpu.getSegment("GS", GS, errs)
	regs.SS, errs = vcpu.getSegment("SS", SS, errs)
	regs.TR, errs = vcpu.getSegment("TR", TR, errs)
	regs.LDT, errs = vcpu.getSegment("LDT", LDT, errs)

	// Return a simple error.
	// We could actually return a more
	// meaningful error here that describes
	// all the registers which had errors,
	// but for now this will do the trick.
	for _, err := range errs {
		if err != nil {
			return Registers{}, err
		}
	}

	return regs, nil
}

func (vcpu *VCPU) setRegister(
	name string,
	reg Register,
	value *RegisterValue,
	errs []error) []error {

	if value != nil {
		err := vcpu.SetRegister(reg, *value)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (vcpu *VCPU) setControlRegister(
	name string,
	reg ControlRegister,
	value *ControlRegisterValue,
	errs []error) []error {

	if value != nil {
		err := vcpu.SetControlRegister(reg, *value, false)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (vcpu *VCPU) setDescriptor(
	name string,
	desc Descriptor,
	value *DescriptorValue,
	errs []error) []error {

	if value != nil {
		err := vcpu.SetDescriptor(desc, *value, false)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (vcpu *VCPU) setSegment(
	name string,
	seg Segment,
	value *SegmentValue,
	errs []error) []error {

	if value != nil {
		err := vcpu.SetSegment(seg, *value, false)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (vcpu *VCPU) SetRegisters(regs Registers) error {
	vcpu.Pause(false)
	defer vcpu.Unpause(false)

	errs := make([]error, 0, 0)

	errs = vcpu.setSegment("CS", CS, regs.CS, errs)
	errs = vcpu.setSegment("DS", DS, regs.DS, errs)
	errs = vcpu.setSegment("ES", ES, regs.ES, errs)
	errs = vcpu.setSegment("FS", FS, regs.FS, errs)
	errs = vcpu.setSegment("GS", GS, regs.GS, errs)
	errs = vcpu.setSegment("SS", SS, regs.SS, errs)
	errs = vcpu.setSegment("TR", TR, regs.TR, errs)
	errs = vcpu.setSegment("LDT", LDT, regs.LDT, errs)

	errs = vcpu.setRegister("RAX", RAX, regs.RAX, errs)
	errs = vcpu.setRegister("RBX", RBX, regs.RBX, errs)
	errs = vcpu.setRegister("RCX", RCX, regs.RCX, errs)
	errs = vcpu.setRegister("RDX", RDX, regs.RDX, errs)
	errs = vcpu.setRegister("RSI", RSI, regs.RSI, errs)
	errs = vcpu.setRegister("RDI", RDI, regs.RDI, errs)
	errs = vcpu.setRegister("RSP", RSP, regs.RSP, errs)
	errs = vcpu.setRegister("RBP", RBP, regs.RBP, errs)
	errs = vcpu.setRegister("R8", R8, regs.R8, errs)
	errs = vcpu.setRegister("R9", R9, regs.R9, errs)
	errs = vcpu.setRegister("R10", R10, regs.R10, errs)
	errs = vcpu.setRegister("R11", R11, regs.R11, errs)
	errs = vcpu.setRegister("R12", R12, regs.R12, errs)
	errs = vcpu.setRegister("R13", R13, regs.R13, errs)
	errs = vcpu.setRegister("R14", R14, regs.R14, errs)
	errs = vcpu.setRegister("R15", R15, regs.R15, errs)
	errs = vcpu.setRegister("RIP", RIP, regs.RIP, errs)
	errs = vcpu.setRegister("RFLAGS", RFLAGS, regs.RFLAGS, errs)

	errs = vcpu.setControlRegister("CR0", CR0, regs.CR0, errs)
	errs = vcpu.setControlRegister("CR2", CR2, regs.CR2, errs)
	errs = vcpu.setControlRegister("CR3", CR3, regs.CR3, errs)
	errs = vcpu.setControlRegister("CR4", CR4, regs.CR4, errs)
	errs = vcpu.setControlRegister("CR8", CR8, regs.CR8, errs)
	errs = vcpu.setControlRegister("EFER", EFER, regs.EFER, errs)
	errs = vcpu.setControlRegister("APIC_BASE", APIC_BASE, regs.APIC_BASE, errs)

	errs = vcpu.setDescriptor("GDT", GDT, regs.GDT, errs)
	errs = vcpu.setDescriptor("IDT", IDT, regs.IDT, errs)

	// As per GetRegisters(), return a simple error.
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return vcpu.flushAllRegs()
}
