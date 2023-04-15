/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/wim-web/xpx/internal/handler"
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

		err = handler.DeleteHandler(name)

		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().String(arnFlag, "", "stack arn")
	deleteCmd.MarkFlagRequired(arnFlag)
}
