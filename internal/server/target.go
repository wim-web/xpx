package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	myaws "github.com/wim-web/xpx/internal/aws"
)

type TargetServer interface {
	Port() int
	Host() string
	VpcId() string
}

func CreateTargetServer(host string) (t TargetServer, err error) {
	if strings.HasSuffix(host, "rds.amazonaws.com") {
		t, err = CreateRdsServer(host)
		return
	}

	return t, errors.New("no support host")
}

type RdsServer struct {
	host  string
	port  int
	vpcId string
}

func (s RdsServer) Port() int {
	return s.port
}

func (s RdsServer) Host() string {
	return s.host
}

func (s RdsServer) VpcId() string {
	return s.vpcId
}

func CreateRdsServer(host string) (RdsServer, error) {
	rdsSvc := rds.NewFromConfig(myaws.Cfg)
	r := RdsServer{
		host: host,
	}

	switch getRdsEndpointType(host) {
	case cluster:
		describeDBClustersOutput, err := rdsSvc.DescribeDBClusters(context.Background(), &rds.DescribeDBClustersInput{})

		if err != nil {
			return r, errors.New("failed to describe db clusters")
		}

		for _, cluster := range describeDBClustersOutput.DBClusters {
			if *cluster.Endpoint == host || *cluster.ReaderEndpoint == host {
				r.port = int(*cluster.Port)
				describeDBSubnetGroupsOutput, err := rdsSvc.DescribeDBSubnetGroups(context.Background(), &rds.DescribeDBSubnetGroupsInput{
					DBSubnetGroupName: cluster.DBSubnetGroup,
				})
				if err != nil {
					return r, errors.New("failed to describe db subnet group")
				}
				r.vpcId = *describeDBSubnetGroupsOutput.DBSubnetGroups[0].VpcId
				return r, nil
			}
		}
	case instance:
		describeDBInstancesOutput, err := rdsSvc.DescribeDBInstances(context.Background(), &rds.DescribeDBInstancesInput{})

		if err != nil {
			return r, errors.New("failed to describe db instances")
		}

		for _, instance := range describeDBInstancesOutput.DBInstances {
			if *instance.Endpoint.Address == host {
				r.port = int(instance.Endpoint.Port)
				r.vpcId = *instance.DBSubnetGroup.VpcId
				return r, nil
			}
		}
	}

	return r, fmt.Errorf("not found host: %s", host)
}

type rdsEndpointType string

const (
	cluster  rdsEndpointType = "cluster"
	instance rdsEndpointType = "instance"
)

func getRdsEndpointType(host string) rdsEndpointType {
	s := strings.Split(host, ".")

	if strings.Contains(s[1], "cluster") {
		return cluster
	}

	return instance
}
