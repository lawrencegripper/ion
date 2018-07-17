package management

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/lawrencegripper/ion/internal/app/management/servers"
	"github.com/lawrencegripper/ion/internal/app/management/types"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/lawrencegripper/ion/internal/pkg/management/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Run the GRPC server
func Run(config *types.Configuration) {

	var moduleServer module.ModuleServiceServer
	switch strings.ToLower(config.Provider) {
	case "kubernetes":
		var err error
		moduleServer, err = servers.NewKubernetesManagementServer(config)
		if err != nil {
			panic(fmt.Errorf("failed to initialize kubernetes management server: %+v", err))
		}
	default:
		panic(fmt.Errorf("unrecognized provider name %s", config.Provider))
	}

	traceServer, err := servers.NewTraceServer(config)
	if err != nil {
		panic(fmt.Errorf("failed to initialize the trace management server: %+v", err))
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		panic(fmt.Errorf("failed to listen: %v", err))
	}
	s := grpc.NewServer()

	module.RegisterModuleServiceServer(s, moduleServer)
	trace.RegisterTraceServiceServer(s, traceServer)

	reflection.Register(s)

	fmt.Printf("Starting GRPC server on port %s", strconv.FormatInt(int64(config.Port), 10))
	if err := s.Serve(lis); err != nil {
		panic(fmt.Errorf("failed to serve: %v", err))
	}
}
