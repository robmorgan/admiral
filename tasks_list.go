package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/dynport/gocli"
)

type containerInstance struct {
	ContainerInstanceArn *string
	Ec2InstanceId        *string
	Ec2PrivateIp         *string
}

type containersList struct {
	Cluster string `cli:"arg"`
}

func (r *containersList) Run() error {
	if r.Cluster == "" {
		return fmt.Errorf("You must specify an ECS cluster")
	}

	svc := ecs.New(session.New(), &aws.Config{})

	// get a list of running tasks from the specified ECS cluster
	// https://docs.aws.amazon.com/sdk-for-go/api/service/ecs/#ListTasksInput
	ecsListTasksRequest := &ecs.ListTasksInput{
		Cluster:       aws.String(r.Cluster),
		DesiredStatus: aws.String("RUNNING"),
	}

	ecsListTasksResp, err := svc.ListTasks(ecsListTasksRequest)
	if err != nil {
		if ecserr, ok := err.(awserr.Error); ok && ecserr.Code() == "NoSuchEntity" {
			fmt.Printf("[WARN] No ECS Cluster by name (%s) found", r.Cluster)
			return nil
		}
		return fmt.Errorf("Error reading ECS Cluster %s: %s", r.Cluster, err)
	}

	var tasks []*ecs.Task
	var containerInstanceArns []*string
	for _, taskArn := range ecsListTasksResp.TaskArns {
		fmt.Sprintf("Describing task for task arn: %v", taskArn)

		// describe each task
		// TODO - you could optimize this code and batch all of the tasks
		// into a single request.
		describeTaskRequest := &ecs.DescribeTasksInput{
			Cluster: aws.String(r.Cluster),
			Tasks:   []*string{taskArn},
		}

		describeTaskResp, err := svc.DescribeTasks(describeTaskRequest)
		if err != nil {
			if ecserr, ok := err.(awserr.Error); ok && ecserr.Code() == "NoSuchEntity" {
				fmt.Printf("[WARN] Could not describe task definition for arn (%v)", taskArn)
				return nil
			}
			return fmt.Errorf("Error describing task definition for arn (%v): %s", taskArn, err)
		}

		containerInstanceArns = append(containerInstanceArns, describeTaskResp.Tasks[0].ContainerInstanceArn)
		tasks = append(tasks, describeTaskResp.Tasks[0])
	}

	// lookup all container instances
	containerInstancesRequest := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(r.Cluster),
		ContainerInstances: containerInstanceArns,
	}

	ecsContainerInstancesResp, err := svc.DescribeContainerInstances(containerInstancesRequest)
	if err != nil {
		if ecserr, ok := err.(awserr.Error); ok && ecserr.Code() == "NoSuchEntity" {
			fmt.Printf("[WARN] No ECS Cluster by name (%s) found", r.Cluster)
			return nil
		}
		return fmt.Errorf("Error reading ECS Cluster %s: %s", r.Cluster, err)
	}

	var containerInstanceIds []*string
	var containerInstances []*containerInstance
	var containerInstances2 []*containerInstance
	for _, cInstance := range ecsContainerInstancesResp.ContainerInstances {
		ci := &containerInstance{
			ContainerInstanceArn: cInstance.ContainerInstanceArn,
			Ec2InstanceId:        cInstance.Ec2InstanceId,
		}
		containerInstances = append(containerInstances, ci)
		containerInstanceIds = append(containerInstanceIds, cInstance.Ec2InstanceId)
	}

	ec2Svc := ec2.New(session.New(), &aws.Config{})
	describeInstancesRequest := &ec2.DescribeInstancesInput{
		InstanceIds: containerInstanceIds,
	}

	ec2DescribeInstancesResp, err := ec2Svc.DescribeInstances(describeInstancesRequest)
	instances, _ := awsutil.ValuesAtPath(ec2DescribeInstancesResp, "Reservations[].Instances[]")
	for _, instance := range instances {
		h := instance.(*ec2.Instance)
		// loop over the container slice and add the private ips
		for _, ci := range containerInstances {
			if *h.InstanceId == *ci.Ec2InstanceId {
				ci.Ec2PrivateIp = h.PrivateIpAddress
			}

			containerInstances2 = append(containerInstances2, ci)
		}
	}

	// construct table
	t := gocli.NewTable()
	t.Header("task_id", "task_definition", "created_at", "instance_ip", "cpu", "memory", "started_by")
	for _, tk := range tasks {
		instanceIp := lookupInstanceIp(*tk.ContainerInstanceArn, containerInstances2)
		if instanceIp == "" {
			instanceIp = gocli.Red("ERROR")
		}
		t.Add(formatArn(*tk.TaskArn), formatArn(*tk.TaskDefinitionArn), tk.CreatedAt.Format("2006-01-02T15:04"), instanceIp, tk.Cpu, tk.Memory, tk.StartedBy)
	}

	t.SortBy = 1
	sort.Sort(sort.Reverse(t))
	fmt.Println(t)

	return nil
}

func formatArn(s string) string {
	str := strings.SplitAfter(s, "/")
	return str[1]
}

func lookupInstanceIp(arn string, instances []*containerInstance) string {
	for _, instance := range instances {
		if arn == *instance.ContainerInstanceArn {
			return *instance.Ec2PrivateIp
		}
	}
	return ""
}
