package module

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/lawrencegripper/ion/cmd/ion/root"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//Client A shared GRPC module server client
var Client module.ModuleServiceClient
var managementEndpoint, certFile, keyFile, caCertFile string
var timeoutSec int

// moduleCmd represents the module command
var moduleCmd = &cobra.Command{
	Use:               "module",
	Short:             "execute commands to manage ion modules",
	PersistentPreRunE: Setup,
	Run:               Module,
}

// Module prints help
func Module(cmd *cobra.Command, args []string) {
	cmd.Help() // nolint: errcheck
}

// Setup is called before Run and is used to setup any
// persistent components needed by sub commands.
func Setup(cmd *cobra.Command, args []string) error {

	if cmd.HasSubCommands() {
		return nil
	}

	var options []grpc.DialOption

	if certFile != "" && keyFile != "" && caCertFile != "" {
		certificate, err := tls.LoadX509KeyPair(
			certFile,
			keyFile,
		)
		if err != nil {
			panic(fmt.Errorf("failed to read server certificate file '%s' and key file '%s': %+v", certFile, keyFile, err))
		}

		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			panic(fmt.Errorf("failed to read client ca cert: %s", err))
		}
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			panic(fmt.Errorf("failed to append client certs"))
		}
		creds := credentials.NewTLS(&tls.Config{
			ServerName:   strings.Split(managementEndpoint, ":")[0],
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	options = append(options, grpc.WithBlock())
	options = append(options, grpc.WithTimeout(time.Duration(timeoutSec)*time.Second))

	// Initialize a global GRPC connection to the management server
	conn, err := grpc.Dial(managementEndpoint, options...)
	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %+v", managementEndpoint, err)
	}
	Client = module.NewModuleServiceClient(conn)

	return nil
}

// Register adds to root command
func Register() {

	// Add module sub commands
	moduleCmd.AddCommand(createCmd)
	moduleCmd.AddCommand(deleteCmd)
	moduleCmd.AddCommand(listCmd)
	moduleCmd.AddCommand(getCmd)

	// Add module to root command
	root.RootCmd.AddCommand(moduleCmd)
}

func init() {

	// Local flags for the root command
	moduleCmd.PersistentFlags().StringVar(&managementEndpoint, "endpoint", "localhost:9000", "management server endpoint")
	moduleCmd.PersistentFlags().IntVar(&timeoutSec, "timeout", 30, "timeout in seconds for cli to connect to management server")
	moduleCmd.PersistentFlags().StringVar(&certFile, "certfile", "", "x509 PEM formatted client certificate")
	moduleCmd.PersistentFlags().StringVar(&keyFile, "keyfile", "", "x509 PEM formatted client key")
	moduleCmd.PersistentFlags().StringVar(&caCertFile, "cacertfile", "", "Root CA certificate file")
}
