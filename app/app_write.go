package app

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/lipgloss"
)

var checkMarkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1bef52")).Bold(true)
var filenameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#efe51b"))
var fadedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fffff"))

func (app *Application) Write(format string, args ...any) (n int, err error) {
	temp := []interface{}{}
	var style *lipgloss.Style = nil

	for _, arg := range args {
		if reflect.TypeOf(arg).Kind() == reflect.TypeOf(filenameStyle).Kind() {
			s := arg.(lipgloss.Style)
			style = &s
			continue
		}
		temp = append(temp, arg)
	}

	if style != nil {
		format = style.Render(format)
	}

	return fmt.Fprintf(app.Output, format, temp...)
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

func (app *Application) WriteCheck(newLine bool) {
	if newLine {
		app.WriteLine(checkMarkStyle.Render("✔ "))
	} else {
		app.Write(checkMarkStyle.Render("✔ "))
	}
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
