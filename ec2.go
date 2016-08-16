package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Get a list of running EC2 instances
func LoadInstances() ([]*ec2.Instance, error) {
	svc := ec2.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})

	// we are only concerned with running instances
	filterChain := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-state-name"),
				Values: aws.StringSlice([]string{"running"}),
			},
		},
	}

	// Call the DescribeInstances Operation
	resp, err := svc.DescribeInstances(filterChain)
	instances, _ := awsutil.ValuesAtPath(resp, "Reservations[].Instances[]")

	ret := make([]*ec2.Instance, len(instances))
	for i := range instances {
		ret[i] = instances[i].(*ec2.Instance)
	}

	return ret, err
}

// Get an instance's name based on the tags
func GetInstanceName(instance *ec2.Instance) string {
	tags := aggregateTags(instance.Tags)
	if val, ok := tags["Name"]; ok {
		//do something here
		return val
	}
	return "(unknown)"

}

func aggregateTags(tags []*ec2.Tag) map[string]string {
	tagsHash := map[string]string{}
	for _, t := range tags {
		tagsHash[p2s(t.Key)] = p2s(t.Value)
	}
	return tagsHash
}
