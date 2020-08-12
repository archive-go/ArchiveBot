package main

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

func test() {
	// 获取单独的知乎回答，爬虫解决静态页面
	spider := colly.NewCollector()
	extensions.RandomUserAgent(spider)
	extensions.Referer(spider)

	spider.OnRequest(func(req *colly.Request) {
		fmt.Printf("fetching: %s\n", req.URL.String())
	})

	spider.OnResponse(func(res *colly.Response) {
		dom, err := goquery.NewDocumentFromReader(bytes.NewReader((res.Body)))
		errHandler("初始化goquery失败", err)

		// 回答内容
		dom.Find("body script").Each(func(_ int, s *goquery.Selection) {
			// 判断是否存在src属性，不存在的话说明是里面有微博内容的那个标签。
			_, exists := s.Attr("src")
			if !exists {
				text := s.Text()
				// fmt.Println(text)
				// 尝试正则匹配出微博正文。
				reg := regexp.MustCompile(`(?s)<div class=\"WB_text.*`)
				got := reg.FindStringSubmatch(text)
				fmt.Println(got)

				// fmt.Println("匹配到了？", got[1])
			}
		})
	})

	// Set error handler
	spider.OnError(func(r *colly.Response, wrong error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", "\nError:", wrong)
	})

	spider.Visit("https://weibo.com/2656274875/JfxuT9UWK?type=comment")
}
