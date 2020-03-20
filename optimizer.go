package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"math"
	"time"
)

type Optimizer struct {
	cloudWatch cloudwatchiface.CloudWatchAPI
	ecs        *ecs.ECS
	config     *config
	currentDef map[string]int64
}

type OptimizerOutput struct {
	Cluster           string             `json:"cluster"`
	Service           string             `json:"service"`
	DesiredPercentage map[string]float64 `json:"desiredPercentage"`
	CurrentDef        map[string]int64   `json:"currentDef"`
	Utilization       map[string]float64 `json:"utilization"`
	Proposal          map[string]float64 `json:"proposal"`
}

func NewOptimizer(config *config) *Optimizer {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:                  aws.Config{Region: aws.String(config.region)},
		Profile:                 config.profile,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
	}))

	return &Optimizer{
		cloudWatch: cloudwatch.New(sess),
		ecs:        ecs.New(sess),
		config:     config,
		currentDef: make(map[string]int64, 3),
	}
}

func (o *Optimizer) Run() (*OptimizerOutput, error) {
	if err := o.loadCurrentDefinition(); err != nil {
		return nil, err
	}

	cpuAvgMetric, err := o.getAvgOfMetricStatistics("CPUUtilization")
	if err != nil {
		return nil, err
	}

	memoryAvgMetric, err := o.getAvgOfMetricStatistics("MemoryUtilization")
	if err != nil {
		return nil, err
	}

	return &OptimizerOutput{
		Cluster:           o.config.ecsCluster,
		Service:           o.config.ecsService,
		DesiredPercentage: o.config.desiredPercentage,
		CurrentDef:        o.currentDef,
		Utilization: map[string]float64{
			"cpu":    cpuAvgMetric,
			"memory": memoryAvgMetric,
		},
		Proposal: map[string]float64{
			"cpu":    o.calculateProposal(cpuAvgMetric, "cpu"),
			"memory": o.calculateProposal(memoryAvgMetric, "memory"),
		},
	}, nil
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
				Value: aws.String(o.config.ecsCluster),
			},
			{
				Name:  aws.String("ServiceName"),
				Value: aws.String(o.config.ecsService),
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
		Cluster:  aws.String(o.config.ecsCluster),
		Services: []*string{aws.String(o.config.ecsService)},
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

func (o *Optimizer) calculateProposal(avgUtilization float64, attr string) float64 {
	return round((avgUtilization / 100) * float64(o.currentDef[attr]) * (100 / o.config.desiredPercentage[attr]))
}
