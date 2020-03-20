package main

import (
	"encoding/json"
	"fmt"
)

func renderReportAsJSON(o *OptimizerOutput) error {
	bytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}
