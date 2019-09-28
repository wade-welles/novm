package kvm

// #include "kvm_run.h"

import (
	"sync"
	"syscall"
)

// TODO: These funcitons can be merged they are almost exactly the same for
// weird reason

var SigVCPUInt = syscall.SIGUSR1

type RunInfo struct {
	Running     bool
	Pause       bool
	ThreadId    int64
	Lock        *sync.Mutex
	Pauses      int
	PauseEvent  *sync.Cond
	PesumeEvent *sync.Cond
}

func (self *VCPU) InitRunInfo() error {
	// Initialize our structure.
	self.RunInfo = RunInfo{
		Lock:    &sync.Mutex{},
		Running: true,
	}

	self.RunInfo.PauseEvent = sync.NewCond(self.RunInfo.Lock)
	self.RunInfo.ResumeEvent = sync.NewCond(self.RunInfo.Lock)

	e := syscall.Errno(kvm_run_init(self.Fd, &self.RunInfo.runLock))
	if e != 0 {
		return e
	}

	// Setup the lock.

	// We're okay.
	return nil
}

func (self *VCPU) Run() error {
	for {
		// Make sure our registers are flushed.
		// This will also refresh registers after we
		// execute but are interrupted (i.e. EINTR).
		err := self.flushAllRegs()
		if err != nil {
			return err
		}

		// Ensure we can run.
		//
		// For exact semantics, see Pause() and Unpause().
		// NOTE: By default, we are always "running". We are
		// only not running when we arrive at this point in
		// the pipeline and are waiting on the resume_event.
		//
		// This is because we want to ensure that our registers
		// have been flushed and all that devices are up-to-date
		// before we can declare a VCPU as "paused".
		self.RunInfo.lock.Lock()

		for self.RunInfo.is_paused || self.RunInfo.paused > 0 {
			// Note that we are not running,
			// See NOTE above about what this means.
			self.RunInfo.is_running = false

			// Send a notification that we are paused.
			self.RunInfo.pause_event.Broadcast()

			// Wait for a wakeup notification.
			self.RunInfo.resume_event.Wait()
		}

		self.RunInfo.is_running = true
		self.RunInfo.lock.Unlock()

		// Execute our run ioctl.
		rc := C.kvm_run(C.int(self.Fd), C.int(SigVCPUInt), &self.RunInfo.info)
		e := syscall.Errno(rc)

		if e == syscall.EINTR || e == syscall.EAGAIN {
			continue
		} else if e != 0 {
			return e
		} else {
			break
		}
	}

	return self.GetExitError()
}

func (self *VCPU) Pause(manual bool) error {
	// Acquire our runlock.
	self.RunInfo.lock.Lock()
	defer self.RunInfo.lock.Unlock()
	if manual {
		// Already paused?
		if self.RunInfo.is_paused {
			return AlreadyPaused
		}
		self.RunInfo.is_paused = true
	} else {
		// Bump our pause count.
		self.RunInfo.paused += 1
	}

	// Are we running? Need to interrupt.
	// We don't return from this function (even if there
	// are multiple callers) until we are sure that the VCPU
	// is actually paused, and all devices are up-to-date.
	if self.isRunning {
		// Only the first caller need interrupt.
		if manual || self.RunInfo.paused == 1 {
			e := kvm_run_interrupt(C.int(self.Dd), C.int(SigVCPUInt), &self.RunInfo.info)
			if e != 0 {
				return syscall.Errno(e)
			}
		}
		// Wait for the self to notify that it is paused.
		self.RunInfo.pause_event.Wait()
	}
	return nil
}

func (self *VCPU) Unpause(manual bool) error {
	// Acquire our runlock.
	self.RunInfo.lock.Lock()
	defer self.RunInfo.lock.Unlock()

	// Are we actually paused?
	// This was not a valid call.
	if manual {
		// Already unpaused?
		if !self.RunInfo.is_paused {
			return NotPaused
		}
		self.RunInfo.is_paused = false
	} else {
		if self.RunInfo.paused == 0 {
			return NotPaused
		}

		// Decrease our pause count.
		self.RunInfo.paused -= 1
	}

	// Are we still paused?
	if self.RunInfo.is_paused || self.RunInfo.paused > 0 {
		return nil
	}

	// Allow the self to resume.
	self.RunInfo.resume_event.Broadcast()

	return nil
}
