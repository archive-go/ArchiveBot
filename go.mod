module archive-bot

go 1.14

require (
	github.com/MakeGolangGreat/archive-go v1.0.0
	github.com/MakeGolangGreat/telegraph-go v1.1.0
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/antchfx/xmlquery v1.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.9.0
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gocolly/colly v1.2.0
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202
)

// 本地开发telegraph-go时会用到
// replace github.com/MakeGolangGreat/archive-go => ../archive-go
// replace github.com/MakeGolangGreat/telegraph-go v1.1.0 => ../telegraph-go
