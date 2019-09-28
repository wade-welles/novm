package machine

import (
	"syscall"

	kvm "github.com/multiverse-os/portalgun/vm/kvm"
)

//
// Commands.
//
const (
	VirtioBlockTIn       = 0
	VirtioBlockTOut      = 1
	VirtioBlockTFlush    = 4
	VirtioBlockTFlushOut = 5
	VirtioBlockTBarrier  = 0x80000000
)

//
// Status values.
const (
	VirtioBlockSOk          = 0
	VirtioBlockSIoErr       = 1
	VirtioBlockSUnsupported = 2
)

type VirtioBlockDevice struct {
	*VirtioDevice

	// The device.
	Dev string `json:"dev"`

	// The backing file.
	Fd int `json:"fd"`
}

func (device *VirtioBlockDevice) processRequests(
	vchannel *VirtioChannel) error {

	for buf := range vchannel.incoming {

		header := &Ram{buf.Map(0, 16)}

		// Legit?
		if header.Size() < 16 {
			vchannel.outgoing <- buf
			continue
		}

		// Request offset.
		sector := header.Get64(8)
		offset := int64(512 * sector)

		// What are we doing?
		cmd_type := header.Get32(0)

		// Our status byte.
		status := &Ram{buf.Map(buf.Length()-1, 1)}

		switch int(cmd_type) {
		case VirtioBlockTIn:
			_, err := buf.PRead(device.Fd, offset, 16, buf.Length()-17)
			if err != nil {
				device.Debug(
					"read err [%x,%x] -> %s",
					offset,
					int(offset)+buf.Length()-18,
					err.Error())
				status.Set8(0, VirtioBlockSIoErr)
			} else {
				device.Debug(
					"read ok [%x,%x]",
					offset,
					int(offset)+buf.Length()-18)
				status.Set8(0, VirtioBlockSOk)
			}
			break

		case VirtioBlockTOut:
			_, err := buf.PWrite(device.Fd, offset, 16, buf.Length()-17)
			if err != nil {
				device.Debug(
					"write err [%x,%x] -> %s",
					offset,
					int(offset)+buf.Length()-18,
					err.Error())
				status.Set8(0, VirtioBlockSIoErr)
			} else {
				device.Debug(
					"write ok [%x,%x]",
					offset,
					int(offset)+buf.Length()-18)
				status.Set8(0, VirtioBlockSOk)
			}
			break

		default:
			device.Debug("unknown command '%d'?", cmd_type)
			status.Set8(0, VirtioBlockSUnsupported)
			break
		}

		// Done.
		vchannel.outgoing <- buf
	}

	return nil
}

func NewVirtioMmioBlock(info *DeviceInfo) (Device, error) {
	device, err := NewMmioVirtioDevice(info, VirtioTypeBlock)
	device.Channels[0] = NewVirtioChannel(0, 256)
	return &VirtioBlockDevice{VirtioDevice: device}, err
}

func NewVirtioPciBlock(info *DeviceInfo) (Device, error) {
	device, err := NewPciVirtioDevice(info, PciClassStorage, VirtioTypeBlock, 16)
	device.Channels[0] = NewVirtioChannel(1, 256)
	return &VirtioBlockDevice{VirtioDevice: device}, err
}

func (block *VirtioBlockDevice) Attach(vm *kvm.VirtualMachine, model *Model) error {
	err := block.VirtioDevice.Attach(vm, model)
	if err != nil {
		return err
	}

	// Setup our config space.
	var stat syscall.Stat_t
	err = syscall.Fstat(block.Fd, &stat)
	if err != nil {
		return err
	}
	block.Config.GrowTo(24)
	block.Config.Set64(0, uint64(stat.Size)/512) // Total # of blocks.
	block.Config.Set32(8, 512)                   // Max segment size.
	block.Config.Set32(12, 1024)                 // Max # of segments per req.
	block.Config.Set16(20, uint16(stat.Blksize))

	// Start our network process.
	go block.processRequests(block.Channels[0])

	return nil
}
