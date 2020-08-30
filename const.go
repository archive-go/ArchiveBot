package main

import "github.com/MakeGolangGreat/telegraph-go"

const (
	projectLink = "https://github.com/MakeGolangGreat/ArchiveBot"
	projectName = "MakeGolangGreat"
	projectDesc = `<blockquote>本页面由开源程序「<a href="https://github.com/MakeGolangGreat/ArchiveBot">ArchiveBot</a>」生成，内容来自网络。</blockquote>`
	apiEndpoint = "https://api.telegram.org/bot%s/%s"
)

var attachInfo = &telegraph.NodeElement{
	Tag: "p",
	Children: []telegraph.Node{
		telegraph.NodeElement{
			Tag: "br",
		},
		telegraph.NodeElement{
			Tag: "blockquote",
			Children: []telegraph.Node{
				telegraph.NodeElement{
					Tag:      "strong",
					Children: []telegraph.Node{"本页面由 "},
				},
				telegraph.NodeElement{
					Tag: "a",
					Attrs: map[string]string{
						"href": "https://t.me/beifenbot",
					},
					Children: []telegraph.Node{"@beifenbot"},
				},
				telegraph.NodeElement{
					Tag:      "strong",
					Children: []telegraph.Node{" 备份，"},
				},
				telegraph.NodeElement{
					Tag:      "strong",
					Children: []telegraph.Node{"代码开源："},
				},
				telegraph.NodeElement{
					Tag: "a",
					Attrs: map[string]string{
						"href": "https://github.com/MakeGolangGreat/ArchiveBot",
					},
					Children: []telegraph.Node{"https://github.com/MakeGolangGreat/ArchiveBot"},
				},
			},
		},
	},
}
