package doublecheck

import (
	"encoding/json"
	"io"
)

type JSONFormatter struct {
	w io.Writer
}

func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{w: w}
}

func (jf *JSONFormatter) Format(result *CheckResult) error {
	buf, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	_, err = jf.w.Write(buf)
	return err
}
