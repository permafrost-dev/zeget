package app

import (
	"fmt"
)

func (app *Application) write(format string, args ...any) (n int, err error) {
	return fmt.Fprintf(app.Output, format, args...)
}

func (app *Application) writeLine(format string, args ...any) (n int, err error) {
	return app.write(format+"\n", args...)
}

func (app *Application) writeError(format string, args ...any) {
	fmt.Fprintf(app.Outputs.Stderr, format, args...)
}

func (app *Application) writeErrorLine(format string, args ...any) {
	app.writeError(format+"\n", args...)
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
