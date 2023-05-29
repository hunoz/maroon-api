package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	v1 "github.com/hunoz/maroon-api/api/v1"
	"github.com/hunoz/maroon-api/authentication"
)

const (
	BETA = "beta"
	PROD = "prod"
)

var ginLambda *ginadapter.GinLambdaV2
var ginRouter *gin.Engine

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func setGinMode() {
	var stage string
	if stage = os.Getenv("STAGE"); stage == "" {
		stage = BETA
	}

	if strings.ToLower(stage) == PROD {
		gin.SetMode(gin.ReleaseMode)
	}
}

func getRegionAndPoolId() (region string, poolId string) {
	var cognitoRegion string
	var cognitoPoolId string
	if cognitoRegion = os.Getenv("COGNITO_REGION"); cognitoRegion == "" {
		color.Red("'COGNITO_REGION' environment variable not set!")
		os.Exit(1)
	}
	if cognitoPoolId = os.Getenv("COGNITO_POOL_ID"); cognitoPoolId == "" {
		color.Red("'COGNITO_POOL_ID' environment variable not set!")
		os.Exit(1)
	}

	return cognitoRegion, cognitoPoolId
}

func setupRoutes() {
	setGinMode()

	cognitoRegion, cognitoPoolId := getRegionAndPoolId()

	auth := authentication.NewAuth(&authentication.Config{
		CognitoRegion:     cognitoRegion,
		CognitoUserPoolID: cognitoPoolId,
	})

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(authentication.JWTMiddleware(*auth))

	api := router.Group("/api")

	v1Api := api.Group("/v1")

	v1Api.GET("/console-url", v1.GetConsoleUrl)

	ginRouter = router
}

func init() {
	log.Println("Setting up routes")
	setupRoutes()
}

func main() {
	if _, exists := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); exists {
		log.Println("Running in lambda mode")
		ginLambda = ginadapter.NewV2(ginRouter)
		lambda.Start(Handler)
	} else {
		log.Println("Running in local mode")
		ginRouter.Run("127.0.0.1:8080")
	}
}
