package cmd

import (
	"fmt"
	"time"

	"github.com/chrisgavin/ovpn/internal/tail"
	"github.com/chrisgavin/ovpn/internal/vpn"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type DisconnectCommand struct {
	*RootCommand
}

func registerDisconnectCommand(rootCommand *RootCommand) {
	command := &DisconnectCommand{
		RootCommand: rootCommand,
	}
	disconnectCommand := &cobra.Command{
		Use:           "disconnect",
		Short:         "Disconnect a connection.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration, err := vpn.LoadConfiguration()
			if err != nil {
				return err
			}

			color.Blue("Disconnecting %s...", args[0])

			connection := configuration.Connection(args[0])
			if connection == nil {
				color.Red("%s not found.", args[0])
				return SilentErr
			}

			status, err := connection.Status()
			if err != nil {
				return err
			}
			if status.State == vpn.StateDisconnected {
				color.Red("%s is already disconnected.", connection.Name)
				return SilentErr
			}

			tailFile, err := tail.TailFile(connection.LogPath(), tail.Options{NewLinesOnly: true})
			if err != nil {
				return err
			}
			defer tailFile.Cleanup()

			err = connection.Disconnect()
			if err != nil {
				return err
			}

			running := true
			for running {
				select {
				case line := <-tailFile.Lines:
					fmt.Println(line.Text)
				default:
					running, err = connection.Running()
					if err != nil {
						command.logger.Sugar().Errorf("%+v", err)
					}
					if running {
						time.Sleep(100 * time.Millisecond)
					}
				}
			}

			tailFile.StopAtEOF()
			for line := range tailFile.Lines {
				fmt.Println(line.Text)
			}

			color.Green("%s is now disconnected.", connection.Name)

			return nil
		},
	}
	command.root.AddCommand(disconnectCommand)
}
