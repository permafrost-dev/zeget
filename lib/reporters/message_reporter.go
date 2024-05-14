package reporters

import (
	"fmt"
	"io"

	"github.com/permafrost-dev/zeget/lib/utilities"
)

type MessageReporter struct {
	Output       io.Writer
	FormatString string
	Arguments    []interface{}
	Reporter
}

func (m *MessageReporter) Report(input ...interface{}) error {
	if len(input) == 0 && len(m.FormatString) == 0 {
		return nil
	}

	args := make([]interface{}, 0)
	args = append(args, m.Arguments...)
	args = append(args, input...)
	fmtStr := "› " + utilities.SetIf(m.FormatString == "", m.FormatString, "%v\n")

	_, err := fmt.Fprintf(m.Output, fmtStr, args...)

	return err
}

func NewMessageReporter(output io.Writer, formatString string, arguments ...interface{}) *MessageReporter {
	return &MessageReporter{
		Output:       output,
		FormatString: formatString,
		Arguments:    arguments,
	}
}
