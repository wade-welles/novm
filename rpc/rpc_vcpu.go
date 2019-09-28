package control

import (
	"syscall"
)

//
// Low-level vcpu controls.
type VCPUSettings struct {
	Id     int  `json:"id"`
	Step   bool `json:"step"`
	Paused bool `json:"paused"`
}

func (rpc *RPC) VCPU(settings *VCPUSettings, nop *Nop) error {
	// A valid vcpu?
	vCPUs := rpc.VM.vCPUs()
	if settings.Id >= len(vCPUs) {
		return syscall.EINVAL
	}
	// Grab our specific vcpu.
	vCPU := vCPUs[settings.Id]
	// Ensure steping is as expected.
	err := vCPU.SetStepping(settings.Step)
	if err != nil {
		return err
	}
	// Ensure that the vcpu is paused/unpaused.
	if settings.Paused {
		err = vCPU.Pause(true)
	} else {
		err = vCPU.Unpause(true)
	}
	// Done.
	return err
}
