package control

//
// State-related rpcs.
func (rpc *RPC) State(nop *Nop, res *State) error {
	state, err := SaveState(rpc.VM, rpc.Model)
	if err != nil {
		return err
	}
	// Save our state.
	res.vCPUs = state.vCPUs
	res.Devices = state.Devices
	return err
}

func (rpc *RPC) Reload(in *Nop, out *Nop) error {
	// Pause the vm.
	// This is kept pausing for the entire reload().
	err := rpc.VM.Pause(false)
	if err != nil {
		return err
	}
	defer rpc.VM.Unpause(false)
	// Save a copy of the current state.
	state, err := SaveState(rpc.VM, rpc.Model)
	if err != nil {
		return err
	}
	// Reload all vcpus.
	for i, vCPUSpec := range state.vCPUs {
		err := rpc.VM.vCPUs()[i].Load(vCPUSpec)
		if err != nil {
			return err
		}
	}
	// Reload all device state.
	return rpc.model.Load(rpc.VM)
}
