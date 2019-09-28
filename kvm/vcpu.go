package kvm

import (
	"log"
)

type VCPUState uint32

const (
	Ready  VCPUState = 0
	User   VCPUState = 1
	Guest  VCPUState = 2
	Waiter VCPUState = 4
)

type VCPU struct {
	Id      int
	Fd      uintptr
	Machine *VirtualMachine
	//////////////////////
	//ring0.CPU
	// tid is the last set tid.
	TID uint64
	// switches is a count of world switches (informational only).
	Switches uint32
	// faults is a count of world faults (informational only).
	Faults uint32
	// state is the vCPU state.
	// This is a bitmask of the three fields (vCPU*) described above.
	State VCPUState
	// runData for this vCPU.
	//RunData *runData
	// machine associated with this vCPU.
	// active is the current addressSpace: this is set and read atomically,
	// it is used to elide unnecessary interrupts due to invalidations.
	//Active atomicAddressSpace
	// vCPUArchState is the architecture-specific state.
	ExitMessage string
}

type VCPUInfo struct {
	// Our optional id.
	// If this is not provided, we
	// assume that it is in order.
	Id *uint `json:"id"`
	// Full register state.
	Registers Registers `json:"registers"`
	// Optional multiprocessor state.
	MPState *MPState `json:"state"`
	// Our cpuid (not optional).
	CPUID []CPUID `json:"cpuid"`
	// Our LApic state.
	// This is optional, but is handled
	// within kvm_apic.go and not here.
	LAPIC LAPICState `json:"lapic"`
	// Our msrs (not optional).
	MSRs []MSR `json:"msrs"`
	// Our pending vcpu events.
	Events Events `json:"events"`
	// Optional FRU state.
	FPU *FPU `json:"fpu"`
	// Extended control registers.
	XCRs []XCR `json:"xcrs"`
	// Optional xsave state.
	XSave *XSave `json:"xsave"`
}

func (vm *VirtualMachine) CreateVCPUs(spec []VCPUInfo) ([]*VCPU, error) {
	vcpus := make([]*VCPU, 0, 0)
	// Load all vcpus.
	for index, info := range spec {
		// Sanitize vcpu ids.
		if info.Id == nil {
			newid := uint(index)
			info.Id = &newid
		}
		// Create a new vcpu.
		vcpu, err := vm.NewVCPU(*info.Id)
		if err != nil {
			return nil, err
		}
		// Load the state.
		err = vcpu.Load(info)
		if err != nil {
			return nil, err
		}
		// Good to go.
		vcpus = append(vcpus, vcpu)
	}
	// We've okay.
	return vcpus, nil
}

func (vcpu *VCPU) Load(info VCPUInfo) error {
	// Ensure the registers are loaded.
	log.Printf("vcpu[%d]: setting registers...", vcpu.Id)
	vcpu.SetRegisters(info.Registers)

	// Optional multiprocessing state.
	if info.MpState != nil {
		log.Printf("vcpu[%d]: setting vcpu state...", vcpu.Id)
		err := vcpu.SetMPState(*info.MPState)
		if err != nil {
			return err
		}
	}

	// Set our cpuid if we have one.
	if info.CPUID != nil {
		log.Printf("vcpu[%d]: setting cpuid...", vcpu.Id)
		err := vcpu.SetCPUID(info.CPUID)
		if err != nil {
			return err
		}
	}

	// Always load our Lapic.
	log.Printf("vcpu[%d]: setting apic state...", vcpu.Id)
	err := vcpu.SetLAPIC(info.LAPIC)
	if err != nil {
		return err
	}

	// Load MSRs if available.
	if info.MSRs != nil {
		log.Printf("vcpu[%d]: setting msrs...", vcpu.Id)
		err := vcpu.SetMSRs(info.MSRs)
		if err != nil {
			return err
		}
	}

	// Load events.
	log.Printf("vcpu[%d]: setting vcpu events...", vcpu.Id)
	err = vcpu.SetEvents(info.Events)
	if err != nil {
		return err
	}

	// Load fpu state if available.
	if info.FPU != nil {
		log.Printf("vcpu[%d]: setting fpu state...", vcpu.Id)
		err = vcpu.SetFPUState(*info.FPU)
		if err != nil {
			return err
		}
	}

	// Load Xcrs if available.
	if info.XCRs != nil {
		log.Printf("vcpu[%d]: setting xcrs...", vcpu.Id)
		err = vcpu.SetXCRs(info.XCRs)
		if err != nil {
			return err
		}
	}

	// Load xsave state if available.
	if info.XSave != nil {
		log.Printf("vcpu[%d]: setting xsave state...", vcpu.Id)
		err = vcpu.SetXSave(*info.XSave)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewVCPUInfo(vcpu *VCPU) (VCPUInfo, error) {
	err := vcpu.Pause(false)
	if err != nil {
		return VCPUInfo{}, err
	}
	defer vcpu.Unpause(false)

	registers, err := vcpu.GetRegisters()
	if err != nil {
		return VCPUInfo{}, err
	}

	mpstate, err := vcpu.GetMPState()
	if err != nil {
		return VCPUInfo{}, err
	}

	cpuid, err := vcpu.GetCPUID()
	if err != nil {
		return VCPUInfo{}, err
	}

	lapic, err := vcpu.GetLAPIC()
	if err != nil {
		return VCPUInfo{}, err
	}

	msrs, err := vcpu.GetMSRs()
	if err != nil {
		return VCPUInfo{}, err
	}

	events, err := vcpu.GetEvents()
	if err != nil {
		return VCPUInfo{}, err
	}

	fpu, err := vcpu.GetFPUState()
	if err != nil {
		return VCPUInfo{}, err
	}

	xcrs, err := vcpu.GetXCRs()
	if err != nil {
		return VCPUInfo{}, err
	}

	xsave, err := vcpu.GetXSave()
	if err != nil {
		return VCPUInfo{}, err
	}

	return VCPUInfo{
		Id:        &vcpu.Id,
		Registers: registers,
		MpState:   &mpstate,
		Cpuid:     cpuid,
		LApic:     lapic,
		Msrs:      msrs,
		Events:    events,
		Fpu:       &fpu,
		Xcrs:      xcrs,
		XSave:     &xsave,
	}, nil
}
