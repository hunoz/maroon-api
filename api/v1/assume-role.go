package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

var errMessage = color.New(color.FgRed).SprintFunc()

type IamCredentials struct {
	AccessKeyId     string `json:"AccessKeyId" binding:"required"`
	SecretAccessKey string `json:"SecretAccessKey" binding:"required"`
}

func getIamCredentials() (*IamCredentials, error) {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	client := secretsmanager.NewFromConfig(cfg)

	output, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("MaroonApiIamUser"),
	})
	if err != nil {
		log.Printf("%s\n", errMessage("Error getting IAM user credentials: ", err.Error()))
		return nil, errors.Wrap(err, "Error getting IAM user credentials")
	}

	var credentials IamCredentials

	if err = json.Unmarshal([]byte(*output.SecretString), &credentials); err != nil {
		log.Printf("%s\n", errMessage("Error parsing IAM user credentials: ", err.Error()))
		return nil, errors.Wrap(err, "Error parsing IAM user credentials")
	}

	return &credentials, nil
}

func AssumeRole(roleArn string, username string, duration int32) (*types.Credentials, error) {
	iamCredentials, err := getIamCredentials()
	if err != nil {
		return nil, err
	}
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(iamCredentials.AccessKeyId, iamCredentials.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating config")
	}
	client := sts.NewFromConfig(cfg)
	output, err := client.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		DurationSeconds: &duration,
		RoleSessionName: aws.String(fmt.Sprintf("MaroonApi-%s", username)),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error Assumeing Role")
	}

	return output.Credentials, nil
}
