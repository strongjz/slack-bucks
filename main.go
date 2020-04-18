package main

import (
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/strongjz/go_example_app/app"
	"log"
)

var (
	ginLambda         *ginadapter.GinLambda
)

func init() {

	log.Printf("[INFO] Main: Starting")

	app := app.New()

	ginLambda = ginadapter.New(app.Engine())

}

func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	log.Println("[INFO] Lambda request", req.RequestContext.RequestID)

	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.Proxy(req)
}

func main() {

	lambda.Start(Handler)
}
