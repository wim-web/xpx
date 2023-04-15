package aws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var (
	CloudformationTagKey   = aws.String("CreatedBy")
	CloudformationTagValue = aws.String("xpx")
)

func IsCreatedFromXpx(t types.Tag) bool {
	return t.Key == CloudformationTagKey && t.Value == CloudformationTagValue
}

func CreateFromStack(stackName string, templateBody string, parameters []types.Parameter) error {
	cfSvc := cloudformation.NewFromConfig(Cfg)

	input := &cloudformation.CreateStackInput{
		Tags: []types.Tag{
			{Key: CloudformationTagKey, Value: CloudformationTagValue},
		},
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(templateBody),
		Capabilities: []types.Capability{
			types.CapabilityCapabilityNamedIam,
		},
		Parameters: parameters,
	}

	_, err := cfSvc.CreateStack(context.Background(), input)

	if err != nil {
		return errors.New("failed to create stack")
	}

	err = WaitForStackCompletion(stackName)

	return err
}

func WaitForStackCompletion(stackName string) error {
	cfSvc := cloudformation.NewFromConfig(Cfg)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	for {
		describeStacksOutput, err := cfSvc.DescribeStacks(context.Background(), describeStacksInput)

		if err != nil {
			return errors.New("failed to describe stack")
		}

		stack := describeStacksOutput.Stacks[0]

		switch stack.StackStatus {
		case types.StackStatusCreateComplete, types.StackStatusDeleteComplete:
			return nil
		case types.StackStatusCreateInProgress, types.StackStatusDeleteInProgress:
			// continue
		default:
			return errors.New("unexpected stack status")
		}

		// Sleep for a while before polling again
		fmt.Print(".")
		time.Sleep(3 * time.Second)
	}
}
