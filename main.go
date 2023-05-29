package main

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	v1 "github.com/hunoz/maroon-api/api/v1"
	"github.com/hunoz/maroon-api/authentication"
	"github.com/hunoz/maroon-api/logging"
	"github.com/sirupsen/logrus"
)

const (
	BETA = "beta"
	PROD = "prod"
)

var stage string
var ginLambda *ginadapter.GinLambdaV2
var ginRouter *gin.Engine

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func setGinMode() {
	switch stageVariable := strings.ToLower(os.Getenv("STAGE")); stageVariable {
	case PROD:
		stage = stageVariable
		gin.SetMode(gin.ReleaseMode)
	default:
		stage = BETA
	}
}

func getRegionAndPoolId() (region string, poolId string) {
	var cognitoRegion string
	var cognitoPoolId string
	if cognitoRegion = os.Getenv("COGNITO_REGION"); cognitoRegion == "" {
		logrus.Fatal("'COGNITO_REGION' environment variable not set!")
		os.Exit(1)
	}
	if cognitoPoolId = os.Getenv("COGNITO_POOL_ID"); cognitoPoolId == "" {
		logrus.Fatal("'COGNITO_POOL_ID' environment variable not set!")
		os.Exit(1)
	}

	return cognitoRegion, cognitoPoolId
}

func setupRoutes() {
	cognitoRegion, cognitoPoolId := getRegionAndPoolId()

	auth := authentication.NewAuth(&authentication.Config{
		CognitoRegion:     cognitoRegion,
		CognitoUserPoolID: cognitoPoolId,
	})

	router := gin.New()
	router.Use(logging.JSONLogMiddleware(stage))
	router.Use(gin.Recovery())
	router.Use(authentication.JWTMiddleware(*auth))

	api := router.Group("/api")

	v1Api := api.Group("/v1")

	v1Api.GET("/console-url", v1.GetConsoleUrl)
	v1Api.GET("/assume-role", v1.AssumeRole)

	ginRouter = router
}

func init() {
	setGinMode()
	logging.SetLogMode(stage)
	logrus.Info("Setting up routes")
	setupRoutes()
}

func main() {
	if _, exists := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); exists {
		logrus.Info("Running in lambda mode")
		ginLambda = ginadapter.NewV2(ginRouter)
		lambda.Start(Handler)
	} else {
		logrus.Info("Running in local mode")
		ginRouter.Run("127.0.0.1:8080")
	}
}
