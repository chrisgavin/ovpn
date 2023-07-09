package vpn

import (
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type State string

const (
	StateDisconnected State = "disconnected"
	StateConnecting   State = "connecting"
	StateConnected    State = "connected"
)

type Status struct {
	State         State
	Uptime        time.Duration
	LocalAddress  string
	RemoteAddress string
}

func (connection *Connection) Status() (*Status, error) {
	managementConnection, err := connection.managementConnection()
	if err != nil {
		if _, err := os.Stat(connection.controlSocketPath()); errors.Is(err, os.ErrNotExist) {
			return &Status{State: StateDisconnected}, nil
		}
		return nil, err
	}
	defer managementConnection.Close()

	stateData, err := managementConnection.LatestState()
	if err != nil {
		if _, err := os.Stat(connection.controlSocketPath()); errors.Is(err, os.ErrNotExist) {
			return &Status{State: StateDisconnected}, nil
		}
		return nil, errors.Wrap(err, "failed to get latest state")
	}

	status := &Status{State: StateConnecting}

	if stateData.NewState() == "CONNECTED" {
		status.State = StateConnected
	}

	timestamp, err := strconv.ParseInt(stateData.RawTimestamp(), 10, 64)
	if err != nil {
		return nil, err
	}
	status.Uptime = time.Since(time.Unix(timestamp, 0))
	status.LocalAddress = stateData.LocalTunnelAddr()
	status.RemoteAddress = stateData.RemoteAddr()

	return status, nil
}
