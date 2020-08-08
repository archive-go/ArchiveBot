package main

import (
	"archive-bot/zhihu"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/fatih/color"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/proxy"

	telegraphGO "github.com/MakeGolangGreat/telegraph-go"
)

var botToken string
var telegraphToken string
var socks5 string

func errHandler(msg string, err error) {
	if err != nil {
		fmt.Printf("%s - %s\n", msg, err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("GOGOGO")

	readConfig()
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
	client := createProxyClient()
	bot, err := tgbotapi.NewBotAPIWithClient(botToken, client)
	errHandler("初始化bot失败", err)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	// 持续监测Bot收到的消息
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		updateText := update.Message.Text

		linkRegExp, _ := regexp.Compile(`(http.*)\s?`)

		replyMessage := "没有监测到知乎链接！"
		// 如果抓到了知乎链接
		if linkRegExp.MatchString(updateText) {
			data := telegraphGO.CreatePageRequest{
				AccessToken:   telegraphToken,
				ReturnContent: "true",
				AuthorURL:     projectLink,
				Title:         "知乎回答备份",
				AuthorName:    "由「" + projectName + "」备份",
				Data:          updateText + "\n\n",
			}
			// 拿到链接，但有可能是个错误的链接。
			link := linkRegExp.FindString(updateText)

			if zhihu.IsZhihuLink(link) {
				pageLink, err := zhihu.Save(link, &data)
				errHandler("知乎链接备份失败", err)
				replyMessage = pageLink
			}
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyMessage)

		bot.Send(msg)
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

// 检查此链接是否之前已经备份过，如果备份过，直接返回上次备份的链接
// 但不确定如何实现。关键在于如何保存每次的记录。本地数据库？那意味着将要长久地租一台服务器...
// 每次将保存记录保存在一个telegra.ph文章里？那么并发将是个问题，毕竟每次都要先读取telegra.ph链接来获取记录以及每次都要编辑telegra.ph文章。太频繁了。
func checkExist(link string) {
}
