package main

import (
	"archive-bot/common"
	"archive-bot/douban"
	"archive-bot/weibo"
	"archive-bot/weixin"
	"archive-bot/zhihu"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/fatih/color"
	"golang.org/x/net/proxy"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

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

	// test()
	botTokenFlag := flag.String("bot-token", "", "Telegram bot token")
	socks5Flag := flag.String("proxy", "", "socks5 proxy schema")
	telegraphTokenFlag := flag.String("telegraph-token", "", "telegraph token")
	flag.Parse()

	botToken = *botTokenFlag
	socks5 = *socks5Flag
	telegraphToken = *telegraphTokenFlag

	fmt.Println("len(botToken) ", len(botToken))

	// 如果没有从参数重获取到botToken，说明程序运行在本地，那么从配置文件中读取即可。
	if botToken == "" {
		readConfig()
	}

	start()
	// test2()
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
	var bot *tgbotapi.BotAPI
	var err error
	// 如果不需要代理（比如跑在Github Action上）
	if socks5 == "" {
		bot, err = tgbotapi.NewBotAPI(botToken)
		fmt.Println("没有经过代理")
	} else {
		// 跑在本地就需要代理
		client := createProxyClient()
		bot, err = tgbotapi.NewBotAPIWithClient(botToken, client)
	}
	errHandler("初始化bot失败", err)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = false
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

		replyMessage := "没有监测到任何链接！"
		// 如果能匹配到链接
		if linkRegExp.MatchString(updateText) {
			// 拿到链接，但有可能是个错误的链接。
			link := linkRegExp.FindString(updateText)

			replyMessage = "监测到链接：" + link + " 开始备份..."
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyMessage)
			bot.Send(msg)

			data := telegraphGO.CreatePageRequest{
				AccessToken: telegraphToken,
				AuthorURL:   link,
				Title:       "内容备份",
				Data:        projectDesc,
			}

			var pageLink string
			var err error
			fmt.Println(link)

			if zhihu.IsZhihuLink(link) {
				pageLink, err = zhihu.Save(link, &data)
				replyMessage = pageLink
			} else if douban.IsDoubanLink(link) {
				pageLink, err = douban.Save(link, &data)
				replyMessage = pageLink
			} else if weibo.IsWeiboLink(link) {
				pageLink, err = weibo.Save(link, &data)
				replyMessage = pageLink
			} else if weixin.IsWeixinLink(link) {
				pageLink, err = weixin.Save(link, &data)
				replyMessage = pageLink
			} else {
				fmt.Println("非适配链接")
				pageLink, err = common.Save(link, &data)
				replyMessage = pageLink
			}

			errHandler("内容备份失败", err)
			if err != nil {
				replyMessage = "内容备份失败，报错信息：" + err.Error()
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
