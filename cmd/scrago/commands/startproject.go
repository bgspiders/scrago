package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StartProjectCommand 处理 startproject 命令
func StartProjectCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ 错误: 请指定项目名称")
		fmt.Println("用法: scrago startproject <project_name>")
		fmt.Println("示例: scrago startproject myspider")
		return
	}

	projectName := args[0]
	
	// 验证项目名称
	if !isValidProjectName(projectName) {
		fmt.Printf("❌ 错误: 无效的项目名称 '%s'\n", projectName)
		fmt.Println("项目名称只能包含字母、数字和下划线，且不能以数字开头")
		return
	}

	// 检查目录是否已存在
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		fmt.Printf("❌ 错误: 目录 '%s' 已存在\n", projectName)
		return
	}

	fmt.Printf("🚀 创建新的爬虫项目: %s\n", projectName)

	// 创建项目结构
	if err := createProjectStructure(projectName); err != nil {
		fmt.Printf("❌ 创建项目失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 项目 '%s' 创建成功！\n\n", projectName)
	fmt.Println("📁 项目结构:")
	fmt.Printf(`%s/
├── go.mod
├── main.go
├── scrago.json
├── spiders/
│   └── example_spider.go
└── README.md

`, projectName)

	fmt.Println("🎯 下一步:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  scrago crawl example")
}

// isValidProjectName 验证项目名称
func isValidProjectName(name string) bool {
	if len(name) == 0 {
		return false
	}
	
	// 不能以数字开头
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}
	
	// 只能包含字母、数字和下划线
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_') {
			return false
		}
	}
	
	return true
}

// createProjectStructure 创建项目结构
func createProjectStructure(projectName string) error {
	// 创建主目录
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return err
	}

	// 创建子目录
	dirs := []string{
		filepath.Join(projectName, "spiders"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// 创建文件
	files := map[string]string{
		filepath.Join(projectName, "go.mod"):     generateGoMod(projectName),
		filepath.Join(projectName, "main.go"):    generateMainGo(projectName),
		filepath.Join(projectName, "scrago.json"): generateScrapyConfig(),
		filepath.Join(projectName, "spiders", "example_spider.go"): generateExampleSpider(projectName),
		filepath.Join(projectName, "README.md"):  generateReadme(projectName),
	}

	for filePath, content := range files {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// generateGoMod 生成 go.mod 文件
func generateGoMod(projectName string) string {
	return fmt.Sprintf(`module %s

go 1.21

require (
	scrago v0.0.0
)

replace scrago => ../scrago
`, projectName)
}

// generateMainGo 生成 main.go 文件
func generateMainGo(projectName string) string {
	return `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Welcome to %s!\n", os.Args[0])
	fmt.Println("Use 'scrago crawl <spider>' to run a spider")
	fmt.Println("Use 'scrago list' to see available spiders")
}
`
}

// generateScrapyConfig 生成 scrapy.json 配置文件
func generateScrapyConfig() string {
	return `{
  "bot_name": "MySpider",
  "user_agent": "MySpider (+http://www.yourdomain.com)",
  "concurrent_requests": 16,
  "download_delay": 1.0,
  "randomize_download_delay": true,
  "downloader_middlewares": {
    "UserAgentMiddleware": 100,
    "DelayMiddleware": 200
  },
  "item_pipelines": {
    "JSONPipeline": 100
  },
  "feeds_export": {
    "output.json": {
      "format": "json",
      "encoding": "utf-8"
    }
  }
}
`
}

// generateExampleSpider 生成示例爬虫
func generateExampleSpider(projectName string) string {
	spiderName := strings.Title(projectName) + "Spider"
	return fmt.Sprintf(`package spiders

import (
	"scrago/request"
	"scrago/response"
	"scrago/spider"
	"scrago/settings"
)

// %s 示例爬虫
type %s struct {
	*spider.BaseSpider
	settings *settings.Settings
}

// New%s 创建新的爬虫实例
func New%s(settings *settings.Settings) *%s {
	startURLs := []string{
		"https://example.com",
	}

	base := spider.NewBaseSpider("example", startURLs)

	return &%s{
		BaseSpider: base,
		settings:   settings,
	}
}

// StartRequests 生成初始请求
func (s *%s) StartRequests() []*request.Request {
	var requests []*request.Request

	for _, url := range s.StartURLs {
		req := request.NewRequest("GET", url)
		req.SetMeta("callback", "parse")
		requests = append(requests, req)
	}

	return requests
}

// Parse 解析响应
func (s *%s) Parse(resp *response.Response) []interface{} {
	// TODO: 实现你的解析逻辑
	return []interface{}{}
}
`, spiderName, spiderName, spiderName, spiderName, spiderName, spiderName, spiderName, spiderName)
}

// generateReadme 生成 README.md 文件
func generateReadme(projectName string) string {
	template := `# %s

这是一个使用 Scrago 框架创建的爬虫项目。

## 快速开始

1. 安装依赖:
   ` + "`" + `bash
   go mod tidy
   ` + "`" + `

2. 运行示例爬虫:
   ` + "`" + `bash
   scrago crawl example
   ` + "`" + `

3. 查看可用爬虫:
   ` + "`" + `bash
   scrago list
   ` + "`" + `

## 项目结构

- spiders/ - 爬虫定义
- scrago.json - 配置文件
- main.go - 主入口文件

## 创建新爬虫

` + "`" + `bash
scrago genspider myspider example.com
` + "`" + `

## 配置

编辑 scrago.json 文件来修改爬虫配置。

更多信息请参考 Scrago 文档。
`
	return fmt.Sprintf(template, projectName)
}