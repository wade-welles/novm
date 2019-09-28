package vm

import (
	"log"
	"runtime"

	linux "github.com/multiverse-os/portalgun/vm/linux"
)

func Loop(vm *VirtualMachine, vcpu *VCPU, model *Model, tracer *linux.Tracer) error {
	// It's not really kosher to switch threads constantly when running a
	// KVM VCPU. So we simply lock this goroutine to a single system
	// thread. That way we know it won't be bouncing around.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	log.Printf("Vcpu[%d] running.", vcpu.Id)

	for {
		// Enter the guest.
		err := vcpu.Run()
		// Trace if requested.
		trace_err := tracer.Trace(vcpu, vcpu.IsStepping())
		if trace_err != nil {
			return trace_err
		}

		// No reason for exit?
		if err == nil {
			return ExitWithoutReason
		}

		// Handle the error.
		switch err.(type) {
		case *ExitPIO:
			err = model.HandlePio(vm, err.(*ExitPIO))

		case *ExitMMIO:
			err = model.HandleMmio(vm, err.(*ExitMMIO))

		case *ExitDebug:
			err = nil

		case *ExitShutdown:
			// Vcpu shutdown.
			return nil
		}

		// Error handling the exit.
		if err != nil {
			return err
		}
	}

	// Unreachable.
	return nil
}
