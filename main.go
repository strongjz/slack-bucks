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
	oauthToken string
	buf    bytes.Buffer
	logger = log.New(&buf, "logger: ", log.LstdFlags)


)

func main() {

	logger.SetOutput(os.Stdout)

	flag.StringVar(&verificationToken, "token", "YOUR_VERIFICATION_TOKEN_HERE", "Your Slash Verification Token")
	flag.StringVar(&oauthToken, "oauth", "Oauth token", "Your Oauth Verification Token")
	flag.Parse()

	logger.Printf("[INFO] Verification Token: %s", verificationToken)
	logger.Printf("[INFO] OAUTH Token: %s", oauthToken)
	c := cbuck.New(verificationToken, oauthToken)

	c.Start()

}