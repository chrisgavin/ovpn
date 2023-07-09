package vpn

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"os/user"
	"path"

	"github.com/loynoir/ExpandUser.go"
	"github.com/pkg/errors"
)

func (connection *Connection) Connect() error {
	_, err := os.Stat(connection.statusDirectory())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(connection.statusDirectory(), 0o600)
			if err != nil {
				return errors.Wrap(err, "failed to create status directory")
			}
		} else {
			return errors.Wrap(err, "failed to stat status directory")
		}
	}

	_, err = os.Stat(connection.pidPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Remove(connection.pidPath())
			if err != nil {
				return errors.Wrap(err, "failed to remove pid file")
			}
		} else {
			return errors.Wrap(err, "failed to stat pid file")
		}
	}

	temporaryDirectory, err := os.MkdirTemp("", "ovpn")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(temporaryDirectory)

	currentUser, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "failed to get current user")
	}
	currentGroup, err := user.LookupGroupId(currentUser.Gid)
	if err != nil {
		return errors.Wrap(err, "failed to get current group")
	}

	command := exec.Command("sudo", "openvpn")
	for _, configurationFile := range connection.ConfigurationFiles {
		resolved, err := ExpandUser.ExpandUser(configurationFile)
		if err != nil {
			return errors.Wrap(err, "failed to resolve configuration file")
		}
		command.Args = append(command.Args, "--config", resolved)
	}

	for _, configurationScript := range connection.ConfigurationScripts {
		resolved, err := ExpandUser.ExpandUser(configurationScript)
		if err != nil {
			return errors.Wrap(err, "failed to resolve configuration script")
		}
		digest := hex.EncodeToString(sha256.New().Sum([]byte(resolved)))
		configurationCommand := exec.Command(resolved)
		configurationCommand.Stderr = os.Stderr
		configuration, err := configurationCommand.Output()
		if err != nil {
			return errors.Wrap(err, "failed to run configuration script")
		}
		configurationFile := path.Join(temporaryDirectory, digest+".conf")
		err = os.WriteFile(configurationFile, configuration, 0o600)
		if err != nil {
			return errors.Wrap(err, "failed to write configuration file")
		}
		command.Args = append(command.Args, "--config", configurationFile)
	}

	if connection.AuthenticationScript != "" {
		resolved, err := ExpandUser.ExpandUser(connection.AuthenticationScript)
		if err != nil {
			return errors.Wrap(err, "failed to resolve authentication script")
		}
		authenticationCommand := exec.Command(resolved)
		authenticationCommand.Stderr = os.Stderr
		authentication, err := authenticationCommand.Output()
		if err != nil {
			return errors.Wrap(err, "failed to run authentication script")
		}
		authenticationFile := path.Join(temporaryDirectory, "_authentication.txt")
		err = os.WriteFile(authenticationFile, authentication, 0o600)
		if err != nil {
			return errors.Wrap(err, "failed to write authentication file")
		}
		command.Args = append(command.Args, "--auth-user-pass", authenticationFile)
	}

	command.Args = append(command.Args, "--log-append", connection.LogPath())
	command.Args = append(command.Args, "--management", connection.controlSocketPath(), "unix")
	command.Args = append(command.Args, "--writepid", connection.pidPath())
	command.Args = append(command.Args, "--management-client-user", currentUser.Username)
	command.Args = append(command.Args, "--management-client-group", currentGroup.Name)
	command.Args = append(command.Args, "--daemon")

	resolvedWorkingDirectory, err := ExpandUser.ExpandUser(connection.WorkingDirectory)
	if err != nil {
		return errors.Wrap(err, "failed to resolve working directory")
	}
	command.Dir = resolvedWorkingDirectory

	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return errors.Wrap(err, "failed to run openvpn")
	}

	return nil
}
