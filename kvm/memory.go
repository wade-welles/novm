// +build linux
package kvm

// IOCTL calls.
//const int IoctlSetUserMemoryRegion = KVM_SET_USER_MEMORY_REGION;
//const int IoctlTranslate = KVM_TRANSLATE;
//
//// IOCTL flags.
//const int IoctlFlagMemLogDirtyPages = KVM_MEM_LOG_DIRTY_PAGES;

import (
	"syscall"
	"unsafe"
)

func (self *VirtualMachine) MapUserMemory(start Pointer, size uint64, mmap []byte) error {
	// See NOTE above about read-only memory.
	// As we will not support it for the moment,
	// we do not expose it through the interface.
	// Leveraging that feature will likely require
	// a small amount of re-architecting in any case.
	var region struct_kvm_userspace_memory_region
	region.slot = uint32(self.MemoryRegion)
	region.flags = uint32(0)
	region.guest_phys_addr = uint64(start)
	region.memory_size = uint64(size)
	region.userspace_addr = uint64(uintptr(unsafe.Pointer(&mmap[0])))
	// Execute the ioctl.
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(self.Fd), uintptr(IOctlSetUserMemoryRegion), uintptr(unsafe.Pointer(&region)))
	if e != 0 {
		return e
	}
	// We're set, bump our slot.
	self.MemoryRegion += 1
	return nil
}

func (self *VirtualMachine) MapReservedMemory(start Pointer, size uint64) error {
	// Nothing to do.
	return nil
}

func (self *VCPU) Translate(vaddr VirtualAddress) (Pointer, bool, bool, bool, error) {
	// Perform the translation.
	var translation struct_kvm_translation
	translation.linearAddress = uint64(vaddr)
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(self.Fd), uintptr(IOctlTranslate), uintptr(unsafe.Pointer(&translation)))
	if e != 0 {
		return Pointer(0), false, false, false, e
	}
	paddr := Pointer(translation.physicalAddress)
	valid := translation.valid != uint8(0)
	writeable := translation.writeable != uint8(0)
	usermode := translation.valid != uint8(0)

	return paddr, valid, writeable, usermode, nil
}
