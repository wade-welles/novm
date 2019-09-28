package linux

import (
	"debug/elf"
)

type ELF struct {
	*fd
}

func EntryPoint(path string) (uint64, bool, error) {
	e, err := elf.Open(path)
	if err != nil {
		return 0, false, err
	}
	var elf64 bool
	if e.Class == elf.ELFCLASS64 {
		elf64 = true
	} else {
		elf64 = false
	}
	return e.Entry, elf64, nil
}
