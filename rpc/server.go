package control

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"sync"
	"syscall"

	kvm "github.com/multiverse-os/portalgun/vm/kvm"
	linux "github.com/multiverse-os/portalgun/vm/linux"
	vmrpc "github.com/multiverse-os/portalgun/vm/rpc"
)

type Control struct {
	ControlFd int
	RealInit  bool
	Proxy     Proxy
	RPC       *RPC

	// Our bound client (to the in-guest agent).
	// NOTE: We have this setup as a lazy function
	// because the guest may take some small amount of
	// time before it's actually ready to process RPC
	// requests. We don't want this to interfere with
	// our ability to process our host-side requests.
	// TODO This should be INSIDE a client object
	ClientResult chan error
	ClientErr    error
	ClientOnce   sync.Once
	ClientCodec  rpc.ClientCodec
	Client       *rpc.Client
}

func (control *Control) handle(connFd int, server *RPC.Server) {

	controlFile := os.NewFile(uintptr(connFd), "control")
	defer controlFile.Close()

	// Read single header.
	// Our header is exactly 9 characters, and we
	// expect the last character to be a newline.
	// This is a simple plaintext protocol.
	headerBuffer := make([]byte, 9, 9)
	n, err := controlFile.Read(headerBuffer)
	if n != 9 || header_buf[8] != '\n' {
		if err != nil {
			controlFile.Write([]byte(err.Error()))
		} else {
			controlFile.Write([]byte("invalid header"))
		}
		return
	}
	header := string(headerBuffer)
	// We read a special header before diving into RPC
	// mode. This is because for the novmrun case, we turn
	// the socket into a stream of input/output events.
	// These are simply JSON serialized versions of the
	// events for the guest RPC interface.

	if header == "PORTAL RUN\n" {
		decoder := NewDecoder(controlFile)
		encoder := NewEncoder(controlFile)

		var start vmrpc.StartCommand
		err := decoder.Decode(&start)
		if err != nil {
			// Poorly encoded command.
			encoder.Encode(err.Error())
			return
		}
		// Grab our client.
		client, err := control.Ready()
		if err != nil {
			encoder.Encode(err.Error())
			return
		}
		// Call start.
		result := vmrpc.StartResult{}
		err = client.Call("Server.Start", &start, &result)
		if err != nil {
			encoder.Encode(err.Error())
			return
		}
		// Save our pid.
		pid := result.Pid
		inputs := make(chan error)
		outputs := make(chan error)
		exitcode := make(chan int)
		// This indicates we're okay.
		encoder.Encode(nil)
		// Wait for the process to exit.
		go func() {
			wait := vmrpc.WaitCommand{
				Pid: pid,
			}
			var waitResult vmrpc.WaitResult
			err := client.Call("Server.Wait", &wait, &waitResult)
			if err != nil {
				exitcode <- 1
			} else {
				exitCode <- waitResult.ExitCode
			}
		}()
		// Read from stdout & stderr.
		go func() {
			read := vmrpc.ReadCommand{
				Pid: pid,
				N:   4096,
			}
			var readResult vmrpc.ReadResult
			for {
				err := client.Call("Server.Read", &read, &readResult)
				if err != nil {
					inputs <- err
					return
				}
				err = encoder.Encode(readResult.Data)
				if err != nil {
					inputs <- err
					return
				}
			}
		}()
		// Write to stdin.
		go func() {
			write := vmrpc.WriteCommand{
				Pid: pid,
			}
			var writeResult vmrpc.WriteResult
			for {
				err := decoder.Decode(&write.Data)
				if err != nil {
					outputs <- err
					return
				}
				err = client.Call("Server.Write", &write, &writeResult)
				if err != nil {
					outputs <- err
					return
				}
			}
		}()

		// Wait till exit.
		status := <-exitCode
		encoder.Encode(status)
		// Wait till EOF.
		<-inputs
		// Send a notice and close the socket.
		encoder.Encode(nil)
	} else if header == "PORTAL RPC\n" {

		// Run as JSON RPC connection.
		codec := jsonrpc.NewServerCodec(controlFile)
		server.ServeCodec(codec)
	}
}

func (control *Control) Serve() {
	// Bind our rpc server.
	server := rpc.NewServer()
	server.Register(control.RPC)
	for {
		// Accept clients.
		nfd, _, err := syscall.Accept(control.ControlFd)
		if err == nil {
			go control.handle(nfd, server)
		}
	}
}

func NewControl(fd int, init bool, model *Model, vm *kvm.VirtualMachine, tracer *linux.Tracer, proxy Proxy, isLoad bool) (*Control, error) {
	if fd < 0 {
		return nil, InvalidControlSocketErr
	}

	// Create our control object.
	control := new(Control)
	control.controlFd = controlFd
	control.RealInit = init
	control.Proxy = proxy
	control.RPC = NewRPC(model, vm, tracer)

	// Start our barrier.
	control.clientResult = make(chan error, 1)
	if isLoad {
		go control.init()
	} else {
		// Already synchronized.
		control.clientResult <- nil
	}
	return control, nil
}
