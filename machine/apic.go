package machine

import (
	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

type APIC struct {
	BaseDevice
	// Our addresses.
	// At the moment, these are at fixed address.
	// But just check that they meet expectations.
	IOAPIC kvm.Pointer `json:"ioapic"`
	LAPIC  kvm.Pointer `json:"lapic"`
	// Our kvm APIC.
	State kvm.IRQChip `json:"state"`
}

func NewAPIC(info *DeviceInfo) (Device, error) {
	apic := new(APIC)
	// Figure out our APIC addresses.
	// See the note above re: fixed addresses.
	apic.IOAPIC = kvm.IOAPIC()
	apic.LAPIC = kvm.LAPIC()

	return apic, apic.init(info)
}

func (apic *APIC) Attach(vm *kvm.VirtualMachine, model *Model) error {
	// Reserve our IOAPIC and LAPIC.
	err := model.Reserve(vm, apic, MemoryTypeReserved, apic.IOAPIC, kvm.PageSize, nil)
	if err != nil {
		return err
	}
	err = model.Reserve(vm, apic, MemoryTypeReserved, apic.LAPIC, kvm.PageSize, nil)
	if err != nil {
		return err
	}
	// Create our irqchip.
	err = vm.CreateIrqChip()
	if err != nil {
		return err
	}
	// We're good.
	return nil
}

func (apic *APIC) Save(vm *kvm.VirtualMachine) error {
	var err error
	// Save our IrqChip state.
	apic.State, err = vm.GetIrqChip()
	if err != nil {
		return err
	}
	// We're good.
	return nil
}

func (apic *APIC) Load(vm *kvm.VirtualMachine) error { return vm.SetIrqChip(apic.State) }
