package main

import (
	"bytes"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/strongjz/slack-bucks/buck"
	"log"
	"os"
)

var (
	verificationToken string
	oauthToken        string
	db                string
	buf               bytes.Buffer
	debug             bool
	logger            = log.New(&buf, "logger: ", log.LstdFlags)
	ginLambda *ginadapter.GinLambda
)


func init () {

	logger.SetOutput(os.Stdout)

	verificationToken := os.Getenv("verificationToken")
	oauthToken := os.Getenv("oauthToken")
	db := os.Getenv("db")

	logger.Printf("[INFO] Verification Token: %s", verificationToken)
	logger.Printf("[INFO] OAUTH Token: %s", oauthToken)
	logger.Printf("[INFO] DB Endpoint: %s", db)
	logger.Printf("[INFO] Main: Creating New Slack Bucks")

	c := buck.New(db, verificationToken, oauthToken)

	logger.Print("[INFO] Main: Starting Slack Bucks")


	ginLambda = ginadapter.New(c.Start())

}

func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	logger.Println("Lambda request", req.RequestContext.RequestID)

	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.Proxy(req)
}


func main() {

	lambda.Start(Handler)
}
