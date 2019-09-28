// +build linux
package kvm

import (
	"syscall"
)

type exitReason int

// These are exit reasons; not capabiltiy
//const (
//	KVM_EXIT_UNKNOWN exitReason = iota
//	KVM_EXIT_EXCEPTION
//	KVM_EXIT_IO
//	KVM_EXIT_HYPERCALL
//	KVM_EXIT_DEBUG
//	KVM_EXIT_HLT
//	KVM_EXIT_MMIO
//	KVM_EXIT_IRQ_WINDOW_OPEN
//	KVM_EXIT_SHUTDOWN
//	KVM_EXIT_FAIL_ENTRY
//	KVM_EXIT_INTR
//	KVM_EXIT_SET_TPR
//	KVM_EXIT_TPR_ACCESS
//	KVM_EXIT_S390_SIEIC
//	KVM_EXIT_S390_RESET
//	KVM_EXIT_DCR
//	KVM_EXIT_NMI
//)

type capabilityType int
type capability struct {
	Pointer uintptr
}

const (
	KVM_CAP_IRQCHIP                     capabilityType = 0
	KVM_CAP_HLT                                        = 1
	KVM_CAP_MMU_SHADOW_CACHE_CONTROL                   = 2
	KVM_CAP_USER_MEMORY                                = 3
	KVM_CAP_SET_TSS_ADDR                               = 4
	KVM_EXIT_HLT                                       = 5
	KVM_CAP_VAPIC                                      = 6
	KVM_CAP_EXT_CPUID                                  = 7
	KVM_CAP_CLOCKSOURCE                                = 8
	KVM_CAP_NR_VCPUS                                   = 9  /* returns recommended max vcpus per vm */
	KVM_CAP_NR_MEMSLOTS                                = 10 /* returns max memory slots per vm */
	KVM_CAP_PIT                                        = 11
	KVM_CAP_NOP_IO_DELAY                               = 12
	KVM_CAP_PV_MMU                                     = 13
	KVM_CAP_MP_STATE                                   = 14
	KVM_CAP_COALESCED_MMIO                             = 15
	KVM_CAP_SYNC_MMU                                   = 16 /* Changes to host mmap are reflected in guest */
	KVM_CAP_DEVICE_ASSIGNMENT                          = 17
	KVM_CAP_IOMMU                                      = 18
	KVM_EXIT_PAPR_HCALL                                = 19
	KVM_CAP_DEVICE_MSI                                 = 20
	KVM_CAP_DESTROY_MEMORY_REGION_WORKS                = 21
	KVM_CAP_USER_NMI                                   = 22
	KVM_CAP_SET_GUEST_DEBUG                            = 23
	KVM_CAP_REINJECT_CONTROL                           = 24
	KVM_CAP_IRQ_ROUTING                                = 25
	KVM_CAP_IRQ_INJECT_STATUS                          = 26
	KVM_CAP_DEVICE_DEASSIGNMENT                        = 27
	KVM_CAP_DEVICE_MSIX                                = 28
	KVM_CAP_ASSIGN_DEV_IRQ                             = 29
	KVM_CAP_JOIN_MEMORY_REGIONS_WORKS                  = 30
	KVM_CAP_MCE                                        = 31
	KVM_CAP_IRQFD                                      = 32
	KVM_CAP_PIT2                                       = 33
	KVM_CAP_SET_BOOT_CPU_ID                            = 34
	KVM_CAP_PIT_STATE2                                 = 35
	KVM_CAP_IOEVENTFD                                  = 36
	KVM_CAP_SET_IDENTITY_MAP_ADDR                      = 37
	KVM_CAP_XEN_HVM                                    = 38
	KVM_CAP_ADJUST_CLOCK                               = 39
	KVM_CAP_INTERNAL_ERROR_DATA                        = 40
	KVM_CAP_VCPU_EVENTS                                = 41
	KVM_CAP_S390_PSW                                   = 42
	KVM_CAP_PPC_SEGSTATE                               = 43
	KVM_CAP_HYPERV                                     = 44
	KVM_CAP_HYPERV_VAPIC                               = 45
	KVM_CAP_HYPERV_SPIN                                = 46
	KVM_CAP_PCI_SEGMENT                                = 47
	KVM_CAP_PPC_PAIRED_SINGLES                         = 48
	KVM_CAP_INTR_SHADOW                                = 49
	KVM_CAP_DEBUGREGS                                  = 50
	KVM_CAP_X86_ROBUST_SINGLESTEP                      = 51
	KVM_CAP_PPC_OSI                                    = 52
	KVM_CAP_PPC_UNSET_IRQ                              = 53
	KVM_CAP_ENABLE_CAP                                 = 54
	KVM_CAP_XSAVE                                      = 55
	KVM_CAP_XCRS                                       = 56
	KVM_CAP_PPC_GET_PVINFO                             = 57
	KVM_CAP_PPC_IRQ_LEVEL                              = 58
	KVM_CAP_ASYNC_PF                                   = 59
	KVM_CAP_TSC_CONTROL                                = 60
	KVM_CAP_GET_TSC_KHZ                                = 61
	KVM_CAP_PPC_BOOKE_SREGS                            = 62
	KVM_CAP_SPAPR_TCE                                  = 63
	KVM_CAP_PPC_SMT                                    = 64
	KVM_CAP_PPC_RMA                                    = 65
	KVM_CAP_MAX_VCPUS                                  = 66 /* returns max vcpus per vm */
	KVM_CAP_PPC_HIOR                                   = 67
	KVM_CAP_PPC_PAPR                                   = 68
	KVM_CAP_SW_TLB                                     = 69
	KVM_CAP_ONE_REG                                    = 70
	KVM_CAP_S390_GMAP                                  = 71
	KVM_CAP_TSC_DEADLINE_TIMER                         = 72
	KVM_CAP_S390_UCONTROL                              = 73
	KVM_CAP_SYNC_REGS                                  = 74
	KVM_CAP_PCI_2_3                                    = 75
)

// IOCTL calls.
const (
	IOCtlCheckExtension int = 3
)

func (self capability) Error() string {
	return "Missing capability: " + self.name
}

//
// Our required capabilities.
//
// Many of these are actually optional, but none
// of the plumbing has been done to gracefully fail
// when they are not available. For the time being
// development is focused on legacy-free environments,
// so we can split this out when it's necessary later.
//

func APIMethods() capability {
	return map[string]capability{
		"user_memory":           capability(uintptr(KVM_CAP_USER_MEMORY)),
		"set_identity_map_addr": capability(uintptr(KVM_CAP_SET_IDENTITY_MAP_ADDR)),
		"irqchip":               capability(uintptr(KVM_CAP_IRQCHIP)),
		"ioeventfd":             capability(uintptr(KVM_CAP_IOEVENTFD)),
		"irqfd":                 capability(uintptr(KVM_CAP_IRQFD)),
		"pit2":                  capability(uintptr(KVM_CAP_PIT2)),
		"pit2state2":            capability(uintptr(KVM_CAP_PIT2_STATE2)),
		"adjust_clock":          capability(uintptr(KVM_CAP_ADJUST_CLOCK)),
		"cpuid":                 capability(uintptr(KVM_CAP_EXT_CPUID)),
		"device_msi":            capability(uintptr(KVM_CAP_DEVICE_MSI)),
		"vcpu_events":           capability(uintptr(KVM_CAP_VCPU_EVENTS)),
		"xsave":                 capability(uintptr(KVM_CAP_XSAVE)),
		"xcrs":                  capability(uintptr(KVM_CAP_XCRS)),
	}
}

func CheckCapability(fd int, cap capability) error {
	r, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(IOCtlCheckExtension), cap.number)
	if r != 1 || e != 0 {
		return cap
	}

	return nil
}

func CheckCapabilities(fd int) error {
	// Check our extensions.
	for _, capSpec := range requiredCapabilities {
		err := CheckCapability(fd, capSpec)
		if err != nil {
			return err
		}
	}
	return nil
}
