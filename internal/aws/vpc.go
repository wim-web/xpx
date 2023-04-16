package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func PublicSubnetIdsFromVpcId(vpcId string) ([]string, error) {
	ec2Svc := ec2.NewFromConfig(Cfg)

	igw, err := getInternetGW(ec2Svc, vpcId)

	if err != nil {
		return nil, err
	}

	return getPublicSubnetIds(ec2Svc, vpcId, *igw.InternetGatewayId)
}

func getInternetGW(ec2Svc *ec2.Client, vpcId string) (igw types.InternetGateway, err error) {
	describeInternetGatewaysOutput, err := ec2Svc.DescribeInternetGateways(context.Background(), &ec2.DescribeInternetGatewaysInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: []string{vpcId},
			},
		},
	})

	if err != nil {
		return igw, errors.New("failed to describe internet gateway")
	}

	if len(describeInternetGatewaysOutput.InternetGateways) < 1 {
		return igw, errors.New("selected vpc have no internet gateway")
	}

	igw = describeInternetGatewaysOutput.InternetGateways[0]

	return
}

func getPublicSubnetIds(ec2Svc *ec2.Client, vpcId string, igwId string) ([]string, error) {
	describeRouteTablesOutput, err := ec2Svc.DescribeRouteTables(context.Background(), &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
			{
				Name:   aws.String("route.destination-cidr-block"),
				Values: []string{"0.0.0.0/0"},
			},
			{
				Name:   aws.String("route.gateway-id"),
				Values: []string{igwId},
			},
		},
	})

	if err != nil {
		return nil, errors.New("failed to describe route tables")
	}

	if len(describeRouteTablesOutput.RouteTables) < 1 {
		return nil, errors.New("no route table attached with internet gateway")
	}

	return []string{*describeRouteTablesOutput.RouteTables[0].Associations[0].SubnetId}, nil
}
