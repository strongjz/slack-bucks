package buck

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func (b *Buck) sendACK(giverID string, giverUser string, amount float64, receiverInfo *slack.User) error {

	//RECEIVER MESSAGE
	var receiverMsg string

	if amount == 0.00 {
		receiverMsg = fmt.Sprintf("%s gave you %.2f Slack Bucks\n%s", giverUser, amount, thumbsDownGif)
	} else {
		receiverMsg = fmt.Sprintf("%s gave you %.2f Slack Bucks\n%s", giverUser, amount, moneyGifLink)
	}

	err := b.sendSlackIM(receiverInfo.ID, receiverMsg)
	if err != nil {
		log.Printf("[ERROR] Sending %s Message: %s\n", receiverInfo.Profile.RealName, err.Error())
		return err
	}

	//GIVER MESSAGE
	giverMsg := fmt.Sprintf("You gave %s %.2f Slack Bucks\n", receiverInfo.Name, amount)

	err = b.sendSlackIM(giverID, giverMsg)
	if err != nil {
		log.Printf("[ERROR] Sending Giver Message: %s\n", err.Error())
		return err
	}

	return nil

}

func (b *Buck) findReceiver(text string) (*slack.User, error) {

	receiverMatch := regexp.MustCompile(`<@\w+\|.+>`)

	receiverID := receiverMatch.FindString(text)

	//look up receivers ID not username
	//<@UH5RMGCF2|james.strong> ID comes in that form
	receiverID = strings.TrimPrefix(receiverID, "<@")
	receiverIDArray := strings.Split(receiverID, "|")
	receiverID = receiverIDArray[0]

	//Get all the RECEIVERS information
	receiverInfo, err := b.api.GetUserInfo(receiverID)
	if err != nil {
		log.Printf("[ERROR] User %s can not be found\n", receiverID)
		return nil, err

	}

	return receiverInfo, nil
}

func findAmount(text string) (float64, error) {

	amountMatch := regexp.MustCompile(`>\s\d+`)
	amountStr := amountMatch.FindString(text)
	amountStr = strings.TrimPrefix(amountStr, "> ")

	log.Printf("[INFO] Amount String Match: %s\n", amountStr)

	amount, err := strconv.ParseFloat(amountStr, 64)

	if err != nil {
		log.Printf("[ERROR] String to int conversion on amount %s\n", err.Error())
		return -1, err
	}

	if amount <= -1 {
		log.Printf("[INFO] Someone tried to take Slack Bucks %s\n", err.Error())
		return -1, err
	}

	amountRD := amount
	//round up to the nearest 2 decimal places
	if amount != 0 {
		amountRD = math.Floor(amount*100) / 100
	}

	return amountRD, nil
}

func (b *Buck) sendSlackIM(userID string, message string) error {

	//let them know they got Bucks from someone
	_, _, channelID, err := b.api.OpenIMChannel(userID)
	if err != nil {
		log.Printf("[ERROR] Sending %s Message: %s\n", userID, err)
		return err
	}

	log.Printf("[INFO] %s", message)

	_, _, err = b.api.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Printf("[ERROR] Sending Message: %s\n", err)
		return err
	}

	return nil
}

func returnSlackMSG(msg string) ([]byte, error) {

	log.Printf("[INFO] Sending message: %s\n", msg)

	params := &slack.Msg{Text: msg}

	b, err := json.Marshal(params)
	if err != nil {
		log.Printf("[ERROR] Marshalling Slack return message %s", msg)
		return nil, err
	}

	return b, nil

}
