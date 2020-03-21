package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"testing"
)

const (
	CPUUtilizationAvgValue    = 49.53529105870167
	MemoryUtilizationAvgValue = 102.3319085080778
)

type mockedECSiface struct {
	ecsiface.ECSAPI
	DescribeServicesResp       ecs.DescribeServicesOutput
	DescribeTaskDefinitionResp ecs.DescribeTaskDefinitionOutput
}

type mockedCloudwatchiface struct {
	cloudwatchiface.CloudWatchAPI
}

func (m mockedECSiface) DescribeServices(input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	return &m.DescribeServicesResp, nil
}

func (m mockedECSiface) DescribeTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	return &m.DescribeTaskDefinitionResp, nil
}

func (m mockedCloudwatchiface) GetMetricStatistics(input *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
	var v float64
	if aws.StringValue(input.MetricName) == "CPUUtilization" {
		v = CPUUtilizationAvgValue
	} else {
		v = MemoryUtilizationAvgValue
	}

	return &cloudwatch.GetMetricStatisticsOutput{
		Datapoints: []*cloudwatch.Datapoint{
			{
				Average: aws.Float64(v),
			},
		},
	}, nil
}

func TestGetAvgOfMetricStatistics(t *testing.T) {
	c := &config{
		region:     "ap-northeast-1",
		ecsCluster: "test-cluster",
		ecsService: "test-service",
	}
	optimizer := NewOptimizer(c)
	optimizer.cloudWatch = mockedCloudwatchiface{}

	expected := []struct {
		metricName string
		average    float64
	}{
		{
			"CPUUtilization",
			CPUUtilizationAvgValue,
		},
		{
			"MemoryUtilization",
			MemoryUtilizationAvgValue,
		},
	}

	for _, e := range expected {
		actual, err := optimizer.getAvgOfMetricStatistics(e.metricName)
		if err != nil {
			t.Errorf("Got %v", err)
		}

		if actual != e.average {
			t.Errorf("Expected %f, got %f", e.average, actual)
		}
	}
}

type ContainerDefinition struct {
	Cpu               int64
	Memory            int64
	MemoryReservation int64
}

func TestLoadCurrentDefinition(t *testing.T) {
	var c = &config{
		region:     "ap-northeast-1",
		ecsCluster: "test-cluster",
		ecsService: "test-service",
	}
	optimizer := NewOptimizer(c)

	cases := []struct {
		DescribeServicesResp       ecs.DescribeServicesOutput
		DescribeTaskDefinitionResp ecs.DescribeTaskDefinitionOutput
		Expected                   []ContainerDefinition
	}{
		{
			DescribeServicesResp: ecs.DescribeServicesOutput{
				Services: []*ecs.Service{
					{
						TaskDefinition: aws.String("arn:aws:ecs:ap-northeast-1:000000000001:task-definition/test-task-definition:1"),
					},
				},
			},
			DescribeTaskDefinitionResp: ecs.DescribeTaskDefinitionOutput{
				TaskDefinition: &ecs.TaskDefinition{
					ContainerDefinitions: []*ecs.ContainerDefinition{
						{
							Cpu:               aws.Int64(512),
							Memory:            aws.Int64(2048),
							MemoryReservation: aws.Int64(1024),
						},
					},
				},
			},
			Expected: []ContainerDefinition{
				{
					Cpu:               512,
					Memory:            2048,
					MemoryReservation: 1024,
				},
			},
		},
	}

	for i, c := range cases {
		optimizer.ecs = mockedECSiface{
			DescribeServicesResp:       c.DescribeServicesResp,
			DescribeTaskDefinitionResp: c.DescribeTaskDefinitionResp,
		}
		if err := optimizer.loadCurrentDefinition(); err != nil {
			t.Errorf("Got %v", err)
		}
		if optimizer.currentDef["cpu"] != c.Expected[i].Cpu {
			t.Errorf("Expected %d, got %d", c.Expected[i].Cpu, optimizer.currentDef["cpu"])
		}
		if optimizer.currentDef["memory"] != c.Expected[i].Memory {
			t.Errorf("Expected %d, got %d", c.Expected[i].Memory, optimizer.currentDef["memory"])
		}
		if optimizer.currentDef["reservedMemory"] != c.Expected[i].MemoryReservation {
			t.Errorf("Expected %d, got %d", c.Expected[i].MemoryReservation, optimizer.currentDef["reservedMemory"])
		}
	}
}
