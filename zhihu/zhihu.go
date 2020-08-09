package zhihu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"

	telegraphGO "github.com/MakeGolangGreat/telegraph-go"
)

//IsZhihuLink 如果是知乎链接，返回true
func IsZhihuLink(url string) bool {
	reg := regexp.MustCompile(`http.*zhihu\.com.*`)

	return reg.MatchString(url)
}

// Save 通过处理知乎链接，保存到Telegraph，然后返回线上链接的入口函数。
func Save(url string, CreatePageRequest *telegraphGO.CreatePageRequest) (link string, err error) {
	if isZhihuQuestionLink(url) {
		// 备份知乎问题下方高赞回答。
		// getAllAnswers
		// 还没考虑好该如何处理，具体保存考前的几条数据？

	} else if isZhihuAnswerLink(url) {
		// 备份单个知乎回答
		data, err2 := getSingleAnswer(url, CreatePageRequest)
		errHandler("获取知乎回答失败", err2)
		link, err = telegraphGO.CreatePage(data)
	} else if isZhihuZhuanLanLink(url) {
		// 备份单个专栏文章
		data, err2 := getSingleZhuanLan(url, CreatePageRequest)
		errHandler("获取知乎专栏失败", err2)
		link, err = telegraphGO.CreatePage(data)
	}

	return
}

// 知乎问题
func isZhihuQuestionLink(url string) bool {
	reg := regexp.MustCompile(`http.*zhihu\.com.*question\d+`)
	return reg.MatchString(url)
}

// 知乎回答
func isZhihuAnswerLink(url string) bool {
	reg := regexp.MustCompile(`http.*zhihu\.com.*question.*answer.*`)
	return reg.MatchString(url)
}

// 知乎专栏
func isZhihuZhuanLanLink(url string) bool {
	reg := regexp.MustCompile(`http.*zhuanlan.zhihu\.com/p/\d+`)
	return reg.MatchString(url)
}

// 获取单独的知乎回答，爬虫解决静态页面
func getSingleAnswer(url string, data *telegraphGO.CreatePageRequest) (*telegraphGO.CreatePageRequest, error) {
	spider := colly.NewCollector()
	extensions.RandomUserAgent(spider)
	extensions.Referer(spider)

	spider.OnRequest(func(req *colly.Request) {
		fmt.Printf("fetching: %s\n", req.URL.String())
	})

	spider.OnResponse(func(res *colly.Response) {
		dom, err := goquery.NewDocumentFromReader(bytes.NewReader((res.Body)))
		errHandler("初始化goquery失败", err)

		// 回答标题
		dom.Find("div.QuestionHeader-content h1.QuestionHeader-title").Each(func(_ int, s *goquery.Selection) {
			data.Title = s.Text()
		})

		// 答主用户名
		dom.Find(".Card.AnswerCard .AuthorInfo-head .UserLink-link").Each(func(_ int, s *goquery.Selection) {
			data.AuthorName = s.Text()
		})

		// 回答内容
		dom.Find(".Card.AnswerCard .RichContent-inner").Each(func(_ int, s *goquery.Selection) {
			html, err := s.Html()
			errHandler("解析回答内容html失败", err)
			data.Data += html
		})

		// 文章发布时间
		dom.Find(".Card.AnswerCard .ContentItem-time a span").Each(func(_ int, s *goquery.Selection) {
			time := s.AttrOr("data-tooltip", "(未能获取到文章发布时间)")
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

// 获取单独的知乎专栏，爬虫解决静态页面
func getSingleZhuanLan(url string, data *telegraphGO.CreatePageRequest) (*telegraphGO.CreatePageRequest, error) {
	spider := colly.NewCollector()
	extensions.RandomUserAgent(spider)
	extensions.Referer(spider)

	spider.OnRequest(func(req *colly.Request) {
		fmt.Printf("fetching: %s\n", req.URL.String())
	})

	spider.OnResponse(func(res *colly.Response) {
		dom, err := goquery.NewDocumentFromReader(bytes.NewReader((res.Body)))
		errHandler("初始化goquery失败", err)

		// 专栏文章标题
		dom.Find("article .Post-Title").Each(func(_ int, s *goquery.Selection) {
			data.Title = s.Text()
		})

		// 专栏文章作者用户名
		dom.Find(".AuthorInfo .UserLink-link").Each(func(_ int, s *goquery.Selection) {
			data.AuthorName += s.Text()
		})

		// 专栏内容
		dom.Find("article > div.Post-RichTextContainer > .RichText").Each(func(_ int, s *goquery.Selection) {
			html, err := s.Html()
			errHandler("解析专栏内容html失败", err)
			data.Data += html
		})

		// 专栏文章发布时间
		dom.Find("article .ContentItem-time").Each(func(_ int, s *goquery.Selection) {
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

// 获取问题下面固定数量的回答
func getAllAnswers(fetchNum int32, questionID string) {
	c := colly.NewCollector()
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	// 这里还没处理好，不应该备份到本地
	// 0766：权限模式
	file, err := os.OpenFile("./answer.txt", os.O_RDWR|os.O_CREATE, 0766) //创建文件
	errHandler("打开文件失败", err)

	defer file.Close()

	total := int32(20) //知乎每次限制返回20个回答
	i := int32(0)      //记录是第几个回答

	c.OnRequest(func(request *colly.Request) {
		fmt.Printf("fetch --->%s\n", request.URL.String())
	})
	c.OnResponse(func(response *colly.Response) {

		var f interface{}
		json.Unmarshal(response.Body, &f) //反序列化
		// 找到改问题下的总回答数量是多少
		paging := f.(map[string]interface{})["paging"]
		totalNum := int32(paging.(map[string]interface{})["totals"].(float64))

		// 0 表示没有限制
		if fetchNum != 0 && fetchNum < totalNum {
			total = fetchNum
		} else {
			total = totalNum
		}

		// 找到当前url返回数据中的所有回答。
		data := f.(map[string]interface{})["data"]
		for _, v := range data.([]interface{}) {
			content := v.(map[string]interface{})["content"]
			file.Write([]byte(content.(string)))
		}
	})

	for ; i <= total; i += 20 {
		url := fmt.Sprintf("https://www.zhihu.com/api/v4/questions/%s/answers?include=data[*].is_normal,admin_closed_comment,reward_info,is_collapsed,annotation_action,annotation_detail,collapse_reason,is_sticky,collapsed_by,suggest_edit,comment_count,can_comment,content,editable_content,voteup_count,reshipment_settings,comment_permission,created_time,updated_time,review_info,relevant_info,question,excerpt,relationship.is_authorized,is_author,voting,is_thanked,is_nothelp,is_labeled,is_recognized,paid_info,paid_info_content;data[*].mark_infos[*].url;data[*].author.follower_count,badge[*].topics&offset=%d&limit=%d&sort_by=default", questionID, i, 20)
		c.Visit(url)
	}
}

func errHandler(msg string, err error) {
	if err != nil {
		fmt.Printf("%s - %s\n", msg, err)
		os.Exit(1)
	}
}
