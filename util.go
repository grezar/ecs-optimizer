package main

import (
	"encoding/json"
	"fmt"
	"io"
)

func renderReportAsJSON(o *OptimizerOutput, outStream io.Writer) error {
	bytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(outStream, string(bytes))
	return nil
}
