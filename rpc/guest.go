package control

import (
	"net/rpc"
	"net/rpc/jsonrpc"

	protocol "github.com/multiverse-os/portalgun/vm/protocol"
)

func (control *Control) init() {
	buffer := make([]byte, 1, 1)
	// Read our control byte back.
	n, err := control.proxy.Read(buffer)
	if n == 1 && err == nil {
		switch buffer[0] {
		case protocol.PortalStatusOkay:
			break
		case protocol.PortalStatusFailed:
			// Something went horribly wrong.
			control.client_res <- InternalGuestError
			return
		default:
			// This isn't good, who knows what happened?
			control.client_res <- protocol.UnknownStatus
			return
		}
	} else if err != nil {
		// An actual error.
		control.client_res <- err
		return
	}

	// Send our control byte to portal.
	// This essentially controls how the guest
	// will proceed during execution. If it is the
	// real init process, it will wait for run commands
	// and execute the given processes inside the VM.
	// If it is not the real init process, it will fork
	// and execute the real init before starting to
	// process any other RPC commands.
	if control.realInit {
		buffer[0] = protocol.PortalCommandRealInit
	} else {
		buffer[0] = protocol.PortalCommandFakeInit
	}
	n, err = control.proxy.Write(buffer)
	if n != 1 {
		// Can't send anything?
		control.client_res <- InternalGuestError
		return
	}

	// Looks like we're good.
	control.client_res <- nil
}

func (control *Control) barrier() {
	control.client_err = <-control.client_res
	control.client_codec = jsonrpc.NewClientCodec(control.proxy)
	control.client = rpc.NewClientWithCodec(control.client_codec)
}

func (control *Control) Ready() (*rpc.Client, error) {
	control.client_once.Do(control.barrier)
	return control.client, control.client_err
}
