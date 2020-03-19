package main

import (
	"fmt"
	"github.com/grezar/ecs-optimizer"
	flag "github.com/spf13/pflag"
	"os"
)

var Version = "0.1.0"

var (
	region  *string = flag.StringP("region", "r", "", "AWS region")
	cluster *string = flag.StringP("cluster", "c", "", "ECS cluster name")
	profile *string = flag.StringP("profile", "p", "", "A named profile. When you specify a profile, the settings and credentials are used to run this command")
)

func main() {
	os.Exit(_main())
}

func _main() int {
	flag.Parse()
	optimizer := optimizer.NewOptimizer(*region, *cluster, *profile)
	if err := optimizer.Run(); err != nil {
		fmt.Println("error: ", err)
		return 1
	}
	return 0
}
