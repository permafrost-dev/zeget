package app

import (
	"fmt"
)

func (app *Application) Write(format string, args ...any) (n int, err error) {
	return fmt.Fprintf(app.Output, format, args...)
}

func (app *Application) WriteLine(format string, args ...any) (n int, err error) {
	return app.Write(format+"\n", args...)
}

func (app *Application) WriteVerbose(format string, args ...any) (n int, err error) {
	if !app.Opts.Verbose {
		return 0, nil
	}

	return fmt.Fprintf(app.Output, format, args...)
}

func (app *Application) WriteVerboseLine(format string, args ...any) (n int, err error) {
	if !app.Opts.Verbose {
		return 0, nil
	}

	return app.Write(format+"\n", args...)
}

func (app *Application) WriteError(format string, args ...any) {
	fmt.Fprintf(app.Outputs.Stderr, format, args...)
}

func (app *Application) WriteErrorLine(format string, args ...any) {
	app.WriteError(format+"\n", args...)
}

func (app *Application) initOutputs() {
	if app.Output != nil {
		return
	}

	app.Output = app.Outputs.Stderr

	if app.Opts.Quiet {
		app.Output = app.Outputs.Discard
	}
}
