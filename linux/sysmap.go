package linux

import (
	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

type SystemMap interface {
	Lookup(addr kvm.VirtualAddress) (string, uint64)
}

type Convention struct {
	instruction kvm.Register
	arguments   []kvm.Register
	rvalue      kvm.Register
	stack       kvm.Register
}
