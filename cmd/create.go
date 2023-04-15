/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	c_types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create tunnel server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadDefaultConfig(context.Background())

		if err != nil {
			log.Fatalln("failed to load aws configure: ", err)
		}

		svc := ec2.NewFromConfig(cfg)
		input := &ec2.DescribeVpcsInput{}

		res, err := svc.DescribeVpcs(context.Background(), input)

		if err != nil {
			log.Fatalln("failed to describe VPCs: ", err)
		}

		idx, err := fuzzyfinder.Find(
			res.Vpcs,
			func(i int) string {
				return *res.Vpcs[i].VpcId
			},
			fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
				if i == -1 {
					return ""
				}

				var m string

				for _, tag := range res.Vpcs[i].Tags {
					m += fmt.Sprintf("%s: %s\n", *tag.Key, *tag.Value)
				}

				return m
			}),
			fuzzyfinder.WithHeader("Select VPC"),
		)

		if err != nil {
			log.Fatalln("unexpected error: ", err)
		}

		vpc := res.Vpcs[idx]

		res2, err := svc.DescribeInternetGateways(context.Background(), &ec2.DescribeInternetGatewaysInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("attachment.vpc-id"),
					Values: []string{*vpc.VpcId},
				},
			},
		})

		if err != nil {
			log.Fatalln("error describing internet gateways: ", err)
		}

		if len(res2.InternetGateways) < 1 {
			log.Fatalln("selected vpc have no internet gateway")
		}

		response, err := svc.DescribeRouteTables(context.Background(), &ec2.DescribeRouteTablesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{*vpc.VpcId},
				},
				{
					Name:   aws.String("route.destination-cidr-block"),
					Values: []string{"0.0.0.0/0"},
				},
				{
					Name:   aws.String("route.gateway-id"),
					Values: []string{*res2.InternetGateways[0].InternetGatewayId},
				},
			},
		})
		if err != nil {
			log.Fatalln("failed to describe route tables: ", err)
		}

		if len(response.RouteTables) < 1 {
			log.Fatalln("no route table attached with internet gateway")
		}

		subnetId := *response.RouteTables[0].Associations[0].SubnetId

		client := cloudformation.NewFromConfig(cfg)

		createFromStack(client, vpc.VpcId, &subnetId)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func createFromStack(svc *cloudformation.Client, vpcId *string, subnetId *string) {
	now := time.Now()
	stackName := fmt.Sprintf("%s-%s", "xpx", now.Format("20060102150405"))
	templateFile := "network.yaml"

	template, err := ioutil.ReadFile(templateFile)
	if err != nil {
		log.Fatalf("unable to read CloudFormation template file, %v", err)
	}

	input := &cloudformation.CreateStackInput{
		Tags: []c_types.Tag{
			{Key: aws.String("CreatedBy"), Value: aws.String("xpx")},
		},
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(string(template)),
		Capabilities: []c_types.Capability{
			c_types.CapabilityCapabilityNamedIam,
		},
		Parameters: []c_types.Parameter{
			{
				ParameterKey:   aws.String("VpcId"),
				ParameterValue: vpcId,
			},
			{
				ParameterKey:   aws.String("SubnetId"),
				ParameterValue: subnetId,
			},
		},
	}

	_, err = svc.CreateStack(context.Background(), input)
	if err != nil {
		log.Fatalf("unable to create stack, %v", err)
	}

	fmt.Printf("Stack %s has been created successfully.\n", stackName)

	waitForStackCompletion(svc, stackName)

	fmt.Printf("Stack %s has been created successfully.\n", stackName)
}

func waitForStackCompletion(client *cloudformation.Client, stackName string) {
	fmt.Printf("Waiting for stack %s to complete...\n", stackName)

	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	for {
		resp, err := client.DescribeStacks(context.Background(), describeStacksInput)
		if err != nil {
			log.Fatalf("unable to describe stack, %v", err)
		}

		stack := resp.Stacks[0]

		switch stack.StackStatus {
		case c_types.StackStatusCreateComplete:
			return
		case c_types.StackStatusCreateInProgress:
			// Continue polling
		case c_types.StackStatusDeleteInProgress:
		case c_types.StackStatusDeleteComplete:
			return
		default:
			log.Fatalf("Stack creation failed with status: %s", stack.StackStatus)
		}

		// Sleep for a while before polling again
		fmt.Print(".")
		time.Sleep(3 * time.Second)
	}
}
