package reporters

import (
	"fmt"
	"io"

	"github.com/permafrost-dev/eget/lib/utilities"
)

type Message struct {
	Output       io.Writer
	FormatString string
	Arguments    []interface{}
	Reporter
}

func (m *Message) Report(input ...interface{}) error {
	if len(input) == 0 {
		return nil
	}

	args := make([]interface{}, 0)
	args = append(args, m.Arguments...)
	args = append(args, input...)
	fmtStr := "› " + utilities.SetIf(m.FormatString == "", m.FormatString, "%v\n")

	_, err := fmt.Fprintf(m.Output, fmtStr, args...)

	return err
}

func NewMessage(output io.Writer, formatString string, arguments ...interface{}) *Message {
	return &Message{
		Output:       output,
		FormatString: formatString,
		Arguments:    arguments,
	}
}
