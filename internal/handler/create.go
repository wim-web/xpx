package handler

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cf_types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2_types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	myaws "github.com/wim-web/xpx/internal/aws"
	"github.com/wim-web/xpx/internal/view"
)

func CreateHandler() error {
	vpc, err := selectVpc()

	if err != nil {
		return err
	}

	subnetIds, err := myaws.PublicSubnetIdsFromVpcId(*vpc.VpcId)

	if err != nil {
		return err
	}

	now := time.Now()
	stackName := fmt.Sprintf("%s-%s", "xpx", now.Format("20060102150405"))
	templateFile := "network.yaml"

	template, err := ioutil.ReadFile(templateFile)

	if err != nil {
		return errors.New("unable to read CloudFormation template file")
	}

	myaws.CreateFromStack(stackName, string(template), []cf_types.Parameter{
		{
			ParameterKey:   aws.String("VpcId"),
			ParameterValue: vpc.VpcId,
		},
		{
			ParameterKey:   aws.String("SubnetId"),
			ParameterValue: &subnetIds[0],
		},
	})

	return nil
}

func selectVpc() (vpc ec2_types.Vpc, err error) {
	ec2Svc := ec2.NewFromConfig(myaws.Cfg)

	describeVpcsOutput, err := ec2Svc.DescribeVpcs(context.Background(), &ec2.DescribeVpcsInput{})

	if err != nil {
		return vpc, errors.New("failed to describe VPCs")
	}

	finder := view.NewFinder(
		describeVpcsOutput.Vpcs,
		func(v ec2_types.Vpc) string {
			return *v.VpcId
		},
		func(v ec2_types.Vpc) string {
			var m string
			for _, tag := range v.Tags {
				m += fmt.Sprintf("%s: %s\n", *tag.Key, *tag.Value)
			}

			return m
		},
		"Select VPC",
	)

	return finder.Find()
}
