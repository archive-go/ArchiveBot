package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/fatih/color"
	"golang.org/x/net/proxy"

	"github.com/MakeGolangGreat/archive-go"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

var botToken string
var telegraphToken string
var socks5 string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func errHandler(msg string, err error) {
	if err != nil {
		fmt.Printf("%s - %s\n", msg, err)
		os.Exit(1)
	}
}

func main() {
	if IsHeroku() {
		botToken = os.Getenv("ARCHIVE_BOT_TOKEN")
		socks5 = os.Getenv("SOCKS5")
		telegraphToken = os.Getenv("ARCHIVE_TELEGRAPH_TOKEN")
	}

	fmt.Println("len(botToken) ", len(botToken))

	// 如果没有从参数重获取到botToken，说明程序运行在本地，那么从配置文件中读取即可。
	if botToken == "" {
		readConfig()
	}

	start()
}

// 从配置中读取配置
func readConfig() {
	file, err := os.OpenFile("./config.json", os.O_RDWR|os.O_CREATE, 0766) //打开或创建文件，设置默认权限
	errHandler("读取配置失败", err)
	defer file.Close()

	var conf config
	err2 := json.NewDecoder(file).Decode(&conf)
	errHandler("解码配置失败", err2)

	botToken = conf.BotToken
	telegraphToken = conf.TelegraphToken
	socks5 = conf.Socks5
	color.Green("读取配置成功，botToken is: %s\n telegraphToken is %s\nsocks5 is %s", botToken, telegraphToken, socks5)
}

// 启动Telegram Bot
func start() {
	var bot *tgbot.BotAPI
	var err error
	// 如果不需要代理（比如跑在Github Action上）
	if socks5 == "" {
		bot, err = tgbot.NewBotAPI(botToken)
		fmt.Println("没有经过代理")
	} else {
		// 跑在本地就需要代理
		client := createProxyClient()
		bot, err = tgbot.NewBotAPIWithClient(botToken, client)
	}
	errHandler("初始化bot失败", err)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = false
	u := tgbot.NewUpdate(0)
	u.Timeout = 60
	// 被取回的新消息数量（1-100）
	u.Limit = 100
	updates, err := bot.GetUpdatesChan(u)
	replyMessage := ""

	// 持续监测Bot收到的消息
	for updateObj := range updates {
		// 收到消息后，并行处理
		go func(update tgbot.Update) {
			if update.Message == nil {
				return
			}

			for _, entity := range *update.Message.Entities {
				if entity.Type == "text_link" && entity.URL != "" {
					update.Message.Text += " " + entity.URL + " "
				}
			}

			updateText := update.Message.Text

			link, saveError := archive.Save(updateText, telegraphToken, attachInfo)
			if saveError != nil {
				replyMessage = "文章保存出错：" + saveError.Error()
			} else {
				// 不知道Telegraph的换行符是什么，所以这里处理了先。
				replyMessage = link + `
	
	文章由 @beifenbot 备份
	代码开源：` + projectLink
			}

			msg := tgbot.NewMessage(update.Message.Chat.ID, replyMessage)

			bot.Send(msg)
		}(updateObj)
	}
}

// 配置翻墙用的Client
func createProxyClient() *http.Client {
	client := &http.Client{}
	tgProxyURL, err := url.Parse(socks5)
	errHandler("解析socks5失败", err)

	tgDialer, err := proxy.FromURL(tgProxyURL, proxy.Direct)
	if err != nil {
		log.Printf("Failed to obtain proxy dialer: %s\n", err)
	}
	tgTransport := &http.Transport{
		Dial: tgDialer.Dial,
	}
	client.Transport = tgTransport
	return client
}

// IsHeroku func
func IsHeroku() bool {
	return os.Getenv("WHERE") == "heroku"
}
