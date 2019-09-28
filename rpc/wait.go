package rpc

type WaitCommand struct {
	Pid int `json:"pid"`
}

type WaitResult struct {
	// The exit code.
	// (If > 0 then this event is an exit event).
	ExitCode int `json:"exitcode"`
}

func (server *Server) Wait(wait *WaitCommand, result *WaitResult) error {
	process := server.lookup(wait.Pid)
	if process == nil {
		result.ExitCode = -1
		return nil
	}
	process.wait()
	result.ExitCode = process.ExitCode
	return nil
}
