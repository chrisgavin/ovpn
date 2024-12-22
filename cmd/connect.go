package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/chrisgavin/ovpn/internal/vpn"
	"github.com/fatih/color"
	"github.com/nxadm/tail"
	"github.com/spf13/cobra"
)

type ConnectCommand struct {
	*RootCommand
}

func registerConnectCommand(rootCommand *RootCommand) {
	command := &ConnectCommand{
		RootCommand: rootCommand,
	}
	connectCommand := &cobra.Command{
		Use:           "connect",
		Short:         "Connect to a connection.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration, err := vpn.LoadConfiguration()
			if err != nil {
				return err
			}

			color.Blue("Connecting to %s...", args[0])

			connection := configuration.Connection(args[0])
			if connection == nil {
				color.Red("%s not found.", args[0])
				return SilentErr
			}

			status, err := connection.Status()
			if err != nil {
				return err
			}
			if status.State == vpn.StateConnected {
				color.Red("%s is already connected.", connection.Name)
				return SilentErr
			}

			tailFile, err := tail.TailFile(connection.LogPath(), tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}})
			if err != nil {
				return err
			}
			defer tailFile.Cleanup()

			err = connection.Connect()
			if err != nil {
				return err
			}

			started := false
			for !started {
				started, err = connection.Started()
				if err != nil {
					return err
				}
			}

			finished := false
			for !finished {
				select {
				case line := <-tailFile.Lines:
					if line != nil {
						fmt.Println(line.Text)
					}
				default:
					running, err := connection.Running()
					if err != nil {
						return err
					}
					if !running {
						finished = true
					}
					status, err := connection.Status()
					if err != nil {
						return err
					}
					if status.State == vpn.StateConnected {
						finished = true
					}
					if !finished {
						time.Sleep(100 * time.Millisecond)
					}
				}
			}

			tailFile.StopAtEOF()
			for line := range tailFile.Lines {
				fmt.Println(line.Text)
			}

			status, err = connection.Status()
			if err != nil {
				return err
			}

			if status.State != vpn.StateConnected {
				color.Red("%s failed to connect.", connection.Name)
				return SilentErr
			}
			color.Green("%s is now connected.", connection.Name)

			return nil
		},
	}
	command.root.AddCommand(connectCommand)
}
