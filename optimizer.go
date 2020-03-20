package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ecs"
	"math"
	"time"
)

type Optimizer struct {
	cloudWatch *cloudwatch.CloudWatch
	ecs        *ecs.ECS
	ecsCluster string
	ecsService string
	currentDef map[string]int64
	percentage float64
}

func NewOptimizer(region string, cluster string, service string, profile string) *Optimizer {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:                  aws.Config{Region: aws.String(region)},
		Profile:                 profile,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
	}))

	return &Optimizer{
		cloudWatch: cloudwatch.New(sess),
		ecs:        ecs.New(sess),
		ecsCluster: cluster,
		ecsService: service,
		currentDef: make(map[string]int64, 3),
		percentage: 80,
	}
}

func (o *Optimizer) Run() error {
	if err := o.loadCurrentDefinition(); err != nil {
		return err
	}

	cpuAvgMetric, err := o.getAvgOfMetricStatistics("CPUUtilization")
	if err != nil {
		return err
	}

	memoryAvgMetric, err := o.getAvgOfMetricStatistics("MemoryUtilization")
	if err != nil {
		return err
	}

	fmt.Println("Proposed CPU Units", round((cpuAvgMetric/100)*float64(o.currentDef["cpu"])*(100/o.percentage)))
	fmt.Println("Proposed Memory", round((memoryAvgMetric/100)*float64(o.currentDef["memory"])*(100/o.percentage)))
	return nil
}

func (o *Optimizer) getAvgOfMetricStatistics(metricName string) (float64, error) {
	resp, err := o.cloudWatch.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		EndTime:    aws.Time(time.Now()),
		StartTime:  aws.Time(time.Now().Add(time.Duration(24*30) * time.Hour * -1)),
		MetricName: aws.String(metricName),
		Namespace:  aws.String("AWS/ECS"),
		Period:     aws.Int64(60 * 60 * 15),
		Statistics: []*string{
			aws.String(cloudwatch.StatisticAverage),
		},
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("ClusterName"),
				Value: aws.String(o.ecsCluster),
			},
			{
				Name:  aws.String("ServiceName"),
				Value: aws.String(o.ecsService),
			},
		},
		Unit: aws.String(cloudwatch.StandardUnitPercent),
	})
	if err != nil {
		return 0, err
	}

	return aws.Float64Value(resp.Datapoints[0].Average), nil
}

func (o *Optimizer) loadCurrentDefinition() error {
	ss, err := o.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String(o.ecsCluster),
		Services: []*string{aws.String(o.ecsService)},
	})
	if err != nil {
		return err
	}

	output, err := o.ecs.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: ss.Services[0].TaskDefinition,
	})
	if err != nil {
		return err
	}

	o.currentDef["cpu"] = aws.Int64Value(output.TaskDefinition.ContainerDefinitions[0].Cpu)
	o.currentDef["memory"] = aws.Int64Value(output.TaskDefinition.ContainerDefinitions[0].Memory)
	o.currentDef["reservedMemory"] = aws.Int64Value(output.TaskDefinition.ContainerDefinitions[0].MemoryReservation)

	return nil
}

func round(f float64) float64 {
	// +1.5 in order to return 1 at least
	return math.Ceil(f+1.5) - 1
}
