package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	myaws "github.com/wim-web/xpx/internal/aws"
)

func ListHandler() error {
	cfClient := cloudformation.NewFromConfig(myaws.Cfg)

	listStacksOutput, err := cfClient.ListStacks(context.Background(), &cloudformation.ListStacksInput{
		StackStatusFilter: []types.StackStatus{"CREATE_COMPLETE"},
	})

	if err != nil {
		return errors.New("failed to list stacks")
	}

	for _, stack := range listStacksOutput.StackSummaries {
		tags, err := getStackTags(cfClient, *stack.StackId)

		if err != nil {
			return errors.New("failed to get tags of stack")
		}

		for _, tag := range tags {
			if myaws.IsCreatedFromXpx(tag) {
				fmt.Printf("%s | %s\n", *stack.StackName, *stack.StackId)
			}
		}
	}

	return nil
}

func getStackTags(client *cloudformation.Client, stackID string) ([]types.Tag, error) {
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackID),
	}

	resp, err := client.DescribeStacks(context.TODO(), params)
	if err != nil {
		return nil, err
	}

	if len(resp.Stacks) == 0 {
		return nil, fmt.Errorf("no stack found with ID: %s", stackID)
	}

	return resp.Stacks[0].Tags, nil
}
