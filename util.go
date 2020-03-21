package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
)

func round(f float64) float64 {
	// +1.5 in order to return 1 at least
	return math.Ceil(f+1.5) - 1
}

func renderReportAsJSON(o *OptimizerOutput, outStream io.Writer) error {
	bytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(outStream, string(bytes))
	return nil
}
