package main

import (
	"bytes"
	"flag"
	"github.com/strongjz/contino-bucks/cbuck"
	"log"
	"os"
)

var (
	verificationToken string
	oauthToken        string
	dynamodb  string
	buf               bytes.Buffer
	debug    bool
	logger            = log.New(&buf, "logger: ", log.LstdFlags)
)

func main() {

	logger.SetOutput(os.Stdout)

	flag.StringVar(&verificationToken, "token", "YOUR_VERIFICATION_TOKEN_HERE", "Your Slash Verification Token")
	flag.StringVar(&oauthToken, "oauth", "Oauth token", "Your Oauth Verification Token")
	flag.StringVar(&dynamodb, "dynamodb", "dynamodb", "dynamodb Endpoint")
	flag.BoolVar(&debug, "debug", false, "Show JSON output")

	flag.Parse()

	if debug {logger.Printf("[INFO] Verification Token: %s", verificationToken)}
	if debug {logger.Printf("[INFO] OAUTH Token: %s", oauthToken)}

	c := cbuck.New(dynamodb,verificationToken, oauthToken)

	c.Start(debug)

}
