package handler

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	myaws "github.com/wim-web/xpx/internal/aws"
)

func DeleteHandler(stackName string) error {
	svc := cloudformation.NewFromConfig(myaws.Cfg)

	_, err := svc.DeleteStack(context.TODO(), &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		return errors.New("failed to delete stack")
	}

	err = myaws.WaitForStackCompletion(stackName)

	return err
}
