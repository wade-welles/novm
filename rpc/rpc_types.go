package control

import (
	machine "github.com/multiverse-os/portalgun/vm"
	kvm "github.com/multiverse-os/portalgun/vm/kvm"
	linux "github.com/multiverse-os/portalgun/vm/linux"
)

// The Noop --
// Many of our operations do not require
// a specific parameter or a specific return.
type Nop struct{}

// Rpc --
// This is basic state provided to the
// Rpc interface. All Rpc functions have
// access to this state (but nothing else).
// TODO: Lets switch to REST API, not ssure exaclty
// how this RPC is implemented
type RPC struct {
	// Our device model.
	Model *machine.Model
	// Our underlying Vm object.
	VM *kvm.VirtualMachine
	// Our tracer.
	Tracer *linux.Tracer
}

func NewRpc(model *machine.Model, vm *kvm.VirtualMachine, tracer *linux.Tracer) *RPC {
	return &RPC{
		Model:  model,
		VM:     vm,
		Tracer: tracer,
	}
}
