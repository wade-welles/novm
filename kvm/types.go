package kvm

// Basic abstractions.
//type IRQ uint32

// Address types.
type VirtualAddress uint64
type Pointer uint64

func Align(addr uint64, alignment uint, up bool) uint64 {

	// Aligned already?
	if addr%uint64(alignment) == 0 {
		return addr
	}

	// Give the closest aligned address.
	addr = addr - (addr % uint64(alignment))

	if up {
		// Should we align up?
		return addr + uint64(alignment)
	}
	return addr
}

func (paddr Pointer) Align(alignment uint, up bool) Pointer {
	return Paddr(Align(uint64(paddr), alignment, up))
}

func (paddr Pointer) OffsetFrom(base Pointer) uint64 {
	return uint64(paddr) - uint64(base)
}

func (paddr Pointer) After(length uint64) Pointer {
	return Paddr(uint64(paddr) + uint64(length))
}
