package root

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/lawrencegripper/ion/internal/pkg/common"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var cfgFile string
var managementEndpoint string
var tlsCerts common.TLSCerts
var timeoutSec int

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:               "ion",
	Short:             "ion cli lets you easily work with the ion system",
	PersistentPreRunE: Setup,
	RunE:              Root,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Root will run when no sub command is provided
func Root(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

// Setup runs before the commands to initialise any global data
func Setup(cmd *cobra.Command, args []string) error {
	return nil
}

// GetManagementConnection will return a new GRPC client connection
// to the management server. If the required TLS certificates have
// been provided to the CLI, mutual authentication will be used. If
// not then the connection will be insecure.
func GetManagementConnection() (*grpc.ClientConn, error) {

	var options []grpc.DialOption

	if tlsCerts.Available() {
		certificate, err := tls.LoadX509KeyPair(
			tlsCerts.CertFile,
			tlsCerts.KeyFile,
		)
		if err != nil {
			panic(fmt.Errorf("failed to read client certificate file '%s' and key file '%s': %+v", tlsCerts.CertFile, tlsCerts.KeyFile, err))
		}

		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(tlsCerts.CACertFile)
		if err != nil {
			panic(fmt.Errorf("failed to read client ca cert: %s", err))
		}
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			panic(fmt.Errorf("failed to append client certs"))
		}
		serverName := strings.Split(managementEndpoint, ":")[0]
		creds := credentials.NewTLS(&tls.Config{
			ServerName:   serverName,
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})
		options = append(options, grpc.WithTransportCredentials(creds))
		// options = append(options, grpc.WithWaitForHandshake()) - will force to validate TLS up front
	} else {
		options = append(options, grpc.WithInsecure())
	}

	options = append(options, grpc.WithBlock())
	options = append(options, grpc.WithTimeout(time.Duration(timeoutSec)*time.Second))

	fmt.Printf("connecting to GRPC server on %s\n", managementEndpoint)
	var err error
	cc, err := grpc.Dial(managementEndpoint, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server %s: %+v", managementEndpoint, err)
	}
	return cc, nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags shared by all subcommands
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ioncli.yaml)")
	RootCmd.PersistentFlags().StringVar(&managementEndpoint, "endpoint", "localhost:9000", "management server endpoint")
	RootCmd.PersistentFlags().IntVar(&timeoutSec, "timeout", 30, "timeout in seconds for cli to connect to management server")
	RootCmd.PersistentFlags().StringVar(&tlsCerts.CertFile, "certfile", "", "x509 PEM formatted client certificate")
	RootCmd.PersistentFlags().StringVar(&tlsCerts.KeyFile, "keyfile", "", "client private key for client certificate")
	RootCmd.PersistentFlags().StringVar(&tlsCerts.CACertFile, "cacertfile", "", "Root CA certificate file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".ioncli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ioncli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("using config file:", viper.ConfigFileUsed())
	}
}
