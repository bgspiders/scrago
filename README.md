# Scrago

一个基于 Go 语言开发的高性能爬虫框架，参考 Scrapy 的架构设计，支持并发、异步、分布式、动态加载插件等功能。

## 特性

- 🚀 **高性能并发**: 支持多协程并发抓取
- 🔧 **可扩展架构**: 插件化设计，支持自定义中间件和管道
- 🎯 **多种选择器**: 支持 CSS 选择器、XPath 表达式和正则表达式
- 📊 **多格式输出**: 支持 JSON、CSV、XML 等多种输出格式
- 🌐 **多后端存储**: 支持本地文件系统、FTP、S3 等存储后端
- 🔄 **智能重试**: 内置重试机制和错误处理
- 🍪 **会话管理**: 支持 Cookie 和会话处理
- 🕷️ **反爬虫对抗**: 支持代理、User-Agent 轮换、延迟控制等
- 📈 **实时统计**: 内置爬取统计和监控功能

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 基础使用

```go
package main

import (
    "scrago/engine"
    "scrago/spider"
    "scrago/pipeline"
)

func main() {
    // 创建爬虫引擎
    e := engine.NewEngine()
    
    // 添加数据管道
    e.AddPipeline(pipeline.NewConsolePipeline())
    e.AddPipeline(pipeline.NewJSONPipeline("output.json"))
    
    // 创建并运行爬虫
    s := spider.NewExampleSpider()
    e.Run(s)
}
```

### 自定义爬虫

```go
type MySpider struct {
    *spider.BaseSpider
}

func NewMySpider() *MySpider {
    base := spider.NewBaseSpider("myspider", []string{
        "https://example.com",
    })
    return &MySpider{BaseSpider: base}
}

func (s *MySpider) Parse(resp *response.Response) []interface{} {
    results := make([]interface{}, 0)
    
    // 使用 CSS 选择器提取数据
    titles := resp.CSS("h1").Texts()
    for _, title := range titles {
        item := map[string]interface{}{
            "title": title,
            "url":   resp.URL,
        }
        results = append(results, item)
    }
    
    // 跟随链接
    links := resp.CSS("a").Attrs("href")
    for _, link := range links {
        newReq := resp.Follow(link)
        results = append(results, newReq)
    }
    
    return results
}
```

## 核心组件

### 1. 爬虫引擎 (Engine)

爬虫引擎是框架的核心，负责协调各个组件的工作：

- 调度器管理
- 下载器控制
- 中间件处理
- 数据管道处理
- 并发控制
- 统计信息收集

### 2. 爬虫 (Spider)

爬虫定义了如何抓取特定网站的逻辑：

```go
type Spider interface {
    Name() string
    StartRequests() []*request.Request
    Parse(resp *response.Response) []interface{}
}
```

### 3. 调度器 (Scheduler)

支持多种调度策略：

- **FIFO**: 先进先出队列
- **LIFO**: 后进先出栈
- **Priority**: 优先级队列

### 4. 下载器 (Downloader)

负责执行 HTTP 请求：

- 支持代理设置
- 自动重试机制
- SSL/TLS 支持
- Cookie 管理
- 超时控制

### 5. 选择器 (Selector)

强大的数据提取工具：

```go
// CSS 选择器
titles := resp.CSS("h1.title").Texts()

// XPath 表达式
links := resp.XPath("//a[@class='link']").Attrs("href")

// 正则表达式
emails := resp.Selector().Regex(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
```

### 6. 中间件 (Middleware)

可插拔的请求/响应处理组件：

- **UserAgentMiddleware**: User-Agent 轮换
- **ProxyMiddleware**: 代理轮换
- **DelayMiddleware**: 请求延迟
- **RetryMiddleware**: 重试处理
- **CookieMiddleware**: Cookie 管理

### 7. 数据管道 (Pipeline)

数据处理和存储组件：

- **ConsolePipeline**: 控制台输出
- **JSONPipeline**: JSON 文件存储
- **CSVPipeline**: CSV 文件存储
- **XMLPipeline**: XML 文件存储
- **FilterPipeline**: 数据过滤
- **TransformPipeline**: 数据转换

## 配置选项

```go
settings := &engine.Settings{
    Concurrency:      16,                    // 并发数
    DownloadDelay:    time.Second,           // 下载延迟
    RandomizeDelay:   true,                  // 随机化延迟
    UserAgent:        "Scrago/1.0",       // User-Agent
    RobotsTxtObey:    true,                  // 遵守 robots.txt
    AutoThrottle:     true,                  // 自动限速
    RetryTimes:       3,                     // 重试次数
    RetryHTTPCodes:   []int{500, 502, 503},  // 重试状态码
}
```

## 示例项目

查看 `examples/` 目录中的完整示例代码：

### 🎯 可用示例

| 示例 | 功能描述 | 运行命令 |
|------|----------|----------|
| **basic_spider.go** | 基础爬虫功能演示 | `go run examples/basic_spider.go` |
| **news_spider.go** | 新闻数据多网站抓取 | `go run examples/news_spider.go` |
| **ecommerce_spider.go** | 电商商品信息抓取 | `go run examples/ecommerce_spider.go` |
| **social_spider.go** | 社交媒体内容抓取 | `go run examples/social_spider.go` |
| **api_spider.go** | API接口数据抓取 | `go run examples/api_spider.go` |
| **run_examples.go** | 完整引擎配置示例 | `go run examples/run_examples.go` |

### 📚 示例特性

- **基础爬虫**: HTML解析、CSS选择器、数据提取
- **新闻爬虫**: 多网站处理、分页跟踪、内容分类
- **电商爬虫**: 商品信息、价格监控、分类导航
- **社交爬虫**: 用户内容、评论数据、关系链
- **API爬虫**: JSON数据、接口调用、数据分析
- **完整配置**: 中间件、管道、引擎设置

详细说明请查看 [examples/README.md](examples/README.md)

## 运行示例

```bash
# 安装Go语言环境
./install_go.sh

# 下载依赖包
go mod tidy

# 运行基础示例
go run examples/basic_spider.go

# 运行电商爬虫
go run examples/ecommerce_spider.go

# 运行API爬虫
go run examples/api_spider.go

# 运行主程序
go run main.go
```

## 项目结构

```
scrago/
├── engine/          # 爬虫引擎
├── spider/          # 爬虫基类和示例
├── request/         # 请求对象
├── response/        # 响应对象
├── selector/        # 选择器和数据提取
├── downloader/      # 下载器
├── scheduler/       # 调度器
├── pipeline/        # 数据管道
├── middleware/      # 中间件
├── examples/        # 示例代码
├── main.go          # 主入口
├── go.mod           # 依赖管理
└── README.md        # 项目文档
```

## 依赖包

- `github.com/PuerkitoBio/goquery`: HTML 解析和 CSS 选择器
- `github.com/antchfx/htmlquery`: HTML XPath 支持
- `github.com/antchfx/xmlquery`: XML XPath 支持
- `github.com/antchfx/xpath`: XPath 表达式引擎
- `golang.org/x/net`: 网络相关工具

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

MIT License