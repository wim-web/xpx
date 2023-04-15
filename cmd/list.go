/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list tunnel server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalln("failed to load aws config: ", err)
		}

		cfClient := cloudformation.NewFromConfig(cfg)

		input := &cloudformation.ListStacksInput{
			StackStatusFilter: []types.StackStatus{"CREATE_COMPLETE"},
		}

		resp, err := cfClient.ListStacks(context.Background(), input)
		if err != nil {
			log.Fatalln("failed to list stacks: ", err)
		}

		// var stacks []types.StackSummary

		for _, stack := range resp.StackSummaries {
			tags, err := getStackTags(cfClient, *stack.StackId)

			if err != nil {
				log.Fatalln("Error getting stack tags:", err)
			}

			for _, tag := range tags {
				if *tag.Key == "CreatedBy" && *tag.Value == "xpx" {
					fmt.Println(*stack.StackName)
				}
			}
		}

	},
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

func init() {
	rootCmd.AddCommand(listCmd)
}
