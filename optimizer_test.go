package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"testing"
)

const (
	CPUUtilizationAvgValue    = 49.53529105870167
	MemoryUtilizationAvgValue = 102.3319085080778
)

type mockedCloudwatchiface struct {
	cloudwatchiface.CloudWatchAPI
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
