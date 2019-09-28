// +build linux
package kvm

// IOCTL calls.
//const int IoctlCreateIrqChip = KVM_CREATE_IRQCHIP;
//const int IoctlGetIrqChip = KVM_GET_IRQCHIP;
//const int IoctlSetIrqChip = KVM_SET_IRQCHIP;
//const int IoctlIrqLine = KVM_IRQ_LINE;
//const int IoctlSignalMsi = KVM_SIGNAL_MSI;
//const int IoctlGetLAPIC = KVM_GET_LAPIC;
//const int IoctlSetLAPIC = KVM_SET_LAPIC;
// Size of our lapic state.
//const int APICSize = KVM_APIC_REG_SIZE;
// We need to fudge the types for irq level.
// This is because of the extremely annoying semantics
// for accessing *unions* in Go. Basically it can't.
// See the description below in createIrqChip().
//struct irq_level {
//    __u32 irq;
//    __u32 level;
//};
//static int check_irq_level(void) {
//    if (sizeof(struct kvm_irq_level) != sizeof(struct irq_level)) {
//        return 1;
//    } else {
//        return 0;
//    }
//}

import (
	"errors"
	"syscall"
	"unsafe"
)

// IrqChip --
// The IrqChip state requires three different
// devices: pic1, pic2 and the I/O apic. Each
// of these devices can be represented with a
// simple blob of data (compatibility will be
// the responsibility of KVM internally).
type IRQChip struct {
	PIC1   []byte `json:"pic1"`
	PIC2   []byte `json:"pic2"`
	IOAPIC []byte `json:"ioapic"`
}

// LAPICState --
// Just a blob of data. KVM will be ensure
// forward-compatibility.
type LAPICState struct {
	Data []byte `json:"data"`
}

func (self *VirtualMachine) CreateIRQChip() error {
	// No parameters needed, just create the chip.
	// This is called as the VM is being created in
	// order to ensure that all future vcpus will have
	// their own local apic.
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(self.Fd), uintptr(IOCtlCreateIRQChip), 0)
	if e != 0 {
		return e
	}
	// Ugh. A bit of type-fudging. Because of the
	// way go handles unions, we use a custom type
	// for the Interrupt() function below. Let's just
	// check once that everything is sane.
	if C.checkIRQLevel() != 0 {
		return errors.New("KVM irq_level doesn't match expected!")
	}

	return nil
}

func LAPIC() Pointer  { return Pointer(0xfee00000) }
func IOAPIC() Pointer { return Pointer(0xfec00000) }

type IRQ struct {
	Value uint32
	Level uint32
}

func (vm *VirtualMachine) Interrupt(irq IRQ, level bool) error {
	// Prepare the IRQ.
	irq{
		Value: uint32(irq),
	}
	if irq.Value {
		irq.Level = uint32(1)
	} else {
		irq.Level = uint32(0)
	}
	// Execute the ioctl.
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vm.Fd), uintptr(IoctlIRQLine), uintptr(unsafe.Pointer(&irqLevel)))
	if e != 0 {
		return e
	}
	return nil
}

type Range struct {
	High uint32
	Low  uint32
}

type MSI struct {
	Address Range
	Data    uint32
	Flags   uint32
}

func (vm *VirtualMachine) SignalMSI(addr Pointer, data uint32, flags uint32) error {
	msi := MSI{
		Address: &Range{
			High: uint32(addr >> 32),
			Low:  uint32(addr & 0xffffffff),
		},
		Data:  uint32(data),
		Flags: uint32(flags),
	}
	// Execute the ioctl.
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vm.Fd), uintptr(IoctlSignalMSI), uintptr(unsafe.Pointer(&msi)))
	if e != 0 {
		return e
	}
	return nil
}

func (vm *VirtualMachine) GetIRQChip() (IRQChip, error) {
	var state IRQChip
	// Create our scratch buffer.
	// The expected layout of the structure is:
	//  u32     - chip_id
	//  u32     - pad
	//  byte[0] - data
	buffer := make([]byte, PageSize, PageSize)
	for i := 0; i < 3; i += 1 {
		// Set our chip_id.
		buffer[0] = byte(i)
		// Execute the ioctl.
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vm.Fd), uintptr(IOCtlGetIRQChip), uintptr(unsafe.Pointer(&buffer[0])))
		if e != 0 {
			return state, e
		}
		// Copy appropriate state out.
		switch buffer[0] {
		case 0:
			state.PIC1 = make([]byte, len(buffer)-8, len(buffer)-8)
			copy(state.PIC1, buffer[8:])
		case 1:
			state.PIC2 = make([]byte, len(buf)-8, len(buf)-8)
			copy(state.PIC2, buf[8:])
		case 2:
			state.IOAPIC = make([]byte, len(buf)-8, len(buf)-8)
			copy(state.IOAPIC, buf[8:])
		}
	}
	return state, nil
}

func (vm *VirtualMachine) SetIRQChip(state IRQChip) error {
	// Create our scratch buffer.
	// See GetIrqChip for expected layout.
	buffer := make([]byte, PageSize, PageSize)
	for i := 0; i < 3; i += 1 {
		// Set our chip_id.
		buffer[0] = byte(i)
		// Copy appropriate state in.
		// We also ensure that we have the
		// appropriate state to load. If we don't
		// it's fine, we just continue along.
		switch i {
		case 0:
			if state.PIC1 == nil {
				continue
			}
			copy(buffer[8:], state.PIC1)
		case 1:
			if state.PIC2 == nil {
				continue
			}
			copy(buffer[8:], state.PIC2)
		case 2:
			if state.IOAPIC == nil {
				continue
			}
			copy(buf[8:], state.IOAPIC)
		}
		// Execute the ioctl.
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vm.Fd), uintptr(IOCtlSetIRQChip), uintptr(unsafe.Pointer(&buffer[0])))
		if e != 0 {
			return e
		}
	}
	return nil
}

func (vcpu *VCPU) GetLAPIC() (LAPICState, error) {
	// Prepare the apic state.
	state := LAPICState{make([]byte, C.APICSize, C.APICSize)}
	// Execute the ioctl.
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vcpu.Fd), uintptr(IOCtlGetLAPIC), uintptr(unsafe.Pointer(&state.Data[0])))
	if e != 0 {
		return state, e
	}
	return state, nil
}

func (vcpu *VCPU) SetLAPIC(state LAPICState) error {
	// Is there any state to set?
	// We just eat this error, it's fine.
	if state.Data == nil {
		return nil
	}
	// Check the state is reasonable.
	if len(state.Data) != int(C.APICSize) {
		return LAPICIncompatible
	}
	// Execute the ioctl.
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vcpu.Fd), uintptr(IOCtlSetLAPIC), uintptr(unsafe.Pointer(&state.Data[0])))
	if e != 0 {
		return e
	}

	return nil
}
