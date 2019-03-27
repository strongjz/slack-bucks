package cbuck

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"k8s.io/kubernetes/pkg/kubelet/kubeletconfig/util/log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)



var (
	moneyGifLink = "https://media.giphy.com/media/12Eo7WogCAoj84/giphy.gif"

)

type Cbuck struct{
	verificationToken string
	oauthtoken string

}

func NewCbuck(verificationToken string, oauthtoken string) *Cbuck {

	c := new(Cbuck)
	c.verificationToken = verificationToken
	c.oauthtoken = oauthtoken

	return c


}

func (c *Cbuck) Start() {

	http.HandleFunc("/slash", func(w http.ResponseWriter, r *http.Request) {
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			log.Errorf("[ERROR] parsing slash command %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !s.ValidateToken(c.verificationToken) {
			fmt.Printf("[ERROR] Token unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		fmt.Println("[INFO] S: ", s)


		switch s.Command {
		case "/echo":
			fmt.Printf("[INFO] Text %s\n", s.Text)

			msg := fmt.Sprintf("%s from User %s", s.Text, s.UserName)

			params := &slack.Msg{Text: msg}

			b, err := json.Marshal(params)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			fmt.Printf("[INFO] B: %s\n",b)

			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
		case "/cbuck":

			c.cbuck(s,w)
		case "/":

		default:
			fmt.Println("Default case / was hit")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}

func (c *Cbuck) cbuck(s slack.SlashCommand ,w http.ResponseWriter ) {

	//needs the bot oauth

	api := slack.New(c.oauthtoken)

	_, err := api.AuthTest()
	if err != nil {
		fmt.Printf("[ERROR] Auth Error: %s", err.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}

	text := s.Text


	fmt.Printf("[INFO] Received Text: %s\n", text)

	give, err := regexp.MatchString(`^(give)`, text)

	fmt.Printf("[INFO] Give Match: %t Err: %s\n",give, err)

	if err != nil {
		fmt.Printf("[ERROR] on regex match: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	receiverMatch := regexp.MustCompile(`<@\w+\|.+>`)
	amountMatch := regexp.MustCompile(`>\s\d+`)


	//get user cbuck is for
	var receiverID string

	//how much are they getting
	var amount int

	if give {

		receiverID = receiverMatch.FindString(text)

		amountStr := amountMatch.FindString(text)
		amountStr = strings.TrimPrefix(amountStr, "> ")

		fmt.Printf("[INFO] Amount String Match: %s\n", amountStr)

		amount, err = strconv.Atoi(amountStr)
		if err != nil {
			fmt.Printf("[ERROR] String to int conversion on amount %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//look up receivers ID not username
		//<@UH5RMGCF2|james.strong> ID comes in that form
		receiverID = strings.TrimPrefix(receiverID, "<@")
		receiverIDArray := strings.Split(receiverID, "|")
		receiverID = receiverIDArray[0]

		fmt.Printf("[INFO] Reciver ID : %s\n", receiverID)

		fmt.Printf("[INFO] Amount: %d\n", amount)


	}else{
		fmt.Printf("[ERROR] Not giving so no idea what there doing\n")

		w.Header().Set("Content-Type", "application/json")
		params := &slack.Msg{Text: "Please try command again /cbuck give @USERNAME $AMOUNT"}

		b, err := json.Marshal(params)
		if err != nil {
			fmt.Printf("[ERROR] Marshalling message: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
		return
	}

	//get user giving cbucks
	giverUser := s.UserName


	fmt.Printf("[INFO] Getting Reciver user info: %s\n", receiverID)

	receiverInfo, err := api.GetUserInfo(receiverID)
	if err != nil {

		fmt.Printf("[ERROR] User %s can not be found\n", receiverID)

		w.Header().Set("Content-Type", "application/json")

		params := &slack.Msg{Text: fmt.Sprintf("User %s can not be found", receiverID)}

		b, err := json.Marshal(params)
		if err != nil {
			fmt.Printf("[ERROR] Marshalling message: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
		return
	}

	fmt.Printf("[INFO] ID: %s, Fullname: %s, Email: %s\n", receiverInfo.ID, receiverInfo.Profile.RealName, receiverInfo.Profile.Email)

	returnMsg := fmt.Sprintf("%s gave you %d Contino Bucks\n %s",giverUser,amount, moneyGifLink)

	err = c.sendMsg(receiverInfo.ID,returnMsg)
	if err != nil {
		fmt.Printf("[ERROR] Sending %s Message: %s\n", receiverInfo.Profile.RealName, err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	}


func (c *Cbuck) sendMsg(userID string, message string) error {

	api := slack.New(c.oauthtoken)

	//let them know they got cbucks from someone
	_, _, channelID, err := api.OpenIMChannel(userID)
	if err != nil {
		fmt.Printf("[ERROR] Sending %s Message: %s\n", userID, err)

		return err
	}

	fmt.Printf("[INFO] %s", message)

	_, _, err = api.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil{
		fmt.Printf("[ERROR] Sending Message: %s\n", err)

		return err
	}

	return nil
}