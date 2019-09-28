package kvm

//const int IoctlGetSupportedCPUID = KVM_GET_SUPPORTED_CPUID;
//const int IoctlSetCPUID = KVM_SET_CPUID2;
//const int IoctlGetCPUID = KVM_GET_CPUID2;

import (
	"syscall"
	"unsafe"
)

type CPUID struct {
	Function uint32 `json:"function"`
	Index    uint32 `json:"index"`
	Flags    uint32 `json:"flags"`

	EAX uint32
	EBX uint32
	ECX uint32
	EDX uint32
}

func SupportedCPUID(fd int) ([]CPUID, error) {

	// Initialize our cpuid data.
	cpuidData := make([]byte, PageSize, PageSize)
	cpuids := make([]CPUID, 0, 0)
	C.cpuid_init(unsafe.Pointer(&cpuidData[0]), PageSize)

	for {
		_, _, e := syscall.Syscall(
			syscall.SYS_IOCTL,
			uintptr(fd),
			uintptr(C.IoctlGetSupportedCPUID),
			uintptr(unsafe.Pointer(&cpuidData[0])))

		if e == syscall.ENOMEM {
			// The nent field will now have been
			// adjusted, and we can run it again.
			continue
		} else if e != 0 {
			return nil, e
		}

		// We're good!
		break
	}

	// Extract each function.
	for i := 0; ; i += 1 {
		// Is there a valid function?
		var function C.__u32
		var index C.__u32
		var flags C.__u32
		var eax C.__u32
		var ebx C.__u32
		var ecx C.__u32
		var edx C.__u32

		e := C.cpuid_get(
			unsafe.Pointer(&cpuidData[0]),
			C.int(i),
			&function,
			&index,
			&flags,
			&eax,
			&ebx,
			&ecx,
			&edx)

		// Any left?
		if e != 0 {
			break
		}

		// Add this MSR.
		cpuids = append(cpuids, CPUID{
			Function: uint32(function),
			Index:    uint32(index),
			Flags:    uint32(flags),
			EAX:      uint32(eax),
			EBX:      uint32(ebx),
			ECX:      uint32(ecx),
			EDX:      uint32(edx)})
	}

	// Done.
	return cpuids, nil
}

func nativeCPUID(function uint32) CPUID {

	var eax C.__u32
	var ebx C.__u32
	var ecx C.__u32
	var edx C.__u32

	// Query our native function.
	C.cpuid_native(C.__u32(function), &eax, &ebx, &ecx, &edx)

	// Transform.
	return CPUID{
		Function: function,
		EAX:      uint32(eax),
		EBX:      uint32(ebx),
		ECX:      uint32(ecx),
		EDX:      uint32(edx)}
}

func defaultCPUID(fd int) ([]CPUID, error) {

	// Get the supported cpuids.
	cpuids, err := supportedCPUID(fd)
	if err != nil {
		return nil, err
	}

	// Change the vendor & feature bits.
	result := make([]CPUID, 0, len(cpuids))
	for _, cpuid := range cpuids {

		if cpuid.Function == 0 {
			// Tweak our vendor.
			native_cpuid := nativeCPUID(cpuid.Function)
			cpuid.EBX = native_cpuid.EBX
			cpuid.ECX = native_cpuid.ECX
			cpuid.EDX = native_cpuid.EDX

		} else if cpuid.Function == 1 {
			// Tweak our model & APIC status.
			native_cpuid := nativeCPUID(cpuid.Function)
			cpuid.EAX = native_cpuid.EAX
			cpuid.EDX |= (1 << 9)

		} else if cpuid.Function == 0x80000001 {
			// Mask our NX support.
			// FIXME: This seems to cause the system
			// to freeze up during boot. I'm not sure
			// why NX support would do that, but it's
			// a mystery that should be solved soon.
			cpuid.EDX &= ^uint32(1 << 20)
		}

		result = append(result, cpuid)
	}

	return result, nil
}

func (vcpu *VCPU) SetCPUID(cpuids []CPUID) error {

	// Initialize our cpuid data.
	cpuidData := make([]byte, PageSize, PageSize)
	for i, cpuid := range cpuids {
		e := C.cpuid_set(
			unsafe.Pointer(&cpuidData[0]),
			C.int(PageSize),
			C.int(i),
			C.__u32(cpuid.Function),
			C.__u32(cpuid.Index),
			C.__u32(cpuid.Flags),
			C.__u32(cpuid.EAX),
			C.__u32(cpuid.EBX),
			C.__u32(cpuid.ECX),
			C.__u32(cpuid.EDX))
		if e != 0 {
			return syscall.Errno(e)
		}
	}

	// Set our vcpuid.
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(vcpu.Fd),
		uintptr(C.IoctlSetCPUID),
		uintptr(unsafe.Pointer(&cpuidData[0])))
	if e != 0 {
		return e
	}

	// We're good.
	vcpu.cpuid = cpuids
	return nil
}

func (vcpu *VCPU) GetCPUID() ([]CPUID, error) {
	// This is super annoying. If we are querying
	// capabilities, then it expects us to give the
	// size of the buffer we pass, and it will say ENOMEM
	// if have too many entries. On the other hand, if
	// we are calling GET_CPUID2, then it expects us to
	// pass zero and will only adjust nent after it gives
	// us E2BIG as a result. How dumb is that?
	// Anyways, all this lead to just caching the thing.
	return vcpu.cpuid, nil
}
