package requestHandler

import (
	"encoding/json"
	"fmt"
	"github.com/zwirec/TGChatScanner/TGBotApi"
	"io/ioutil"
	"net/http"
	"regexp"
	"log"
	"bytes"
	"net/url"
	"github.com/rs/xid"
)

const (
	UserStatsUrl = "/stats"
)

func BotUpdateHanlder(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	logger := req.Context().Value(loggerContextKey).(*log.Logger)
	if err != nil {
		logger.Printf("Error during handling request on %s : %s", req.URL.String(), err)
		return
	}

	var update TGBotApi.Update
	err = json.Unmarshal(body, &update)
	if err != nil {
		logger.Printf("Error during unmarshaling request on %s : %s", req.URL.String(), err)
		return
	}
	var message *TGBotApi.Message

	if update.Message.MessageId != 0{
		message = &update.Message
	} else if update.EditedMessage.MessageId != 0{
		message = &update.EditedMessage
	}

	if pl := len(message.Photo); pl != 0 {
		photo := message.Photo[pl-1]
		ctx := make(map[string]interface{})
		ctx["From"] = message.From
		fb := &FileBasic{
			FileId:  photo.FileId,
			Type:    "photo",
			Context: ctx,
		}
		appContext.DownloadRequests <- fb
	} else if len(message.Entities) != 0 && message.Entities[0].Type == "bot_command" {
		if err := BotCommandRouter(&update.Message, logger); err != nil {
			logger.Println(err)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func BotCommandRouter(message *TGBotApi.Message, logger *log.Logger) error {
	r := regexp.MustCompile(`\/(start(?:group)?|mystats)?\s*`)
	command := r.FindStringSubmatch(message.Text)
	if len(command) == 0 {
		return fmt.Errorf("unexpected command %s", message.Text)
	}
	switch command[1] {
	case "start":
	case "startgroup":
		err := AddSubsription(&message.From, &message.Chat)
		if err != nil {
			return err
		}
		return err
	case "mystats":
		token, err := SetUserToken(message.From.Id)

		if err != nil {
			return err
		}
		us := BuildUserStatUrl(token)
		_, err = appContext.BotApi.SendMessage(message.Chat.Id, us, true)
		if err != nil{
			return err
		}
	}
	return nil
}
func AddSubsription(user *TGBotApi.User, chat *TGBotApi.Chat) error {
	//TODO: Add user and chat in "user-chat" association
	return nil
}

func SetUserToken(userId int) (string, error) {
	guid := xid.New()
	token := guid.String()
	//TODO: Store token in db
	return token, nil
}

func BuildUserStatUrl(token string) string {
	var buff bytes.Buffer
	buff.WriteString(appContext.Hostname)
	buff.WriteString(UserStatsUrl)
	buff.WriteString("?")
	params := url.Values{}
	params.Add("token", token)
	buff.WriteString(params.Encode())
	return buff.String()
}
