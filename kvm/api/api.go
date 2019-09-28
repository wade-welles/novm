package api

import (
	"fmt"

	unix "golang.org/x/sys/unix"
)

type api struct {
	Version     int
	HostVersion int
	Endpoint    map[string]method
}

type method struct {
	Pointer uintptr
}

func (self *method) IO(fd uintptr) (uintptr, error) {
	value, _, err := unix.Syscall(unix.SYS_IOCTL, fd, self.Pointer, 0)
	if err != 0 {
		return 0, fmt.Errorf("[syscall] failed to execute syscall:", err)
	} else {
		return value, nil
	}
}

func InitAPI() api {
	return api{
		Endpoint: map[string]method{
			"api_version":             method{uintptr(0xAE00)},
			"vm_create":               method{uintptr(0xAE01)},
			"vm_exec":                 method{uintptr(0xAE80)},
			"vcpu_create":             method{uintptr(0xAE41)},
			"vcpu_mmap_size":          method{uintptr(0xAE04)},
			"msr_index":               method{uintptr(0xC004AE02)},
			"msr_index_features":      method{uintptr(0xC004AE0A)},
			"check_extension":         method{uintptr(0xAE03)},
			"cpuid_supported":         method{uintptr(0xC008AE05)},
			"cpuid_emulated":          method{uintptr(0xC008AE09)},
			"user_memory_region":      method{uintptr(0x4020AE46)},
			"read_registers":          method{uintptr(0x8090AE81)},
			"write_registers":         method{uintptr(0x4090ae82)},
			"read_special_registers":  method{uintptr(0x8138AE83)},
			"write_special_registers": method{uintptr(0x4138AE84)},
			// s390 is an arch, like mips; maybe we break up sie from the arch and
			// piece them together since each arch will have this feature presumably
			"s390_sie": method{uintptr(0xAE06)},
		},
	}
}

// TODO: Hook the below standard object into the abovoe, so without doing
// anything but calling the method and maybe passing a FD you get an appropriate
// object back
//  Architectural interrupt line count, and the size of the bitmap needed
//  to hold them.
//#define KVM_NR_INTERRUPTS 256
//#define KVM_IRQ_BITMAP_SIZE_BYTES    ((KVM_NR_INTERRUPTS + 7) / 8)
//#define KVM_IRQ_BITMAP_SIZE(type)    (KVM_IRQ_BITMAP_SIZE_BYTES / sizeof(type))
///* for KVM_CREATE_MEMORY_REGION */
//struct kvm_memory_region {
//	__u32 slot;
//	__u32 flags;
//	__u64 guest_phys_addr;
//	__u64 memory_size; /* bytes */
//};
//
///* for kvm_memory_region::flags */
//#define KVM_MEM_LOG_DIRTY_PAGES  1UL
//
//struct kvm_memory_alias {
//	__u32 slot;  /* this has a different namespace than memory slots */
//	__u32 flags;
//	__u64 guest_phys_addr;
//	__u64 memory_size;
//	__u64 target_phys_addr;
//};
//
//enum kvm_exit_reason {
//	KVM_EXIT_UNKNOWN          = 0,
//	KVM_EXIT_EXCEPTION        = 1,
//	KVM_EXIT_IO               = 2,
//	KVM_EXIT_HYPERCALL        = 3,
//	KVM_EXIT_DEBUG            = 4,
//	KVM_EXIT_HLT              = 5,
//	KVM_EXIT_MMIO             = 6,
//	KVM_EXIT_IRQ_WINDOW_OPEN  = 7,
//	KVM_EXIT_SHUTDOWN         = 8,
//	KVM_EXIT_FAIL_ENTRY       = 9,
//	KVM_EXIT_INTR             = 10,
//};
//
///* for KVM_RUN, returned by mmap(vcpu_fd, offset=0) */
//struct kvm_run {
//
// Architectural interrupt line count, and the size of the bitmap needed
// to hold them.
//
// #define KVM_NR_INTERRUPTS 256
// #define KVM_IRQ_BITMAP_SIZE_BYTES    ((KVM_NR_INTERRUPTS + 7) / 8)
// #define KVM_IRQ_BITMAP_SIZE(type) (KVM_IRQ_BITMAP_SIZE_BYTES / sizeof(type))
