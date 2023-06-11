package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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
		logrus.Errorf("Error getting IAM user credentials: %s", err.Error())
		return nil, errors.Wrap(err, "Error getting IAM user credentials")
	}

	var credentials IamCredentials

	if err = json.Unmarshal([]byte(*output.SecretString), &credentials); err != nil {
		logrus.Errorf("Error parsing IAM user credentials: %s", err.Error())
		return nil, errors.Wrap(err, "Error parsing IAM user credentials")
	}

	return &credentials, nil
}

func assumeRole(roleArn string, username string, duration int32, cfg *aws.Config) (*types.Credentials, error) {
	iamCredentials, err := getIamCredentials()
	if err != nil {
		return nil, err
	}
	var conf aws.Config
	if cfg == nil {
		conf, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(iamCredentials.AccessKeyId, iamCredentials.SecretAccessKey, ""),
			),
		)
	} else {
		conf = *cfg
	}
	if err != nil {
		return nil, errors.Wrap(err, "Error creating config")
	}
	client := sts.NewFromConfig(conf)
	output, err := client.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		DurationSeconds: &duration,
		RoleSessionName: aws.String(fmt.Sprintf("MaroonApi-%s", username)),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error Assuming Role")
	}

	return output.Credentials, nil
}

type AssumeRoleInput struct {
	RoleArn         string `json:"roleArn" binding:"required" form:"roleArn"`
	SessionDuration int32  `json:"sessionDuration" binding:"required,numeric,min=900,max=43200" form:"sessionDuration"`
}

type AssumeRoleOutput struct {
	AccessKeyId     string `json:"accessKeyid"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      time.Time
}

func toCamelCase(str string) string {
	firstLetter := str[0]
	return strings.ToLower(string(firstLetter)) + str[1:]
}

func AssumeRole(ctx *gin.Context) {
	input := AssumeRoleInput{}
	username, _ := ctx.Get("cognito:username")

	if err := ctx.ShouldBindQuery(&input); err != nil {
		err := parseBindingError(err)
		ctx.JSON(err.Status, err)
		return
	}

	matches, _ := regexp.MatchString(`^arn:aws:iam::\d{12}:role/[0-9A-Za-z_+=,.@-]{1,64}$`, input.RoleArn)
	if !matches {
		logrus.Errorf("String does not match role ARN regex: %s", input.RoleArn)
		err := BadRequestError()
		ctx.JSON(err.Status, err)
		return
	}

	credentials, err := assumeRole(input.RoleArn, username.(string), input.SessionDuration, nil)
	if err != nil {
		logrus.Errorf("Error fetching role credentials: %s", err.Error())
		var e *RestError
		if strings.Contains(err.Error(), "is not authorized to perform") {
			e = ForbiddenError()
		} else {
			e = BadRequestError()
		}
		ctx.JSON(e.Status, e)
		return
	}

	ctx.JSON(200, AssumeRoleOutput{
		AccessKeyId:     *credentials.AccessKeyId,
		SecretAccessKey: *credentials.SecretAccessKey,
		SessionToken:    *credentials.SessionToken,
		Expiration:      *credentials.Expiration,
	})
}
