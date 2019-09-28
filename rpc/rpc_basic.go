package control

//
// High-level rpcs.
func (self *RPC) Pause(nopin *Nop, nopout *Nop) error   { return self.VM.Pause(true) }
func (self *RPC) Unpause(nopin *Nop, nopout *Nop) error { return self.VM.Unpause(true) }
