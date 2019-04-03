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
	db                string
	buf               bytes.Buffer
	debug             bool
	logger            = log.New(&buf, "logger: ", log.LstdFlags)
)

func main() {

	logger.SetOutput(os.Stdout)

	flag.StringVar(&verificationToken, "token", "YOUR_VERIFICATION_TOKEN_HERE", "Your Slash Verification Token")
	flag.StringVar(&oauthToken, "oauth", "Oauth token", "Your Oauth Verification Token")
	flag.StringVar(&db, "db", "db", "db Endpoint")
	flag.BoolVar(&debug, "debug", false, "Show JSON output")

	flag.Parse()

	if debug {
		logger.Printf("[INFO] Verification Token: %s", verificationToken)
		logger.Printf("[INFO] OAUTH Token: %s", oauthToken)
		logger.Printf("[INFO] DB Endpoint: %s", db)

		logger.Printf("[INFO] Main: Creating New CBuck")
	}

	c := cbuck.New(db, verificationToken, oauthToken)


	if debug {logger.Print("[INFO] Main: Starting Cbuck")}

	c.Start(debug)

}
