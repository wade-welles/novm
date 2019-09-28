package kvm

// I/O events & operations --
// All I/O events (PIO & MMIO) are constrained to one
// simple interface. This simply makes writing devices
// a bit easier as you the read/write functions that
// must be implemented in each case are identical.
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

type IOMap map[MemoryRegion]IOOperations
type IOHandlers map[MemoryRegion]*IOHandler

// I/O queues --
// I/O requests are serviced by a single go-routine,
// which pulls requests from a channel, performs the
// read/write as necessary and sends the result back
// on a requested channel.
// This structure was selected in order to allow all
// devices to operate without any locks and allowing
// their internal operation to be concurrent with the
// rest of the system.
type IORequest struct {
	Event  IOEvent
	Offset uint64
	Result chan error
}

type IOQueue chan IORequest

func (queue IOQueue) Submit(event IOEvent, offset uint64) error {

	// Send the request to the device.
	req := IORequest{event, offset, make(chan error)}
	queue <- req

	// Pull the result when it's done.
	return <-req.result
}

func (self *IOHandler) Run() {
	for {
		// Pull first request.
		req := <-self.queue
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
		self.Device.Acquire()
		// Perform the operation.
		if req.event.IsWrite() {
			val := normalize(req.event.GetData(), size)
			// Debug?
			self.Debug("write %x @ %x [size: %d]", val, self.start.After(req.offset), size)
			err := self.operations.Write(req.offset, size, val)
			req.result <- err
		} else {
			val, err := self.operations.Read(req.offset, size)
			val = normalize(val, size)
			if err == nil {
				req.event.SetData(val)
			}
			req.result <- err
			// Debug?
			self.Debug("read %x @ %x [size: %d]", val, self.start.After(req.offset), size)
		}
		// We've finished, we could now pause.
		self.Device.Release()
	}
}

// I/O Handler --
//
// A handler represents a device instance, combined
// with a set of operations (typically for a single address).
// Effectively, this is the unit of concurrency and would
// represent a single port for a single device.

type IOHandler struct {
	Device
	Start      Pointer
	Operations IOOperations
	Queue      IOQueue
}

func NewIOHandler(device Device, start Pointer, operations IOOperations) *IOHandler {
	io := &IOHandler{
		Device:     device,
		Start:      start,
		Operations: operations,
		Queue:      make(IOQueue),
	}
	// Start the handler.
	go io.Run()
	return io
}

// I/O cache --
// Our I/O cache stores paddr => handler mappings.
// When a device returns a SaveIO error, we actually try to
// setup an EventFD for that specific addr and value. This
// will avoid having to go through the cache every time. We
// do this only after accruing sufficient hits however, in
// order to avoid wasting eventfds on devices that only hit
// a few times (like an unused NIC, for example).
// See eventfd.go for the save() function where this is done.
type IOCache struct {
	// Our set of I/O handlers.
	Handlers []IOHandlers
	// Our I/O cache.
	Memory map[Pointer]*IOHandler
	// Our hits.
	Hits map[Pointer]uint64
	// Is this a Pio cache?
	IsPIO bool
}

func (self *IOCache) Lookup(addr Pointer) *IOHandler {
	handler, ok := self.Memory[addr]
	if ok {
		self.Hits[addr] += 1
		return handler
	}
	// See if we can find a matching device.
	for _, handlers := range self.Handlers {
		for portRegion, handler := range handlers {
			if portRegion.Contains(addr, 1) {
				self.Memory[addr] = handler
				self.Hits[addr] += 1
				return handler
			}
		}
	}
	// Nothing found.
	return nil
}

func NewIOCache(handlers []IOHandlers, isPIO bool) *IOCache {
	return &IOCache{
		Handlers: handlers,
		Memory:   make(map[Pointer]*IOHandler),
		Hits:     make(map[Pointer]uint64),
		IsPIO:    isPIO,
	}
}
