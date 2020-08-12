package weibo

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"

	telegraphGO "github.com/MakeGolangGreat/telegraph-go"
)

// Save 通过处理知乎链接，保存到Telegraph，然后返回线上链接的入口函数。
func Save(url string, CreatePageRequest *telegraphGO.CreatePageRequest) (link string, err error) {
	if isWeiboStatus(url) {
		// 微博博文
		data, err2 := getSingleStatus(url, CreatePageRequest)
		errHandler("获取豆瓣广告失败", err2)
		link, err = telegraphGO.CreatePage(data)
	} else if isWeiboStatusPC(url) {
		link = "暂时不提供微博PC版链接内容备份，请提供对应的移动端链接（m.weibo.cn这种）"
		err = nil
	}

	return
}

//IsWeiboLink 如果是微博链接，返回true
func IsWeiboLink(url string) bool {
	reg := regexp.MustCompile(`http.*weibo\.cn|om`)
	return reg.MatchString(url)
}

// 微博博文链接。
// 微博博文有好几种URL，对应不同的页面结构。
// demo: https://m.weibo.cn/detail/4536843289235930?display=0&retcode=6102
// 但PC端链接容易触发登录限制，因此暂时只提供移动端内容爬取。
func isWeiboStatus(url string) bool {
	reg := regexp.MustCompile(`http.*m.weibo\.cn|om\/detail\/\d+.*`)
	return reg.MatchString(url)
}

func isWeiboStatusPC(url string) bool {
	reg := regexp.MustCompile(`http.*weibo\.cn|om\/\d+.*`)
	return reg.MatchString(url)
}

// 获取单独的博文内容，爬虫获取静态页面，直接从页面中正则匹配拿到script中的数据
func getSingleStatus(url string, data *telegraphGO.CreatePageRequest) (*telegraphGO.CreatePageRequest, error) {
	spider := colly.NewCollector()
	extensions.RandomUserAgent(spider)
	extensions.Referer(spider)

	spider.OnRequest(func(req *colly.Request) {
		fmt.Printf("fetching: %s\n", req.URL.String())
	})

	spider.OnResponse(func(res *colly.Response) {
		dom, err := goquery.NewDocumentFromReader(bytes.NewReader((res.Body)))
		errHandler("初始化goquery失败", err)

		// 博文内容
		dom.Find("body script").Each(func(_ int, s *goquery.Selection) {
			// 判断是否存在src属性，不存在的话说明是里面有微博内容的那个标签。
			_, exists := s.Attr("src")
			if !exists {
				text := s.Text()
				// 尝试正则匹配出微博正文。
				contentReg := regexp.MustCompile(`(?s)"text": "(.*)",.*"textLength"`)
				data.Data = contentReg.FindStringSubmatch(text)[1]

				// 匹配用户名
				authorReg := regexp.MustCompile(`"screen_name": "(.*)",`)
				data.AuthorName = authorReg.FindStringSubmatch(text)[1]

				titleReg := regexp.MustCompile(`"status_title": "(.*)",`)
				data.Title = titleReg.FindStringSubmatch(text)[1]
			}
		})
	})

	var err error
	// Set error handler
	spider.OnError(func(r *colly.Response, wrong error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", "\nError:", wrong)
		err = wrong
	})

	spider.Visit(url + "?display=0&retcode=6102")

	return data, err
}

func errHandler(msg string, err error) {
	if err != nil {
		fmt.Printf("%s - %s\n", msg, err)
		os.Exit(1)
	}
}
