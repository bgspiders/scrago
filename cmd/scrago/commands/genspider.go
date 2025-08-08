package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenSpiderCommand 处理 genspider 命令
func GenSpiderCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("❌ 错误: 请指定爬虫名称和域名")
		fmt.Println("用法: scrago genspider <spider_name> <domain>")
		fmt.Println("示例: scrago genspider quotes quotes.toscrape.com")
		return
	}

	spiderName := args[0]
	domain := args[1]
	
	// 验证爬虫名称
	if !isValidSpiderName(spiderName) {
		fmt.Printf("❌ 错误: 无效的爬虫名称 '%s'\n", spiderName)
		fmt.Println("爬虫名称只能包含字母、数字和下划线，且不能以数字开头")
		return
	}

	// 确保 spiders 目录存在
	spidersDir := "spiders"
	if err := os.MkdirAll(spidersDir, 0755); err != nil {
		fmt.Printf("❌ 创建 spiders 目录失败: %v\n", err)
		return
	}

	// 生成文件路径
	fileName := fmt.Sprintf("%s_spider.go", spiderName)
	filePath := filepath.Join(spidersDir, fileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		fmt.Printf("❌ 错误: 爬虫文件 '%s' 已存在\n", filePath)
		return
	}

	fmt.Printf("🚀 创建新爬虫: %s (域名: %s)\n", spiderName, domain)

	// 生成爬虫代码
	spiderCode := generateSpiderCode(spiderName, domain)

	// 写入文件
	if err := os.WriteFile(filePath, []byte(spiderCode), 0644); err != nil {
		fmt.Printf("❌ 创建爬虫文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 爬虫 '%s' 创建成功！\n", spiderName)
	fmt.Printf("📁 文件位置: %s\n\n", filePath)
	
	fmt.Println("🎯 下一步:")
	fmt.Printf("  1. 编辑 %s 实现你的爬取逻辑\n", filePath)
	fmt.Printf("  2. 运行爬虫: scrago crawl %s\n", spiderName)
}

// isValidSpiderName 验证爬虫名称
func isValidSpiderName(name string) bool {
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

// generateSpiderCode 生成爬虫代码
func generateSpiderCode(spiderName, domain string) string {
	structName := strings.Title(spiderName) + "Spider"
	startURL := fmt.Sprintf("https://%s", domain)
	
	return fmt.Sprintf(`package spiders

import (
	"fmt"
	"scrago/request"
	"scrago/response"
	"scrago/selector"
	"scrago/settings"
	"scrago/spider"
	"strings"
)

// %sItem 数据结构
type %sItem struct {
	Title string ` + "`" + `json:"title"` + "`" + `
	URL   string ` + "`" + `json:"url"` + "`" + `
	// TODO: 添加更多字段
}

// %s 爬虫
type %s struct {
	*spider.BaseSpider
	settings *settings.Settings
}

// New%s 创建新的爬虫实例
func New%s(settings *settings.Settings) *%s {
	startURLs := []string{
		"%s",
		// TODO: 添加更多起始URL
	}

	base := spider.NewBaseSpider("%s", startURLs)

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
		req.SetHeader("User-Agent", "Mozilla/5.0 (compatible; Go-Scrapy/1.0)")
		req.SetMeta("callback", "parse")
		requests = append(requests, req)
	}

	fmt.Printf("🚀 %s爬虫：生成了 %%d 个初始请求\n", len(requests))
	return requests
}

// Parse 解析响应
func (s *%s) Parse(resp *response.Response) []interface{} {
	if resp.StatusCode != 200 {
		fmt.Printf("❌ 请求失败，状态码: %%d, URL: %%s\n", resp.StatusCode, resp.URL)
		return []interface{}{}
	}

	sel := selector.NewSelector(string(resp.Body))
	var results []interface{}

	// TODO: 实现你的解析逻辑
	// 示例：提取所有链接
	links := sel.CSS("a").Attrs("href")
	for _, link := range links {
		if strings.HasPrefix(link, "http") {
			item := &%sItem{
				Title: "示例标题", // TODO: 提取实际标题
				URL:   link,
			}
			results = append(results, item)
		}
	}

	fmt.Printf("📄 从 %%s 提取了 %%d 个项目\n", resp.URL, len(results))
	return results
}
`, structName, structName, structName, structName, structName, structName, structName, startURL, spiderName, structName, structName, structName, structName)
}