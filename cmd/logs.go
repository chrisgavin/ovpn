package cmd

import (
	"fmt"

	"github.com/chrisgavin/ovpn/internal/tail"
	"github.com/chrisgavin/ovpn/internal/vpn"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type LogsCommand struct {
	*RootCommand
	follow bool
}

func registerLogsCommand(rootCommand *RootCommand) {
	command := &LogsCommand{
		RootCommand: rootCommand,
	}
	logsCommand := &cobra.Command{
		Use:           "logs",
		Short:         "Show the VPN client logs.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration, err := vpn.LoadConfiguration()
			if err != nil {
				return err
			}

			color.Blue("Showing logs for %s...", args[0])

			connection := configuration.Connection(args[0])
			if connection == nil {
				color.Red("Connection \"%s\" not found.", args[0])
				return SilentErr
			}

			tailFile, err := tail.TailFile(connection.LogPath(), tail.Options{})
			if err != nil {
				return err
			}
			defer tailFile.Cleanup()
			if !command.follow {
				tailFile.StopAtEOF()
			}

			for line := range tailFile.Lines {
				fmt.Println(line.Text)
			}

			return err
		},
	}
	logsCommand.Flags().BoolVar(&command.follow, "follow", false, "Follow the log file.")
	command.root.AddCommand(logsCommand)
}
