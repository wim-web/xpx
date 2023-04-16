package cmd

import (
	_ "embed"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
)

//go:embed network.yaml
var cloudformationFile string

var rootCmd = &cobra.Command{
	Use:   "xpx",
	Short: "build tunnel easily",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	if cloudformationFile == "" {
		template, err := ioutil.ReadFile("network.yaml")
		if err != nil {
			log.Fatalln(err)
		}
		cloudformationFile = string(template)
	}
}
