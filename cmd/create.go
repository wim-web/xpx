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
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
