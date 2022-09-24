package io

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-tty"
)

type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer

	TTYIn  io.Reader
	TTYOut io.Writer

	LogPath string
}

func New(logPath string) *Streams {
	return &Streams{
		In:      os.Stdin,
		Out:     os.Stdout,
		Err:     os.Stderr,
		LogPath: logPath,
		TTYIn:   nil,
		TTYOut:  nil,
	}
}

func (s *Streams) Wrap(fn func(*Streams) error) error {
	if s.LogPath != "" {
		// Since Git eats both stdout and stderr, we don't have a good way of
		// getting error information back from clients if things go wrong.
		// As a janky way to preserve error message, tee stderr to
		// a temp file.
		if f, err := os.Create(s.LogPath); err == nil {
			defer f.Close()
			s.Err = io.MultiWriter(s.Err, f)
		}
	}

	// A TTY may not be available in all environments (e.g. in CI), so only
	// set the input/output if we can actually open it.
	tty, err := tty.Open()
	if err == nil {
		defer tty.Close()
		s.TTYIn = tty.Input()
		s.TTYOut = tty.Output()
	} else {
		// If we can't connect to a TTY, fall back to stderr for output (which
		// will also log to file if GITSIGN_LOG is set).
		s.TTYOut = s.Err
	}

	// Log any panics to ttyout, since otherwise they will be lost to os.Stderr.
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(s.TTYOut, r)
		}
	}()

	if err := fn(s); err != nil {
		fmt.Fprintln(s.TTYOut, err)
		return err
	}
	return nil
}
