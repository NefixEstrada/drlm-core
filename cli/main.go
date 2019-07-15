package cli

import "github.com/brainupdaters/drlm-core/transport/grpc"

// Main is the main function of DRLM Core
func Main() {
	grpc.Serve()
}
