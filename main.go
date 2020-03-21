package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"os"
)

var Version = "0.1.0"

var (
	version                 *bool    = flag.BoolP("version", "v", false, "Print version")
	region                  *string  = flag.StringP("region", "r", "ap-northeast-1", "AWS region")
	cluster                 *string  = flag.StringP("cluster", "c", "", "ECS cluster name")
	service                 *string  = flag.StringP("service", "s", "", "ECS service name")
	profile                 *string  = flag.StringP("profile", "p", "", "A named profile. When you specify a profile, the settings and credentials are used to run this command")
	cpuDesiredPercentage    *float64 = flag.Float64P("cpu-desired-percentage", "C", 80, "Desired percentage of CPU utilization")
	memoryDesiredPercentage *float64 = flag.Float64P("memory-desired-percentage", "M", 80, "Desired percentage of memory utilization")
)

type config struct {
	outStream, errStream io.Writer
	region               string
	ecsCluster           string
	ecsService           string
	profile              string
	desiredPercentage    map[string]float64
}

func main() {
	os.Exit(_main())
}

func _main() int {
	flag.Parse()
	if *version {
		fmt.Println("ecs-optimizer", Version)
		return 0
	}
	config := newConfig()
	optimizer := NewOptimizer(config)
	output, err := optimizer.Run()
	if err != nil {
		fmt.Fprintln(config.errStream, "Error:", err)
		return 1
	}

	if err := renderReportAsJSON(output, config.outStream); err != nil {
		fmt.Fprintln(config.errStream, "Error:", err)
		return 1
	}
	return 0
}

func newConfig() *config {
	return &config{
		outStream:  os.Stdout,
		errStream:  os.Stderr,
		region:     *region,
		ecsCluster: *cluster,
		ecsService: *service,
		profile:    *profile,
		desiredPercentage: map[string]float64{
			"cpu":    *cpuDesiredPercentage,
			"memory": *memoryDesiredPercentage,
		},
	}
}
