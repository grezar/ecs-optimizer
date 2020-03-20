package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
)

var Version = "0.1.0"

var (
	region  *string = flag.StringP("region", "r", "", "AWS region")
	cluster *string = flag.StringP("cluster", "c", "", "ECS cluster name")
	service *string = flag.StringP("service", "s", "", "ECS service name")
	profile *string = flag.StringP("profile", "p", "", "A named profile. When you specify a profile, the settings and credentials are used to run this command")
)

func main() {
	os.Exit(_main())
}

func _main() int {
	flag.Parse()
	optimizer := NewOptimizer(*region, *cluster, *service, *profile)
	output, err := optimizer.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return 1
	}

	if err := renderReportAsJSON(output); err != nil {
		fmt.Println("Error:", err)
		return 1
	}
	return 0
}
