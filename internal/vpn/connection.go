package vpn

import (
	"os"
	"path/filepath"

	"github.com/NordSecurity/gopenvpn/openvpn"
	"github.com/pkg/errors"
)

type Connection struct {
	Name                 string
	ConfigurationFiles   []string `yaml:"configuration_files"`
	ConfigurationScripts []string `yaml:"configuration_scripts"`
	AuthenticationScript string   `yaml:"authentication_script"`
	WorkingDirectory     string   `yaml:"working_directory"`
}

func statusDirectory() string {
	temporaryDirectory := os.TempDir()
	return filepath.Join(temporaryDirectory, "ovpn")
}

func (connection *Connection) statusDirectory() string {
	return filepath.Join(statusDirectory(), connection.Name)
}

func (connection *Connection) controlSocketPath() string {
	return filepath.Join(connection.statusDirectory(), "control.sock")
}

func (connection *Connection) LogPath() string {
	return filepath.Join(connection.statusDirectory(), "log.log")
}

func (connection *Connection) pidPath() string {
	return filepath.Join(connection.statusDirectory(), "pid.txt")
}

func (connection *Connection) managementConnection() (*openvpn.MgmtClient, error) {
	managementConnection, err := openvpn.Dial(connection.controlSocketPath(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "management connection failed")
	}
	return managementConnection, nil
}

func (connection *Connection) CreateStatusDirectory() error {
	_, err := os.Stat(connection.statusDirectory())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(connection.statusDirectory(), 0o700)
			if err != nil {
				return errors.Wrap(err, "failed to create status directory")
			}
		} else {
			return errors.Wrap(err, "failed to stat status directory")
		}
	}
	return nil
}
