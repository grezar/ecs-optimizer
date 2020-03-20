package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
)

var Version = "0.1.0"

var (
	region                  *string  = flag.StringP("region", "r", "ap-northeast-1", "AWS region")
	cluster                 *string  = flag.StringP("cluster", "c", "", "ECS cluster name")
	service                 *string  = flag.StringP("service", "s", "", "ECS service name")
	profile                 *string  = flag.StringP("profile", "p", "", "A named profile. When you specify a profile, the settings and credentials are used to run this command")
	cpuDesiredPercentage    *float64 = flag.Float64P("cpu-desired-percentage", "C", 80, "Desired percentage of CPU utilization")
	memoryDesiredPercentage *float64 = flag.Float64P("memory-desired-percentage", "M", 80, "Desired percentage of memory utilization")
)

type config struct {
	region            string
	ecsCluster        string
	ecsService        string
	profile           string
	desiredPercentage map[string]float64
}

func main() {
	os.Exit(_main())
}

func _main() int {
	flag.Parse()
	optimizer := NewOptimizer(&config{
		region:     *region,
		ecsCluster: *cluster,
		ecsService: *service,
		profile:    *profile,
		desiredPercentage: map[string]float64{
			"cpu":    *cpuDesiredPercentage,
			"memory": *memoryDesiredPercentage,
		},
	})
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
