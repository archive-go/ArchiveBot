package main

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

func test2() {
	// reg := regexp.MustCompile(`http.*weibo\.cn`)
	// fmt.Println(reg.MatchString("https://mp.weixin.qq.com/s/bMFAgkvZdh7YJiTFwP7QSw"))
	spider := colly.NewCollector()
	extensions.RandomUserAgent(spider)
	extensions.Referer(spider)

	spider.OnRequest(func(req *colly.Request) {
		req.Headers.Set("Cookie", "pac_uid=0_c72e541fad774; iip=0; pgv_pvid=9044765905; pgv_info=ssid=s1970066010; rewardsn=; wxtokenkey=777")
		req.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		req.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		req.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
		req.Headers.Set("Cache-Control", "no-cache")
		req.Headers.Set("Connection", "keep-alive")
		req.Headers.Set("DNT", "1")
		req.Headers.Set("Host", "mp.weixin.qq.com")
		req.Headers.Set("Pragma", "no-cache")
		req.Headers.Set("Sec-Fetch-Dest", "document")
		req.Headers.Set("Sec-Fetch-Mode", "navigate")
		req.Headers.Set("Sec-Fetch-Site", "none")
		req.Headers.Set("Sec-Fetch-User", "71")
		req.Headers.Set("Upgrade-Insecure-Requests", "1")
		req.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_16_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36")
	})

	spider.OnResponse(func(res *colly.Response) {
		dom, err := goquery.NewDocumentFromReader(bytes.NewReader((res.Body)))
		errHandler("初始化goquery失败", err)

		// 标题
		dom.Find("#js_article #activity-name").Each(func(_ int, s *goquery.Selection) {
			// data.Title = s.Text()
			// fmt.Println(s.Text())
		})

		// dom.Find("#js_article #js_name").Each(func(_ int, s *goquery.Selection) {
		// 	data.AuthorName = s.Text()
		// })

		dom.Find("#js_article #js_content").Each(func(_ int, s *goquery.Selection) {
			html, err := s.Html()
			errHandler("解析内容html失败", err)
			// data.Data += html
			fmt.Println(html)
		})

		// // 文章发布时间
		// dom.Find("#page-content  div.article_modify_area").Each(func(_ int, s *goquery.Selection) {
		// 	time := s.Text()
		// 	// 在文章尾部增加发布时间
		// 	data.Data += "<br/><blockquote>" + time + "</blockquote>"
		// })
	})

	var err error
	// Set error handler
	spider.OnError(func(r *colly.Response, wrong error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", "\nError:", wrong)
		err = wrong
		fmt.Println(err)
	})

	spider.Visit("https://mp.weixin.qq.com/s/bMFAgkvZdh7YJiTFwP7QSw")
}

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
