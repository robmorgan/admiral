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
}

func (r *hostsList) Run() error {
	// Create an EC2 service object in the "eu-west-1" region
	// Note that you can also configure your region globally by
	// exporting the AWS_REGION environment variable
	svc := ec2.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})

	// Call the DescribeInstances Operation
	resp, err := svc.DescribeInstances(nil)
	if err != nil {
		panic(err)
	}

	// id launch_time ami name ip type revision role
	t := gocli.NewTable()
	t.Add("id", "launch_time", "ami", "name", "ip", "type", "revision", "role")

	instances, _ := awsutil.ValuesAtPath(resp, "Reservations[].Instances[]")
	for _, instance := range instances {
		// fmt.Printf("%v", instance)
		h := instance.(*ec2.Instance)
		tags := aggregateTags(h.Tags)
		// TODO - calculate role from tags
		role := gocli.Red("NONE")
		t.Add(h.InstanceId, h.LaunchTime.Format("2006-01-02T15:04"), h.ImageId, tags["Name"], h.PrivateIpAddress, h.InstanceType, "aabbcc", role)
	}

	t.SortBy = 1
	sort.Sort(sort.Reverse(t))

	fmt.Println(t)
	return nil
}

func aggregateTags(tags []*ec2.Tag) map[string]string {
	tagsHash := map[string]string{}
	for _, t := range tags {
		tagsHash[p2s(t.Key)] = p2s(t.Value)
	}
	return tagsHash
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
