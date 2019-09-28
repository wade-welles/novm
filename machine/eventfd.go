package machine

import (
	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

type WriteIOEvent struct {
	size uint
	data uint64
}

func (event *WriteIOEvent) Size() uint {
	return event.size
}

func (event *WriteIOEvent) GetData() uint64 {
	return event.data
}

func (event *WriteIOEvent) SetData(val uint64) {
	// This really shouldn't happen.
	// Perhaps we should consider recording
	// this and raising an error later?
}

func (event *WriteIOEvent) IsWrite() bool {
	return true
}

func (cache *IOCache) save(
	vm *kvm.VirtualMachine,
	addr kvm.Pointer,
	handler *IOHandler,
	ioevent IOEvent,
	offset uint64) error {

	// Do we have sufficient hits?
	if cache.hits[addr] < 100 {
		return nil
	}

	// Bind an eventfd.
	// Note that we pass in the exactly address here,
	// not the address associated with the IOHandler.
	boundfd, err := vm.NewBoundEventFd(
		addr,
		ioevent.Size(),
		cache.is_pio,
		true,
		ioevent.GetData())
	if err != nil || boundfd == nil {
		return err
	}

	// Create a fake event.
	// This is because the real event will actually
	// reach into the vcpu registers to get the data.
	fake_event := &WriteIOEvent{ioevent.Size(), ioevent.GetData()}

	// Run our function.
	go func(ioevent IOEvent) {

		for {
			// Wait for the next event.
			_, err := boundfd.Wait()
			if err != nil {
				break
			}

			// Call our function.
			// We keep handling this device the same
			// way until it tells us to stop by returning
			// anything other than the SaveIO error.
			err = handler.queue.Submit(ioevent, offset)
			if err != SaveIO {
				break
			}
		}

		// Finished with the eventfd.
		boundfd.Close()

	}(fake_event)

	// Success.
	return nil
}
