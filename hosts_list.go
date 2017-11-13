package main

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/dynport/gocli"
)

type hostsList struct {
	//Query string `cli:"opt"`
}

func (r *hostsList) Run() error {
	// Create an EC2 service object in the "eu-west-1" region
	// Note that you can also configure your region globally by
	// exporting the AWS_REGION environment variable
	svc := ec2.New(session.New(), &aws.Config{})

	// we are only concerned with running instances
	filterChain := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-state-name"),
				Values: aws.StringSlice([]string{"running"}),
			},
			//&ec2.Filter{
			//	Name:   aws.String("tag:Env"),
			//	Values: aws.StringSlice([]string{"staging"}),
			//},
		},
	}

	// filter by environment
	//Query   string `cli:"arg"`

	// Call the DescribeInstances Operation
	resp, err := svc.DescribeInstances(filterChain)
	if err != nil {
		panic(err)
	}

	// id launch_time ami name ip type revision role
	t := gocli.NewTable()
	t.Header("id", "launch_time", "ami", "name", "ip", "public_ip", "type", "revision", "role")

	instances, _ := awsutil.ValuesAtPath(resp, "Reservations[].Instances[]")
	for _, instance := range instances {
		h := instance.(*ec2.Instance)
		tags := aggregateTags(h.Tags)
		// TODO - calculate role from tags
		role := gocli.Red("NONE")
		t.Add(h.InstanceId, h.LaunchTime.Format("2006-01-02T15:04"), h.ImageId, tags["Name"], h.PrivateIpAddress, h.PublicIpAddress, h.InstanceType, "aabbcc", role)
	}

	t.SortBy = 1
	sort.Sort(sort.Reverse(t))

	fmt.Println(t)
	return nil
}

func truncate(s string, length int) string {
	if len(s) > length {
		return s[0:length]
	}
	return s
}

func p2s(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}
