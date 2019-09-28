// +build linux
package kvm

// IOCTL calls.
//const int IOctlIOEventFd = KVM_IOEVENTFD;
// IOCTL flags.
//const int IOctlIOEventFdFlagPio = KVM_IOEVENTFD_FLAG_PIO;
//const int IOctlIOEventFdFlagDatamatch = KVM_IOEVENTFD_FLAG_DATAMATCH;
//const int IOctlIOEventFdFlagDeassign = KVM_IOEVENTFD_FLAG_DEASSIGN;

import (
	"syscall"
	"unsafe"
)

type BoundEventFd struct {
	// Our system eventfd.
	*EventFd
	// Our VM reference.
	*VirtualMachine

	// Address information.
	Address Pointer
	Size    uint
	isPIO   bool

	// Value information.
	hasValue bool
	value    uint64
}

func (vm *VirtualMachine) SetEventFd(eventfd *EventFd, paddr Pointer, size uint, isPIO bool, unbind bool, has_value bool, value uint64) error {

	var ioeventfd C.struct_kvm_ioeventfd
	ioeventfd.addr = C.__u64(paddr)
	ioeventfd.len = C.__u32(size)
	ioeventfd.fd = C.__s32(eventfd.Fd())
	ioeventfd.datamatch = C.__u64(value)

	if is_pio {
		ioeventfd.flags |= C.__u32(C.IOctlIOEventFdFlagPio)
	}
	if unbind {
		ioeventfd.flags |= C.__u32(C.IOctlIOEventFdFlagDeassign)
	}
	if has_value {
		ioeventfd.flags |= C.__u32(C.IOctlIOEventFdFlagDatamatch)
	}

	// Bind / unbind the eventfd.
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(vm.Fd), uintptr(C.IOctlIOEventFd), uintptr(unsafe.Pointer(&ioeventfd)))
	if e != 0 {
		return e
	}

	// Success.
	return nil
}

func (vm *VirtualMachine) NewBoundEventFd(paddr Pointer, size uint, is_pio bool, has_value bool, value uint64) (*BoundEventFd, error) {
	// Create our system eventfd.
	eventfd, err := NewEventFd()
	if err != nil {
		return nil, err
	}
	// Bind the eventfd.
	err = vm.SetEventFd(eventfd, paddr, size, is_pio, false, has_value, value)
	if err != nil {
		eventfd.Close()
		return nil, err
	}
	// Return our bound event.
	return &BoundEventFd{
		EventFd:        eventfd,
		VirtualMachine: vm,
		Pointer:        paddr,
		Size:           size,
		isPIO:          is_pio,
		hasValue:       has_value,
		Value:          value,
	}, nil
}

func (fd *BoundEventFd) Close() error {

	// Unbind the event.
	err := fd.VirtualMachine.SetEventFd(
		fd.EventFd,
		fd.paddr,
		fd.size,
		fd.is_pio,
		true,
		fd.has_value,
		fd.value)
	if err != nil {
		return err
	}

	// Close the eventfd.
	return fd.Close()
}

// Event server.
//
// This file was created in the hopes that I would
// be able to bolt on an event server to the internal
// network hub. Not so simple. That's all in the net
// namespace, and very much network-specific.
//
// So... for now, this will just use blocking system
// calls. It's relatively lightweight and we're not scaling
// to thousands of concurrent goroutines, just dozens.
//
// In the future, this is a great opportunity to improve
// the core archiecture (and eliminate a few system threads).
type EventFd struct {
	// Underlying FD.
	// NOTE: In reality we may want to serialize read/write
	// access to this fd as I'm fairly sure we will end up
	// with unexpected errors if this interface is used in
	// this way. However, we'll keep this as a simple runtime
	// adaptor and punt that complexity up to the next level.
	Fd int
}

func NewEventFd() (*EventFd, error) {
	// Create new eventfd.
	// NOTE: It's critical that it's non-blocking for the hub
	// integration below (otherwise it'll just end up blocking
	// in the Read() or Write() system call.
	// But given that we aren't using the hub, for now this is
	// just a regular blocking call. C'est la vie.
	fd, _, e := syscall.Syscall(syscall.SYS_EVENTFD, 0, uintptr(C.EfdCloExec), 0)
	if e != 0 {
		return nil, syscall.Errno(e)
	}
	eventFd := &EventFd{Fd: int(fd)}
	return eventFd, nil
}

func (self *EventFd) Close() error {
	return syscall.Close(self.Fd)
}

func (self *EventFd) Wait() (uint64, error) {
	for {
		var val uint64
		_, _, err := syscall.Syscall(syscall.SYS_READ, uintptr(self.Fd), uintptr(unsafe.Pointer(&val)), 8)
		if err != 0 {
			if err == syscall.EAGAIN || err == syscall.EINTR {
				continue
			}
			return 0, err
		}
		return val, nil
	}
	// Unreachable.
	return 0, nil
}

func (self *EventFd) Signal(val uint64) error {
	for {
		var val uint64
		_, _, err := syscall.Syscall(syscall.SYS_WRITE, uintptr(self.Fd), uintptr(unsafe.Pointer(&val)), 8)
		if err != 0 {
			if err == syscall.EAGAIN || err == syscall.EINTR {
				continue
			}
			return err
		}
		return nil
	}
	// Unreachable.
	return nil
}
