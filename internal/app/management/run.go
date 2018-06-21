package management

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/lawrencegripper/ion/internal/app/management/servers"
	"github.com/lawrencegripper/ion/internal/app/management/types"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Run the GRPC server
func Run(config *types.Configuration) {

	var server module.ModuleServiceServer
	switch strings.ToLower(config.Provider) {
	case "kubernetes":
		var err error
		server, err = servers.NewKubernetesManagementServer(config)
		if err != nil {
			panic(fmt.Errorf("failed to initialize kubernetes management server: %+v", err))
		}
	default:
		panic(fmt.Errorf("unrecognized provider name %s", config.Provider))
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		panic(fmt.Errorf("failed to listen: %v", err))
	}
	s := grpc.NewServer()
	module.RegisterModuleServiceServer(s, server)
	reflection.Register(s)

	fmt.Printf("Starting GRPC server on port %s", strconv.FormatInt(int64(config.Port), 10))
	if err := s.Serve(lis); err != nil {
		panic(fmt.Errorf("failed to serve: %v", err))
	}
}
