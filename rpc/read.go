package rpc

type ReadCommand struct {
	Pid int  `json:"pid"`
	N   uint `json:"n"`
}

type ReadResult struct {
	Data []byte `json:"data"`
}

func (server *Server) Read(
	read *ReadCommand,
	result *ReadResult) error {

	process := server.lookup(read.Pid)
	if process == nil {
		result.Data = []byte{}
		return nil
	}

	// Read available data.
	buffer := make([]byte, read.N, read.N)
	n, err := process.output.Read(buffer)
	if n > 0 {
		result.Data = buffer[:n]
	} else {
		result.Data = []byte{}
	}
	return err
}
