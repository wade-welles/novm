package kvm

import (
	"sync"
)

type VirtualMachine struct {
	Fd           uintptr
	MmapSize     int
	NextSlot     uint32
	MappingCache sync.Map
	Mutex        sync.RWMutex
	Available    sync.Cond
	VCPUs        map[uint64]*VCPU
	MaxVCPUs     int
	MemoryRegion int
	CPUID        []CPUID
	MSRs         []uint32
}

func (self *KVM) NewVM() (*VirtualMachine, error) {
	virtualMachine := &VirtualMachine{
		VCPUs: make([]*VCPU, 0, 0),
	}
	self.fd, err = self.OpenKVM()
	if err != nil {
		panic("[fatal] could not access KVM kernel module:", err)
	}
	defer syscall.Close(self.fd)
	// Check API version.
	// Check our extensions.
	for _, capSpec := range requiredCapabilities {
		err = checkCapability(fd, capSpec)
		if err != nil {
			return nil, err
		}
	}
	// Make sure we have the mmap size.
	mmapSize, err := GetMmapSize(fd)
	if err != nil {
		return nil, err
	}
	// Make sure we have cpuid data.
	cpuid, err := DefaultCPUID(fd)
	if err != nil {
		return nil, err
	}
	// Get our list of available MSRs.
	msrs, err := AvailableMsrs(fd)
	if err != nil {
		return nil, err
	}
	// Create new VM.
	vmFd, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(IOctlCreateVirtualMachine), 0)
	if e != 0 {
		return nil, e
	}
	// Make sure this VM gets closed.
	// (Same thing is done for Vcpus).
	syscall.CloseOnExec(int(vmfd))
	// Prepare our VM object.
	vm := &VirtualMachine{
		Fd:       int(vmFd),
		CPUID:    cpuid,
		MSRs:     msrs,
		MmapSize: mmapSize,
	}
	return vm, nil
}

func (self *VirtualMachine) Dispose() error {
	for _, vcpu := range self.vcpus {
		vcpu.Dispose()
	}
	return syscall.Close(self.fd)
}

func (self *VirtualMachine) VCPUInfo() ([]VCPUInfo, error) {
	err := self.Pause(false)
	if err != nil {
		return nil, err
	}
	defer self.Unpause(false)
	vcpus := make([]VCPUInfo, 0, len(self.vcpus))
	for _, vcpu := range self.vcpus {
		vcpuinfo, err := NewVCPUInfo(vcpu)
		if err != nil {
			return nil, err
		}
		vcpus = append(vcpus, vcpuinfo)
	}
	return vcpus, nil
}

func (vm *VirtualMachine) Pause(manual bool) error {
	// Pause all vcpus.
	for i, vcpu := range vm.vcpus {
		err := vcpu.Pause(manual)
		if err != nil && err != AlreadyPaused {
			// Rollback.
			// NOTE: Start with the previous.
			for i -= 1; i >= 0; i -= 1 {
				vm.vcpus[i].Unpause(manual)
			}
			return err
		}
	}
	// Done.
	return nil
}

func (vm *VirtualMachine) Unpause(manual bool) error {
	// Unpause all vcpus.
	for i, vcpu := range vm.vcpus {
		err := vcpu.Unpause(manual)
		if err != nil && err != NotPaused {
			// Rollback.
			// NOTE: Start with the previous.
			for i -= 1; i >= 0; i -= 1 {
				vm.vcpus[i].Pause(manual)
			}
			return err
		}
	}
	// Done.
	return nil
}
