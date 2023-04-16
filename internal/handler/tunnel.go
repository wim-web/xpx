package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	cf_types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/wim-web/tonneeeeel/pkg/command"
	"github.com/wim-web/tonneeeeel/pkg/port"
	myaws "github.com/wim-web/xpx/internal/aws"
	"github.com/wim-web/xpx/internal/server"
)

func TunnelHandler(host string, localPort int, template string) error {
	s, err := server.CreateTargetServer(host)

	if err != nil {
		return err
	}

	subnetIds, err := myaws.PublicSubnetIdsFromVpcId(s.VpcId())

	if err != nil {
		return err
	}

	now := time.Now()
	stackName := fmt.Sprintf("%s-%s", "xpx", now.Format("20060102150405"))

	if err != nil {
		return errors.New("unable to read CloudFormation template file")
	}

	stackId, outputs, err := myaws.CreateFromStack(stackName, template, []cf_types.Parameter{
		{
			ParameterKey:   aws.String("VpcId"),
			ParameterValue: aws.String(s.VpcId()),
		},
		{
			ParameterKey:   aws.String("SubnetId"),
			ParameterValue: &subnetIds[0],
		},
	})

	if err != nil {
		return err
	}

	clusterName, taskId, containerId, err := getEcsInfo(outputs)

	fmt.Println(clusterName, taskId, containerId)

	if err != nil {
		return err
	}

	sp := ssm.NewFromConfig(myaws.Cfg)

	if localPort == 0 {
		localPort, err = port.AvailablePort()
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	ch := make(chan string, 1)

	cmd, err := command.PortForwardCommand(
		ctx,
		sp,
		clusterName,
		taskId,
		containerId,
		myaws.Cfg.Region,
		command.REMOTE_PORT_FORWARD_DOCUMENT_NAME,
		map[string][]string{
			"portNumber":      {strconv.Itoa(s.Port())},
			"localPortNumber": {strconv.Itoa(localPort)},
			"host":            {host},
		},
	)

	if err != nil {
		return err
	}

	go func() {
		err := cmd.Run()
		if err != nil {
			ch <- err.Error()
			return
		}
		close(ch)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTSTP)

	go func() {
		<-sigCh
		cmd.Cancel()
	}()

	<-ch

	fmt.Println("delete resources...")

	err = DeleteHandler(stackId)

	return err
}

func getEcsInfo(outputs []types.Output) (string, string, string, error) {
	var clusterName string
	var serviceName string

	for _, output := range outputs {
		if *output.OutputKey == "ClusterName" {
			clusterName = *output.OutputValue
		}
		if *output.OutputKey == "ServiceName" {
			serviceName = *output.OutputValue
		}
	}

	ecsSvc := ecs.NewFromConfig(myaws.Cfg)

	listTasksOutput, err := ecsSvc.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Cluster:     aws.String(clusterName),
		ServiceName: aws.String(serviceName),
	})

	if err != nil {
		return "", "", "", errors.New("unable to list tasks")
	}

	describeTasksOutput, err := ecsSvc.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterName),
		Tasks:   listTasksOutput.TaskArns,
	})

	if err != nil {
		return "", "", "", errors.New("unable to describe tasks")
	}

	task := describeTasksOutput.Tasks[0]
	container := task.Containers[0]

	return clusterName, strings.Split(*task.TaskArn, "/")[2], *container.RuntimeId, nil
}
