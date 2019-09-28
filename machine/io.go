package machine

import (
	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

//
// I/O events & operations --
//
// All I/O events (PIO & MMIO) are constrained to one
// simple interface. This simply makes writing devices
// a bit easier as you the read/write functions that
// must be implemented in each case are identical.
//
// This is a decision that could be revisited.

type IOEvent interface {
	Size() uint

	GetData() uint64
	SetData(val uint64)

	IsWrite() bool
}

type IOOperations interface {
	Read(offset uint64, size uint) (uint64, error)
	Write(offset uint64, size uint, value uint64) error
}

//
// I/O queues --
//
// I/O requests are serviced by a single go-routine,
// which pulls requests from a channel, performs the
// read/write as necessary and sends the result back
// on a requested channel.
//
// This structure was selected in order to allow all
// devices to operate without any locks and allowing
// their internal operation to be concurrent with the
// rest of the system.

type IORequest struct {
	event  IOEvent
	offset uint64
	result chan error
}

type IOQueue chan IORequest

func (queue IOQueue) Submit(event IOEvent, offset uint64) error {

	// Send the request to the device.
	req := IORequest{event, offset, make(chan error)}
	queue <- req

	// Pull the result when it's done.
	return <-req.result
}

//
// I/O Handler --
//
// A handler represents a device instance, combined
// with a set of operations (typically for a single address).
// Effectively, this is the unit of concurrency and would
// represent a single port for a single device.

type IOHandler struct {
	Device

	start      kvm.Pointer
	operations IOOperations
	queue      IOQueue
}

func NewIOHandler(
	device Device,
	start kvm.Pointer,
	operations IOOperations) *IOHandler {

	io := &IOHandler{
		Device:     device,
		start:      start,
		operations: operations,
		queue:      make(IOQueue),
	}

	// Start the handler.
	go io.Run()

	return io
}

func normalize(val uint64, size uint) uint64 {
	switch size {
	case 1:
		return val & 0xff
	case 2:
		return val & 0xffff
	case 4:
		return val & 0xffffffff
	}
	return val
}

func (io *IOHandler) Run() {

	for {
		// Pull first request.
		req := <-io.queue
		size := req.event.Size()

		// Note that we are running.
		// NOTE: This is a bit awkward. Theoretically,
		// we could actually be processing an exit from
		// a vcpu that is in the middle of an operation.
		// However, I chose to handle this case in kvm_run,
		// as we don't consider a VCPU to be paused until
		// it's event is completely processed. From the
		// device perspective -- nothing related to this
		// event has yet touched the device, so it's okay
		// to acquire it at this point and continue. If
		// this device is paused, then the VCPU will be
		// unpausable (therefore the normal practice will
		// be to pause all VCPUs, then pause all devices).
		io.Device.Acquire()

		// Perform the operation.
		if req.event.IsWrite() {
			val := normalize(req.event.GetData(), size)

			// Debug?
			io.Debug(
				"write %x @ %x [size: %d]",
				val,
				io.start.After(req.offset),
				size)

			err := io.operations.Write(req.offset, size, val)
			req.result <- err

		} else {
			val, err := io.operations.Read(req.offset, size)
			val = normalize(val, size)
			if err == nil {
				req.event.SetData(val)
			}

			req.result <- err

			// Debug?
			io.Debug(
				"read %x @ %x [size: %d]",
				val,
				io.start.After(req.offset),
				size)
		}

		// We've finished, we could now pause.
		io.Device.Release()
	}
}
