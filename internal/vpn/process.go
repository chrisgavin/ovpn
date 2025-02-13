package vpn

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

func (connection *Connection) processID() (int, error) {
	processIDBytes, err := os.ReadFile(connection.pidPath())
	if err != nil {
		return 0, errors.Wrap(err, "failed to read pid file")
	}

	processID, err := strconv.Atoi(strings.TrimRight(string(processIDBytes), "\n"))
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse pid file")
	}

	return processID, nil
}

func (connection *Connection) Started() (bool, error) {
	_, err := os.Stat(connection.pidPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to stat pid file")
	}
	return true, nil
}

func (connection *Connection) Running() (bool, error) {
	processID, err := connection.processID()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	process, err := os.FindProcess(processID)
	if err != nil {
		return false, errors.Wrap(err, "failed to find process")
	}

	err = process.Signal(syscall.Signal(0))
	if err == syscall.ESRCH || err == os.ErrProcessDone {
		return false, nil
	}

	return true, nil
}
