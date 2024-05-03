package app

import (
	"fmt"
	"io"
	"os"
)

func (a *Application) write(format string, args ...any) (n int, err error) {
	return fmt.Fprintf(a.Output, format, args...)
}

func (app *Application) writeLine(format string, args ...any) (n int, err error) {
	return app.write(format+"\n", args...)
}

func (app *Application) writeError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func (app *Application) writeErrorLine(format string, args ...any) {
	app.writeError(format+"\n", args...)
}

func (app *Application) initOutputWriter() {
	if app.Output == nil && !app.Opts.Quiet {
		app.Output = os.Stderr
	}

	if app.Output == nil && app.Opts.Quiet {
		app.Output = io.Discard
	}
}
