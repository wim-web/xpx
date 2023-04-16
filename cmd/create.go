/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/wim-web/xpx/internal/handler"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create tunnel server",
	Run: func(cmd *cobra.Command, args []string) {
		err := handler.CreateHandler(cloudformationFile)

		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
