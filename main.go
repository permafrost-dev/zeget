package main

import (
	"github.com/permafrost-dev/eget/app"
)

func main() {
	appl := app.NewApplication(nil)
	result := appl.Run()

	appl.Output.Write([]byte(result.Msg + "\n"))
}
