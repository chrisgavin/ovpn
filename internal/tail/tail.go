package tail

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const pollFrequency = 100 * time.Millisecond

type Line struct {
	Text string
	Err  error
}

type Tailer struct {
	handle   *os.File
	Lines    chan Line
	shutdown bool
}

type Options struct {
	NewLinesOnly bool
	Truncate     bool
}

func TailFile(path string, options Options) (*Tailer, error) {
	_, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, errors.Wrap(err, "failed to stat file")
	}
	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create file")
		}
		err = file.Close()
		if err != nil {
			return nil, errors.Wrap(err, "failed to close file")
		}
	} else if options.Truncate {
		err := os.Truncate(path, 0)
		if err != nil {
			return nil, errors.Wrap(err, "failed to truncate file")
		}
	}

	handle, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	if options.NewLinesOnly {
		_, err := handle.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, errors.Wrap(err, "failed to seek to end of file")
		}
	}
	tailer := &Tailer{
		handle: handle,
		Lines:  make(chan Line),
	}
	tailer.Start()
	return tailer, nil
}

func (t *Tailer) Start() {
	reader := bufio.NewReader(t.handle)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				t.Lines <- Line{Err: err}
				return
			}
			if errors.Is(err, io.EOF) {
				if t.shutdown {
					close(t.Lines)
					return
				}
				time.Sleep(pollFrequency)
				continue
			}
			line = strings.TrimSuffix(line, "\n")
			t.Lines <- Line{Text: line}
		}
	}()
}

func (t *Tailer) StopAtEOF() {
	t.shutdown = true
}

func (t *Tailer) Cleanup() error {
	return errors.Wrap(t.handle.Close(), "failed to close file")
}
