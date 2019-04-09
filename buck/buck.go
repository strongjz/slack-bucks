package buck

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
	"github.com/strongjz/slack-bucks/database"
	"log"
	"net/http"
	"os"
	"regexp"
	"errors"

)

var (
	moneyGifLink  = "https://media.giphy.com/media/12Eo7WogCAoj84/giphy.gif"
	selfishGif    = "https://media.giphy.com/media/mXVpbjG2qmC2Y/giphy.gif"
	thumbsDownGif = "https://media.giphy.com/media/9NEH2NjjMA4hi/giphy.gif"
	helpMSG       = "Please try command again /buck give @USERNAME AMOUNT"
	buf           bytes.Buffer
	debug         bool
	logger        = log.New(&buf, "logger: ", log.LstdFlags)
)

type Buck struct {
	verificationToken string
	oauthtoken        string
	dynamodbEndpoint  string
	api               *slack.Client
	db                *database.DB
	router			 *gin.Engine
}

func New(dynamodbEndpoint string, verificationToken string, oauthtoken string) *Buck {

	logger.Printf("[INFO] New Buck, db: %s", dynamodbEndpoint)

	return &Buck{
		verificationToken,
		oauthtoken,
		dynamodbEndpoint,
		slack.New(oauthtoken),
		database.New(dynamodbEndpoint),
		gin.Default(),
	}
}

func (b *Buck) Start() *gin.Engine {

	logger.SetOutput(os.Stdout)

	logger.Print("[INFO] Starting routerEngine")

	// set server mode
	gin.SetMode(gin.DebugMode)


	// Global middleware
	b.router.Use(gin.Logger())
	b.router.Use(gin.Recovery())

	b.router.POST("/buck", b.buckHandler)
	b.router.POST("/echo", b.echoHandler)
	b.router.POST("/", b.rootHandler)

	return b.router
}

func (b *Buck) echoHandler(c *gin.Context) {

	s, err := b.validateSlackMsg(c)
	if err != nil {
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusOK, msg)
		return
	}

	logger.Printf("[INFO] S: %s", s)

	switch s.Command {

	case "/echo":
		logger.Printf("[INFO] Text %s\n", s.Text)
		msg, _ := returnSlackMSG(fmt.Sprintf("%s from User %s", s.Text, s.UserName))
		c.JSON(http.StatusOK, msg)
		return

	case "/":
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusOK, msg)
		return

	default:
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusOK, msg)
		return

	}
	return
}

func (b *Buck) buckHandler(c *gin.Context) {

	s, err := b.validateSlackMsg(c)

	if err != nil {
		logger.Printf("[ERROR] Validating slack message: %s", err)
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusOK, msg)
		return
	}

	logger.Printf("[INFO] S: %s", s)
	logger.Printf("[INFO] Response URL: %s", s.ResponseURL)
	switch s.Command {

	case "/buck":
		b.buck(s, c)
	case "/":
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusOK, msg)
		return

	default:
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusOK, msg)
		return
	}

	return
}

func (b *Buck) validateSlackMsg(c *gin.Context) (*slack.SlashCommand, error) {

	s, err := slack.SlashCommandParse(c.Request)
	if err != nil {
		logger.Printf("[ERROR] parsing slash command %s", err)
		return nil, err
	}

	if !s.ValidateToken(b.verificationToken) {
		logger.Printf("[ERROR] Token unauthorized")
		return nil, errors.New("[ERROR] Token unauthorized")
	}


	return &s, nil
}

func (b *Buck) rootHandler(c *gin.Context) {
	msg, _ := returnSlackMSG(helpMSG)
	c.JSON(http.StatusOK, msg)
	return
}

func (b *Buck) buck(s *slack.SlashCommand, c *gin.Context) {

	_, err := b.api.AuthTest()
	if err != nil {
		logger.Printf("[ERROR] Auth Error: %s", err.Error())
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusOK, msg)
		return
	}

	//give who amount
	text := s.Text
	logger.Printf("[INFO] Received Text: %s\n", text)

	var g database.Gift

	//GIVER
	g.Giver = s.UserName

	give, err := regexp.MatchString(`^(give)`, text)
	if err != nil {

		logger.Printf("[ERROR] on Give match: %s\n", err.Error())
		returnSlackMSG("Not wanting to give Contino Bucks?")

		return
	}

	//get user Buck is for
	var receiverInfo *slack.User
	//how much are they getting
	var amount float64

	if give {
		//RECEIVER
		receiverInfo, err = b.findReceiver(text)
		if err != nil {
			logger.Printf("[ERROR] There was an error with the User: Recieved Text %s", text)

			msg, _ := returnSlackMSG(fmt.Sprintf("Please try again there was an error with the User \n%s", helpMSG))
			c.JSON(http.StatusOK, msg)

			return
		}

		if receiverInfo.ID == g.Giver {
			logger.Printf("[INFO] You can't keep give yourself Contino bucks: %s", text)

			msg, _ := returnSlackMSG(fmt.Sprintf(" You can't keep give yourself Contino bucks\n %s", selfishGif))
			c.JSON(http.StatusOK, msg)
			return
		}

		//AMOUNT
		amount, err = findAmount(text)
		if err != nil {
			logger.Printf("[ERROR] There was an error with the Amount: %s", text)
			msg, _ := returnSlackMSG(fmt.Sprintf("Please try again there was an error with the Amount \n%s", helpMSG))
			c.JSON(http.StatusOK, msg)
			return
		}

		logger.Printf("[INFO] Reciver ID : %s\n", receiverInfo.ID)
		logger.Printf("[INFO] Amount: %f\n", amount)

	} else {
		logger.Printf("[ERROR] Not giving so no idea what there doing\n")
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusOK, msg)
		return
	}

	logger.Printf("[INFO] Reciever ID: %s, Fullname: %s, Email: %s\n", receiverInfo.ID, receiverInfo.Profile.RealName, receiverInfo.Profile.Email)

	g.Receiver = receiverInfo.ID
	g.Amount = amount

	/*
		//Write to the DATABASE HERE

		err = c.updateDB(g)
		if err != nil {
			logger.Printf("[ERROR] Updating db error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	*/

	//send ack to the giver and receiver
	err = b.sendACK(s.UserID, g.Giver, amount, receiverInfo)
	if err != nil {
		logger.Printf("[ERROR] There Send the ACKS: %s", err.Error())
		msg, _ := returnSlackMSG(helpMSG)
		c.JSON(http.StatusInternalServerError, msg)
		return
	}

}
