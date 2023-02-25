package cmd

import (
	"github.com/mgeri/udptunneler/cmd/client"
	"github.com/mgeri/udptunneler/cmd/dump"
	"github.com/mgeri/udptunneler/cmd/ping"
	"github.com/mgeri/udptunneler/cmd/server"
	"github.com/mgeri/udptunneler/pkg/version"
	"github.com/spf13/cobra"
)

var (
	udptunneler = &cobra.Command{
		Use:           "udptunneler",
		Short:         "udptunneler – command-line tool to tunnel UDP multicast traffic thought TCP",
		Long:          ``,
		Version:       version.String(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	return udptunneler.Execute()
}

func init() {
	// Allow to load some default config from file (unused for now)
	cobra.OnInitialize(initConfig)

	// Add subcommands here
	udptunneler.AddCommand(server.Cmd)
	udptunneler.AddCommand(client.Cmd)
	udptunneler.AddCommand(ping.Cmd)
	udptunneler.AddCommand(dump.Cmd)
}

func initConfig() {

}
