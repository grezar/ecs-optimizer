package optimizer

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type Optimizer struct {
	cloudWatch *cloudwatch.CloudWatch
	ecs        *ecs.ECS
	ecsCluster string
}

func NewOptimizer(region string, cluster string, profile string) *Optimizer {
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
	}
}

func (o *Optimizer) Run() error {
	err := o.ecs.ListServicesPages(&ecs.ListServicesInput{
		Cluster:    aws.String(o.ecsCluster),
		LaunchType: aws.String("EC2"),
		MaxResults: aws.Int64(100),
	}, func(page *ecs.ListServicesOutput, lastPage bool) bool {
		fmt.Println(page)
		return !lastPage
	})

	if err != nil {
		return err
	}

	return nil
}
