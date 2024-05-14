package main

import (
	"github.com/permafrost-dev/zeget/app"
)

func main() {
	appl := app.NewApplication(nil)
	result := appl.Run()

	if result.Err != nil {
		appl.WriteErrorLine(result.Msg)
	}
}
