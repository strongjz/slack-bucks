package main

import (
	"flag"
	"fmt"
	"github.com/strongjz/contino-bucks/cbuck"
)

var (
	verificationToken string
	oauthToken string
)

func main() {


	flag.StringVar(&verificationToken, "token", "YOUR_VERIFICATION_TOKEN_HERE", "Your Slash Verification Token")
	flag.StringVar(&oauthToken, "oauth", "Oauth token", "Your Oauth Verification Token")
	flag.Parse()

	fmt.Println("[INFO] Token Read as", verificationToken)

	c := cbuck.NewCbuck(verificationToken, oauthToken)

	c.Start()

}