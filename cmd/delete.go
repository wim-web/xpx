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
	"github.com/spf13/cobra"
)

var (
	arnFlag = "arn"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete stack",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString(arnFlag)

		if err != nil {
			log.Fatalln(err)
		}

		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalln("failed to load AWS config: ", err)
		}

		svc := cloudformation.NewFromConfig(cfg)

		_, err = svc.DeleteStack(context.TODO(), &cloudformation.DeleteStackInput{
			StackName: aws.String(name),
		})
		if err != nil {
			log.Fatalln("failed to delete stack: ", err)
		}

		waitForStackCompletion(svc, name)

		fmt.Printf("Successfully initiated deletion of stack '%s'\n", name)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().String(arnFlag, "", "stack arn")
	deleteCmd.MarkFlagRequired(arnFlag)
}
