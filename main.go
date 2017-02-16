package main

import (
	"encoding/json"
	"fmt"
	"github.com/minya/domofone/lib"
	"github.com/minya/goutils/config"
	"github.com/minya/goutils/web"
	"github.com/minya/telegram"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var settings BotSettings

func Handle(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadAll(r.Body)
	var upd telegram.Update
	json.Unmarshal(bytes, &upd)

	userName := upd.Message.Chat.Username

	var userInfo UserInfo
	userInfoErr := config.UnmarshalJson(
		&userInfo, fmt.Sprintf(".domofoneBot/users/%v.json", userName))
	if nil != userInfoErr {
		fmt.Printf("error: %v\n", userInfoErr)
		SendMessage(upd.Message.Chat.Id, "Not registered")
		return
	}

	fmt.Printf("Login for user %v found: %v\n", userName, userInfo.Login)

	bal, fare, err := lib.GetDomofoneBalance(userInfo.Login, userInfo.Password)
	if nil != err {
		fmt.Printf("error: %v\n", err)
		SendMessage(upd.Message.Chat.Id, "Unable to get balance")
		return
	}

	SendMessage(
		upd.Message.Chat.Id,
		fmt.Sprintf("Balance: %v, Price: %v", bal, fare))

	io.WriteString(w, "ok")
}

func SendMessage(chatId int, msg string) {
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	sendMsgUrl := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", settings.Id)

	data := url.Values{}
	data.Add("chat_id", strconv.Itoa(chatId))
	data.Add("text", msg)
	fmt.Printf("Sending msg to %v\n", chatId)
	resp, err := client.PostForm(sendMsgUrl, data)
	if nil != err {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("%v from telegram api\n", resp.StatusCode)
}

func main() {
	config.UnmarshalJson(&settings, ".domofoneBot/settings.json")
	http.HandleFunc("/", Handle)
	http.ListenAndServe(":8080", nil)
}

type UserInfo struct {
	Login    string `json:login`
	Password string `json:password`
}

type BotSettings struct {
	Id string
}
