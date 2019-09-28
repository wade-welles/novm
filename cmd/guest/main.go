package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"syscall"

	protocol "github.com/multiverse-os/portalgun/vm/protocol"
	rpc "github.com/multiverse-os/portalgun/vm/rpc"
)

// The default control file.
var control = flag.String("control", "/dev/vport0p0", "control file")

// Should this always run a server.
var server_fd = flag.Int("serverfd", -1, "run RPC server")

func mount(fs string, location string) error {
	// Do we have the location?
	_, err := os.Stat(location)
	if err != nil {
		// Make sure it's a directory.
		err = os.Mkdir(location, 0755)
		if err != nil {
			return err
		}
	}
	// Try to mount it.
	cmd := exec.Command("/bin/mount", "-t", fs, fs, location)
	return cmd.Run()
}

func main() {
	var console *os.File

	// Parse flags.
	flag.Parse()

	if *server_fd == -1 {
		// Open the console.
		if f, err := os.OpenFile(*control, os.O_RDWR, 0); err != nil {
			log.Fatal("Problem opening console:", err)
		} else {
			console = f
		}
		// Make sure devpts is mounted.
		err := mount("devpts", "/dev/pts")
		if err != nil {
			log.Fatal(err)
		}

		// Notify novmm that we're ready.
		buffer := make([]byte, 1, 1)
		buffer[0] = protocol.PortalStatusOkay
		n, err := console.Write(buffer)
		if err != nil || n != 1 {
			log.Fatal(err)
		}

		// Read our response.
		n, err = console.Read(buffer)
		if n != 1 || err != nil {
			log.Fatal(protocol.UnknownCommand)
		}

		// Rerun to cleanup argv[0], or create a real init.
		new_args := make([]string, 0, len(os.Args)+1)
		new_args = append(new_args, "portal")
		new_args = append(new_args, "-serverfd", "0")
		new_args = append(new_args, os.Args[1:]...)

		switch buffer[0] {

		case protocol.PortalCommandRealInit:
			// Run our portal server in a new process.
			proc_attr := &syscall.ProcAttr{
				Dir:   "/",
				Env:   os.Environ(),
				Files: []uintptr{console.Fd(), 1, 2},
			}
			_, err := syscall.ForkExec(os.Args[0], new_args, proc_attr)
			if err != nil {
				log.Fatal(err)
			}

			// Exec our real init here in place.
			err = syscall.Exec("/sbin/init", []string{"init"}, os.Environ())
			log.Fatal(err)

		case protocol.PortalCommandFakeInit:
			// Since we don't have any init to setup basic
			// things, like our hostname we do some of that here.
			syscall.Sethostname([]byte("portal"))

		default:
			// What the heck is this?
			log.Fatal(protocol.UnknownCommand)
		}
	} else {
		// Open the defined fd.
		console = os.NewFile(uintptr(*server_fd), "console")
	}

	// Small victory.
	log.Printf("~~~ PORTAL ~~~")

	// Create our RPC server.
	rpc.Run(console)
}
