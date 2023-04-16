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
	hostFlag      = "host"
	localPortFlag = "local-port"
)

var (
	host      string
	localPort int
)

// tunnelCmd represents the tunnel command
var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "easy tunnel",
	Run: func(cmd *cobra.Command, args []string) {
		err := handler.TunnelHandler(host, localPort, cloudformationFile)

		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)

	tunnelCmd.Flags().StringVar(&host, hostFlag, "", "")
	tunnelCmd.Flags().IntVarP(&localPort, localPortFlag, "l", 0, "")

	tunnelCmd.MarkFlagRequired(hostFlag)
}
