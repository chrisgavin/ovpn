package cmd

import (
	"time"

	"github.com/chrisgavin/ovpn/internal/vpn"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type StatusCommand struct {
	*RootCommand
}

var stateColors = map[vpn.State]color.Attribute{
	vpn.StateConnected:    color.FgGreen,
	vpn.StateConnecting:   color.FgYellow,
	vpn.StateDisconnected: color.FgRed,
}

func registerStatusCommand(rootCommand *RootCommand) {
	command := &StatusCommand{
		RootCommand: rootCommand,
	}
	statusCommand := &cobra.Command{
		Use:           "status",
		Short:         "Show the VPN client status.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration, err := vpn.LoadConfiguration()
			if err != nil {
				return err
			}

			if len(args) == 0 {
				for _, connection := range configuration.ConnectionsList() {
					status, err := connection.Status()
					if err != nil {
						return err
					}

					color.New(stateColors[status.State]).Printf("%s [%s]\n", connection.Name, status.State)
				}
			} else {
				connection := configuration.Connection(args[0])
				if connection == nil {
					color.Red("%s not found.", args[0])
					return SilentErr
				}

				status, err := connection.Status()
				if err != nil {
					return err
				}

				colorPrinter := color.New(stateColors[status.State])
				colorPrinter.Printf("Status: %s\n", status.State)
				if status.State != vpn.StateDisconnected {
					colorPrinter.Printf("Connection Uptime: %s\n", status.Uptime.Round(time.Second))
					colorPrinter.Printf("Local Address: %s\n", status.LocalAddress)
					colorPrinter.Printf("Remote Address: %s\n", status.RemoteAddress)
				}
			}

			return nil
		},
	}
	command.root.AddCommand(statusCommand)
}
