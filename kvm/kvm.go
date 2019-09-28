// +build linux
package kvm

import (
	"syscall"

	api "github.com/multiverse-os/portalgun/vm/kvm/api"
)

// https://github.com/kerpz/tomato/blob/60d3c74ab93d43d0e6a22f56f0876c0180c5c7d3/release/src-rt/linux/linux-2.6/include/linux/efi.h
type KVM struct {
	Fd              uintptr
	API             *API
	Path            string
	VirtualMachines []*VirtualMachine
}

func NewKVM() *KVM {
	return &KVM{
		API:             InitAPI(),
		Path:            "/dev/kvm",
		VirtualMachines: make([]*VirtualMachine),
	}
}

func (self *KVM) API(method) (int, error) {
	return self.API.call(self.Fd, method, 0)
}

func (self *KVM) ConnectKVM() (int, error) {
	return syscall.Open(self.Path, syscall.O_RDWR|syscall.O_CLOEXEC, 0)
}

func (self *KVM) APIVersion() (int, error) {
	version, _, e := self.API.Endpoint["api_version"].call(self.Fd())
	if version != 12 || e != 0 {
		return nil, errors.New("[ioctl] failed to execute kvm syscall 'api version'")
	} else {
		return version, nil
	}
}

func (self *KVM) ioctlVCPUMMapSize() (int, error) {
	return osIOctl.ioctl(self.Fd(), self.API.Endpoint["vcpu"], 0)
}

func (self *KVM) APIVersion() (int, error) {
	version, _, e := self.API.Endpoint["api_version"].call(self.Fd())
	if version != 12 || e != 0 {
		return nil, errors.New("[ioctl] failed to execute kvm syscall 'api version'")
	} else {
		return version, nil
	}
}

func (self *KVM) ioctlVCPUMMapSize() (int, error) {
	return osIOctl.ioctl(self.Fd(), self.API.Endpoint["vcpu"], 0)
}
