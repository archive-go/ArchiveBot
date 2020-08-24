package douban

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
	if isDoubanStatusLink(url) {
		// 备份单个豆瓣广告内容
		data, err2 := getSingleStatus(url, CreatePageRequest)
		errHandler("获取豆瓣广告失败", err2)
		link, err = telegraphGO.CreatePage(data)
	} else if isDoubanNoteLink(url) {
		// 备份单个专栏文章
		data, err2 := getSingleNote(url, CreatePageRequest)
		errHandler("获取豆瓣日记失败", err2)
		link, err = telegraphGO.CreatePage(data)
	}

	return
}

//IsDoubanLink 如果是豆瓣链接，返回true
func IsDoubanLink(url string) bool {
	fmt.Println("匹配到豆瓣链接")
	reg := regexp.MustCompile(`http.*douban\.com.*`)
	return reg.MatchString(url)
}

// 豆瓣广播链接
func isDoubanStatusLink(url string) bool {
	reg := regexp.MustCompile(`http.*douban\.com\/people/.*/status\/\d+`)
	return reg.MatchString(url)
}

// 豆瓣日记链接
func isDoubanNoteLink(url string) bool {
	reg := regexp.MustCompile(`http.*douban\.com\/note\/\d+`)
	return reg.MatchString(url)
}

// 获取单独的豆瓣广播内容，爬虫解决静态页面
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

		// 标题
		dom.Find("#content h1").Each(func(_ int, s *goquery.Selection) {
			data.Title = s.Text()
		})

		// 豆瓣用户名
		dom.Find(".status-item .hd .lnk-people").Each(func(_ int, s *goquery.Selection) {
			data.AuthorName = s.Text()
		})

		// 广播内容
		dom.Find(".status-item .status-saying").Each(func(_ int, s *goquery.Selection) {
			html, err := s.Html()
			errHandler("解析内容html失败", err)
			data.Data += html
		})

		// 文章发布时间
		dom.Find(".status-item .hd .pubtime span").Each(func(_ int, s *goquery.Selection) {
			time := s.Text()
			// 在文章尾部增加发布时间
			data.Data += "<br/><blockquote>" + time + "</blockquote>"
		})
	})

	var err error
	// Set error handler
	spider.OnError(func(r *colly.Response, wrong error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", "\nError:", wrong)
		err = wrong
	})

	spider.Visit(url)

	return data, err
}

// 获取单独的豆瓣日记内容，爬虫解决静态页面
func getSingleNote(url string, data *telegraphGO.CreatePageRequest) (*telegraphGO.CreatePageRequest, error) {
	spider := colly.NewCollector()
	extensions.RandomUserAgent(spider)
	extensions.Referer(spider)

	spider.OnRequest(func(req *colly.Request) {
		fmt.Printf("fetching: %s\n", req.URL.String())
	})

	spider.OnResponse(func(res *colly.Response) {
		dom, err := goquery.NewDocumentFromReader(bytes.NewReader((res.Body)))
		errHandler("初始化goquery失败", err)

		// 标题
		dom.Find(".note-container .note-header h1").Each(func(_ int, s *goquery.Selection) {
			data.Title = s.Text()
		})

		// 用户名
		dom.Find(".note-container .note-author").Each(func(_ int, s *goquery.Selection) {
			data.AuthorName = s.Text()
		})

		// 内容
		dom.Find(".note-container .note").Each(func(_ int, s *goquery.Selection) {
			html, err := s.Html()
			errHandler("解析内容html失败", err)
			data.Data += html
		})

		// 文章发布时间
		dom.Find(".note-container .pub-date").Each(func(_ int, s *goquery.Selection) {
			time := s.Text()
			// 在文章尾部增加发布时间
			data.Data += "<br/><blockquote>" + time + "</blockquote>"
		})
	})

	var err error
	// Set error handler
	spider.OnError(func(r *colly.Response, wrong error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", "\nError:", wrong)
		err = wrong
	})

	spider.Visit(url)

	return data, err
}

func errHandler(msg string, err error) {
	if err != nil {
		fmt.Printf("%s - %s\n", msg, err)
		os.Exit(1)
	}
}
