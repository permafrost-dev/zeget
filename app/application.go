package app

import (
	"io"
	"os"
)

type Application struct {
	output io.Writer
	Opts   Flags
}

func createApplicationOutputWriter(quiet bool) io.Writer {
	// when --quiet is passed, send non-essential output to io.Discard
	if quiet {
		return io.Discard
	}

	return os.Stderr
}

func NewApplication(options Flags) *Application {
	return &Application{
		Opts:   options,
		output: createApplicationOutputWriter(false),
	}
}

func (a *Application) Run() error {
	return nil
}
