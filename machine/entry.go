package machine

import (
	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

func (model *Model) Handle(vm *kvm.VirtualMachine, cache *IoCache, handler *IoHandler, ioevent IoEvent, addr kvm.Pointer) error {
	if handler != nil {
		// Our offset from handler start.
		offset := addr.OffsetFrom(handler.start)
		// Submit our function.
		err := handler.queue.Submit(ioevent, offset)
		// Should we save this request?
		if ioevent.IsWrite() && err == SaveIO {
			err = cache.save(vm, addr, handler, ioevent, offset)
		}

		// Return to our vcpu.
		return err

	} else if !ioevent.IsWrite() {

		// Invalid reads return all 1's.
		switch ioevent.Size() {
		case 1:
			ioevent.SetData(0xff)
		case 2:
			ioevent.SetData(0xffff)
		case 4:
			ioevent.SetData(0xffffffff)
		case 8:
			ioevent.SetData(0xffffffffffffffff)
		}
	}

	return nil
}

func (model *Model) HandlePIO(
	vm *kvm.VirtualMachine,
	event *kvm.ExitPIO) error {

	handler := model.pio_cache.lookup(event.Port())
	ioevent := &PIOEvent{event}
	return model.Handle(vm, model.pio_cache, handler, ioevent, event.Port())
}

func (model *Model) HandleMMIO(
	vm *kvm.VirtualMachine,
	event *kvm.ExitMMIO) error {

	handler := model.mmio_cache.lookup(event.Addr())
	ioevent := &MMIOEvent{event}
	return model.Handle(vm, model.mmio_cache, handler, ioevent, event.Addr())
}
