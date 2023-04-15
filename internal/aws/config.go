package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	Cfg aws.Config
)

func init() {
	var err error
	Cfg, err = config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
}
