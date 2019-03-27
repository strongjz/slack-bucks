package cbuck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)



var (
	moneyGifLink = "https://media.giphy.com/media/12Eo7WogCAoj84/giphy.gif"
	buf    bytes.Buffer
	logger = log.New(&buf, "logger: ", log.LstdFlags)
	helpMSG = "Please try command again /cbuck give @USERNAME AMOUNT"
)

type Cbuck struct{
	verificationToken string
	oauthtoken string
}

func New(verificationToken string, oauthtoken string) *Cbuck {

	c := new(Cbuck)
	c.verificationToken = verificationToken
	c.oauthtoken = oauthtoken

	return c


}

func (c *Cbuck) Start() {

	logger.SetOutput(os.Stdout)

	http.HandleFunc("/slash", func(w http.ResponseWriter, r *http.Request) {
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			logger.Printf("[ERROR] parsing slash command %s", err)
			returnHTTPMSG(helpMSG, w,http.StatusOK)
			return
		}

		if !s.ValidateToken(c.verificationToken) {
			logger.Printf("[ERROR] Token unauthorized")
			returnHTTPMSG("[ERROR] unauthorized", w,http.StatusForbidden)

		}

		logger.Printf("[INFO] S: %s", s)


		switch s.Command {
		case "/echo":
			logger.Printf("[INFO] Text %s\n", s.Text)
			returnHTTPMSG(fmt.Sprintf("%s from User %s", s.Text, s.UserName), w,http.StatusOK)

		case "/cbuck":

			c.cbuck(s,w)
		case "/":

		default:
			fmt.Println("Default case / was hit")
			returnHTTPMSG(helpMSG, w, http.StatusOK)
			return
		}
	})

	logger.Printf("[INFO] Server listening")

	err := http.ListenAndServe(":3000", nil)

	if err != nil {
		logger.Fatalf("[ERROR] Starting HTTP Service %s", err)
	}

	}

func (c *Cbuck) cbuck(s slack.SlashCommand , w http.ResponseWriter ) {

	//needs the bot oauth

	api := slack.New(c.oauthtoken)

	_, err := api.AuthTest()
	if err != nil {
		logger.Printf("[ERROR] Auth Error: %s", err.Error())

		w.WriteHeader(http.StatusForbidden)
		return
	}

	//give who amount
	text := s.Text
	logger.Printf("[INFO] Received Text: %s\n", text)

	//GIVER
	giverUser := s.UserName

	give, err := regexp.MatchString(`^(give)`, text)
	if err != nil {

		logger.Printf("[ERROR] on Give match: %s\n", err.Error())
		returnHTTPMSG("Not wanting to give Contino Bucks?",w,http.StatusOK)

		return
	}

	//get user cbuck is for
	var receiverInfo *slack.User
	//how much are they getting
	var amount int

	if give {

		//RECEIVER
		receiverInfo, err = c.findReceiver(text)
		if err != nil{
			logger.Printf("[ERROR] There was an error with the User: Recieved Text %s", text)

			msg := fmt.Sprintf("Please try again there was an error with the User \n%s",helpMSG)
			returnHTTPMSG(msg,w,http.StatusOK)
			return
		}



		//AMOUNT
		amount,err = findAmount(text)
		if err != nil{
			logger.Printf("[ERROR] There was an error with the Amount: %s", text)

			msg := fmt.Sprintf("Please try again there was an error with the Amount \n%s",helpMSG)
			returnHTTPMSG(msg,w,http.StatusOK)
			return
		}


		logger.Printf("[INFO] Reciver ID : %s\n", receiverInfo.ID)
		logger.Printf("[INFO] Amount: %d\n", amount)

	}else{
		logger.Printf("[ERROR] Not giving so no idea what there doing\n")
		returnHTTPMSG(helpMSG,w,http.StatusOK)
		return
	}

	logger.Printf("[INFO] Reciever ID: %s, Fullname: %s, Email: %s\n", receiverInfo.ID, receiverInfo.Profile.RealName, receiverInfo.Profile.Email)


	//Write to the DATABASE HERE


	//send ack to the giver and receiver
	err = c.sendACK(s.UserID,giverUser,amount,receiverInfo)
	if err != nil{
		logger.Printf("[ERROR] There Send the ACKS: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	}


func (c *Cbuck) sendACK(giverID string, giverUser string, amount int, receiverInfo *slack.User) error {

	//RECEIVER MESSAGE
	returnMsg := fmt.Sprintf("%s gave you %d Contino Bucks\n",giverUser,amount)

	err := c.sendSlackIM(receiverInfo.ID,returnMsg)
	if err != nil {
		logger.Printf("[ERROR] Sending %s Message: %s\n", receiverInfo.Profile.RealName, err.Error())
		return err
	}

	//GIVER MESSAGE
	giverMsg := fmt.Sprintf("You gave %s %d Contino Bucks\n %s",receiverInfo.Name,amount, moneyGifLink)

	err = c.sendSlackIM(giverID,giverMsg)
	if err != nil {
		logger.Printf("[ERROR] Sending Giver Message: %s\n", err.Error())
		return err
	}

	return nil

}
func (c *Cbuck) findReceiver(text string) (*slack.User, error){

	api := slack.New(c.oauthtoken)

	receiverMatch := regexp.MustCompile(`<@\w+\|.+>`)

	receiverID := receiverMatch.FindString(text)

	//look up receivers ID not username
	//<@UH5RMGCF2|james.strong> ID comes in that form
	receiverID = strings.TrimPrefix(receiverID, "<@")
	receiverIDArray := strings.Split(receiverID, "|")
	receiverID = receiverIDArray[0]

	//Get all the RECEIVERS information
	receiverInfo, err := api.GetUserInfo(receiverID)

	if err != nil {
		logger.Printf("[ERROR] User %s can not be found\n", receiverID)
		return nil, err

	}

	return receiverInfo, nil
}

func findAmount(text string) (int, error) {

	amountMatch := regexp.MustCompile(`>\s\d+`)
	amountStr := amountMatch.FindString(text)
	amountStr = strings.TrimPrefix(amountStr, "> ")

	logger.Printf("[INFO] Amount String Match: %s\n", amountStr)

	amount, err := strconv.Atoi(amountStr)

	if err != nil {
		logger.Printf("[ERROR] String to int conversion on amount %s\n", err.Error())
		return -1, err
	}
	return amount, nil
}

func (c *Cbuck) sendSlackIM(userID string, message string) error {

	api := slack.New(c.oauthtoken)

	//let them know they got cbucks from someone
	_, _, channelID, err := api.OpenIMChannel(userID)
	if err != nil {
		logger.Printf("[ERROR] Sending %s Message: %s\n", userID, err)
		return err
	}

	logger.Printf("[INFO] %s", message)

	_, _, err = api.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil{
		logger.Printf("[ERROR] Sending Message: %s\n", err)
		return err
	}

	return nil
}

func returnHTTPMSG(msg string,w http.ResponseWriter, status int)  {

	logger.Printf("[INFO] Sending message: %s\n", msg)

	params := &slack.Msg{Text: msg}

	b, err := json.Marshal(params)
	if err != nil {
		logger.Printf("[ERROR] Marshalling Slack return message %s", msg)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

	return

}