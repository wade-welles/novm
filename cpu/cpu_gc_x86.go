// +build 386 amd64 amd64p32
// +build !gccgo

package cpu

// cpuid is implemented in cpu_x86.s for gc compiler
// and in cpu_gccgo.c for gccgo.
func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)

// xgetbv with ecx = 0 is implemented in cpu_x86.s for gc compiler
// and in cpu_gccgo.c for gccgo.
func xgetbv() (eax, edx uint32)
