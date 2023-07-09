package vpn

import "github.com/pkg/errors"

func (connection *Connection) Disconnect() error {
	managementConnection, err := connection.managementConnection()
	if err != nil {
		return err
	}
	defer managementConnection.Close()
	err = managementConnection.SendSignal("SIGTERM")
	if err != nil {
		return errors.Wrap(err, "failed to send signal")
	}
	return nil
}
