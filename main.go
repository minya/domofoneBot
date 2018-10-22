package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/minya/domofone/lib"
	"github.com/minya/goutils/config"
	"github.com/minya/goutils/web"
	"github.com/minya/telegram"
)

var settings BotSettings

var logPath string
var port int

func init() {
	const (
		defaultLogPath = "domofone.log"
		defaultPort    = 8080
	)
	flag.StringVar(&logPath, "logpath", defaultLogPath, "Path to write logs")
	flag.IntVar(&port, "port", defaultPort, "Port to listen on")
}

func handle(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadAll(r.Body)
	var upd telegram.Update
	json.Unmarshal(bytes, &upd)

	userName := upd.Message.Chat.Username

	var userInfo UserInfo
	userInfoErr := config.UnmarshalJson(
		&userInfo, fmt.Sprintf("~/.domofoneBot/users/%v.json", userName))

	if nil != userInfoErr {
		fmt.Printf("error: %v\n", userInfoErr)
		sendMessage(upd.Message.Chat.Id, "Not registered")
		return
	}

	fmt.Printf("Login for user %v found: %v\n", userName, userInfo.Login)

	bal, fare, err := lib.GetDomofoneBalance(userInfo.Login, userInfo.Password)
	if nil != err {
		fmt.Printf("error: %v\n", err)
		sendMessage(upd.Message.Chat.Id, "Unable to get balance")
		return
	}

	sendMessage(
		upd.Message.Chat.Id,
		fmt.Sprintf("Balance: %v, Price: %v", bal, fare))

	io.WriteString(w, "ok")
}

func sendMessage(chatId int, msg string) {
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
	setUpLogger()
	log.Printf("Start. Listen on %v.\n", port)
	config.UnmarshalJson(&settings, "~/.domofoneBot/settings.json")
	http.HandleFunc("/", handle)
	http.ListenAndServe(":8080", nil)
}

func setUpLogger() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

type UserInfo struct {
	Login    string `json:login`
	Password string `json:password`
}

type BotSettings struct {
	Id string
}
