# Scrago 爬虫框架
# 本文档由ai生成，如有错误，请及时指正，现阶段不保证稳定运行，目前为开发阶段

<div align="center">

```
 _______ _______ _______ _______ _______ _______ 
(  ____ (  ____ (  ____ (  ___  (  ____ (  ___  )
| (    \| (    \| (    )| (   ) | (    \| (   ) |
| (_____| |     | (____)| (___) | |     | |   | |
(_____  | |     |     __|  ___  | | ____| |   | |
      ) | |     | (\ (  | (   ) | | \_  | |   | |
/\____) | (____/| ) \ \_| )   ( | (___) | (___) |
\_______(_______|/   \__|/     \(_______(_______)
                                                 
                  
```

**一个基于 Go 语言开发的高性能爬虫框架**

参考 Scrapy 的架构设计，支持并发、异步、分布式、动态加载插件等功能

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/bgspider/scrago)

</div>

## ✨ 核心特性

### 🚀 高性能架构
- **并发处理**: 基于 Go 协程的高并发爬取
- **异步 I/O**: 非阻塞网络请求处理
- **内存优化**: 智能内存管理和垃圾回收
- **负载均衡**: 自动请求分发和资源调度

### 🔧 灵活扩展
- **插件化设计**: 支持自定义中间件和管道
- **模块化架构**: 松耦合组件设计
- **热插拔**: 运行时动态加载插件
- **API 友好**: 清晰的扩展接口

### 🎯 强大的数据提取
- **CSS 选择器**: 类似 jQuery 的元素选择
- **XPath 表达式**: 强大的 XML 路径语言
- **正则表达式**: 灵活的文本模式匹配
- **JSON 解析**: 原生 JSON 数据处理

### 📊 多样化输出
- **格式支持**: JSON、CSV、XML、YAML
- **存储后端**: 本地文件、FTP、S3、数据库
- **实时流**: 支持数据流式处理
- **批量处理**: 高效的批量数据操作

### 🛡️ 反爬虫对抗
- **智能延迟**: 随机化请求间隔
- **代理轮换**: 自动代理池管理
- **User-Agent**: 浏览器标识轮换
- **会话管理**: Cookie 和会话保持

## 📊 数据管道系统

### 内置管道

| 管道类型 | 功能描述 | 配置示例 |
|---------|----------|----------|
| **JSONPipeline** | JSON 格式输出 | `OUTPUT_FORMAT: "json", OUTPUT_FILE: "data.json"` |
| **CSVPipeline** | CSV 格式输出 | `OUTPUT_FORMAT: "csv", CSV_DELIMITER: ","` |
| **XMLPipeline** | XML 格式输出 | `OUTPUT_FORMAT: "xml", XML_ROOT: "items"` |
| **DatabasePipeline** | 数据库存储 | `DB_URL: "mysql://user:pass@host/db"` |
| **FilePipeline** | 文件下载管道 | `FILES_STORE: "./downloads"` |
| **ImagePipeline** | 图片下载管道 | `IMAGES_STORE: "./images", IMAGES_THUMBS: true` |
| **DuplicatesPipeline** | 去重处理 | `DUPEFILTER_CLASS: "RFPDupeFilter"` |
| **ValidationPipeline** | 数据验证 | `VALIDATION_RULES: "rules.json"` |

### 自定义管道

```go
// 数据清洗管道
type DataCleaningPipeline struct {
    stats *stats.Stats
}

func (p *DataCleaningPipeline) ProcessItem(item interface{}, spider Spider) (interface{}, error) {
    // 类型断言
    data, ok := item.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid item type")
    }
    
    // 数据清洗
    if title, exists := data["title"]; exists {
        // 清理标题
        data["title"] = strings.TrimSpace(title.(string))
        data["title"] = regexp.MustCompile(`\s+`).ReplaceAllString(data["title"].(string), " ")
    }
    
    // 数据验证
    if err := p.validateItem(data); err != nil {
        p.stats.IncCounter("pipeline/validation_failed")
        return nil, err
    }
    
    // 数据转换
    data["processed_at"] = time.Now().Unix()
    data["spider_name"] = spider.Name()
    
    p.stats.IncCounter("pipeline/items_processed")
    return data, nil
}

func (p *DataCleaningPipeline) Open(spider Spider) error {
    p.stats = stats.GetStats()
    log.Printf("DataCleaningPipeline opened for spider: %s", spider.Name())
    return nil
}

func (p *DataCleaningPipeline) Close(spider Spider) error {
    processed := p.stats.GetCounter("pipeline/items_processed")
    failed := p.stats.GetCounter("pipeline/validation_failed")
    log.Printf("DataCleaningPipeline closed. Processed: %d, Failed: %d", processed, failed)
    return nil
}

// 数据库存储管道
type DatabasePipeline struct {
    db     *sql.DB
    stmt   *sql.Stmt
    config *DatabaseConfig
}

func (p *DatabasePipeline) ProcessItem(item interface{}, spider Spider) (interface{}, error) {
    data := item.(map[string]interface{})
    
    // 执行插入操作
    _, err := p.stmt.Exec(
        data["title"],
        data["content"], 
        data["url"],
        data["created_at"],
    )
    
    if err != nil {
        return nil, fmt.Errorf("database insert failed: %v", err)
    }
    
    return item, nil
}

func (p *DatabasePipeline) Open(spider Spider) error {
    // 连接数据库
    db, err := sql.Open(p.config.Driver, p.config.DSN)
    if err != nil {
        return err
    }
    
    p.db = db
    
    // 准备语句
    p.stmt, err = db.Prepare(`
        INSERT INTO items (title, content, url, created_at) 
        VALUES (?, ?, ?, ?)
    `)
    
    return err
}

func (p *DatabasePipeline) Close(spider Spider) error {
    if p.stmt != nil {
        p.stmt.Close()
    }
    if p.db != nil {
        p.db.Close()
    }
    return nil
}
```

### 管道配置

```go
// 管道注册和配置
settings := &settings.Settings{
    Pipelines: []string{
        "validation",     // 数据验证
        "cleaning",       // 数据清洗
        "duplicates",     // 去重处理
        "database",       // 数据库存储
        "json",          // JSON 输出
    },
    
    // 管道优先级（数字越小优先级越高）
    PipelinePriority: map[string]int{
        "validation": 100,
        "cleaning":   200,
        "duplicates": 300,
        "database":   400,
        "json":       500,
    },
}
```

## 🔌 扩展系统

### 内置扩展

| 扩展名称 | 功能描述 | 配置选项 |
|---------|----------|----------|
| **StatsExtension** | 统计信息收集 | `STATS_CLASS: "MemoryStats"` |
| **LogStatsExtension** | 日志统计输出 | `LOGSTATS_INTERVAL: 60` |
| **MemoryUsageExtension** | 内存使用监控 | `MEMUSAGE_LIMIT_MB: 1024` |
| **SpiderStateExtension** | 爬虫状态管理 | `SPIDER_STATE_ENABLED: true` |
| **AutoThrottleExtension** | 自动限速 | `AUTOTHROTTLE_ENABLED: true` |
| **CloseDomainExtension** | 域名关闭检测 | `CLOSEDOMAIN_ENABLED: true` |

### 自定义扩展

```go
// 性能监控扩展
type PerformanceExtension struct {
    startTime    time.Time
    requestCount int64
    errorCount   int64
    stats        *stats.Stats
}

func (e *PerformanceExtension) SpiderOpened(spider Spider) {
    e.startTime = time.Now()
    e.stats = stats.GetStats()
    log.Printf("Performance monitoring started for spider: %s", spider.Name())
}

func (e *PerformanceExtension) SpiderClosed(spider Spider, reason string) {
    duration := time.Since(e.startTime)
    requests := e.stats.GetCounter("downloader/request_count")
    errors := e.stats.GetCounter("downloader/exception_count")
    
    // 计算性能指标
    rps := float64(requests) / duration.Seconds()
    errorRate := float64(errors) / float64(requests) * 100
    
    log.Printf("Spider %s finished. Duration: %v, RPS: %.2f, Error Rate: %.2f%%", 
        spider.Name(), duration, rps, errorRate)
    
    // 发送性能报告
    e.sendPerformanceReport(spider.Name(), duration, rps, errorRate)
}

func (e *PerformanceExtension) RequestScheduled(request *http.Request, spider Spider) {
    atomic.AddInt64(&e.requestCount, 1)
}

func (e *PerformanceExtension) RequestDropped(request *http.Request, spider Spider, reason string) {
    atomic.AddInt64(&e.errorCount, 1)
}

// 邮件通知扩展
type EmailNotificationExtension struct {
    config *EmailConfig
}

func (e *EmailNotificationExtension) SpiderOpened(spider Spider) {
    e.sendEmail(
        "Spider Started", 
        fmt.Sprintf("Spider %s has started crawling", spider.Name()),
    )
}

func (e *EmailNotificationExtension) SpiderClosed(spider Spider, reason string) {
    stats := stats.GetStats()
    itemCount := stats.GetCounter("item_scraped_count")
    
    message := fmt.Sprintf(
        "Spider %s finished with reason: %s\nItems scraped: %d",
        spider.Name(), reason, itemCount,
    )
    
    e.sendEmail("Spider Finished", message)
}

// 扩展注册
func init() {
    extensions.Register("performance", &PerformanceExtension{})
    extensions.Register("email_notification", &EmailNotificationExtension{})
}
```

## 🚀 快速开始

### 📦 安装

#### 方式一：使用 Go Modules（推荐）

```bash
# 克隆项目
git clone https://github.com/bgspider/scrago.git
cd scrago

# 安装依赖
go mod tidy

# 构建项目
go build -o scrago ./cmd/scrago
```

#### 方式二：直接下载(自己编译，作者没有编译其他版本)

```bash
# 下载预编译二进制文件
wget https://github.com/bgspider/scrago/releases/latest/download/scrago-linux-amd64
chmod +x scrago-linux-amd64
mv scrago-linux-amd64 /usr/local/bin/scrago
```

### 🎯 命令行工具

Scrago 提供了强大的命令行工具，让您快速上手：

```bash
# 查看帮助
scrago --help

# 查看可用爬虫
scrago list

# 运行爬虫
scrago crawl douban_movie

# 创建新项目
scrago startproject myproject

# 生成新爬虫
scrago genspider quotes quotes.toscrape.com
```

### 💻 基础使用

#### 简单示例

```go
package main

import (
    "fmt"
    "scrago/engine"
    "scrago/spider"
    "scrago/pipeline"
    "scrago/middleware"
)

func main() {
    // 创建爬虫引擎
    e := engine.NewEngine()
    
    // 添加中间件
    e.AddMiddleware(middleware.NewUserAgentMiddleware())
    e.AddMiddleware(middleware.NewDelayMiddleware())
    
    // 添加数据管道
    e.AddPipeline(pipeline.NewConsolePipeline())
    e.AddPipeline(pipeline.NewJSONPipeline("output.json"))
    
    // 设置并发数
    e.SetConcurrency(16)
    
    // 创建并运行爬虫
    s := spider.NewExampleSpider()
    
    fmt.Println("🕷️ 开始爬取...")
    e.Run(s)
    fmt.Println("✅ 爬取完成！")
}
```

#### 自定义爬虫

创建一个新闻爬虫示例：

```go
package main

import (
    "fmt"
    "strings"
    "scrago/spider"
    "scrago/response"
    "scrago/request"
)

// NewsItem 新闻数据结构
type NewsItem struct {
    Title       string `json:"title"`
    Content     string `json:"content"`
    Author      string `json:"author"`
    PublishTime string `json:"publish_time"`
    URL         string `json:"url"`
    Tags        []string `json:"tags"`
}

// NewsSpider 新闻爬虫
type NewsSpider struct {
    *spider.BaseSpider
}

// NewNewsSpider 创建新闻爬虫
func NewNewsSpider() *NewsSpider {
    base := spider.NewBaseSpider("news", []string{
        "https://news.example.com",
        "https://tech.example.com",
    })
    return &NewsSpider{BaseSpider: base}
}

// Parse 解析首页，提取文章链接
func (s *NewsSpider) Parse(resp *response.Response) []interface{} {
    results := make([]interface{}, 0)
    
    // 提取文章链接
    articleLinks := resp.CSS("article a.title").Attrs("href")
    
    for _, link := range articleLinks {
        // 创建新请求，指定回调函数
        req := resp.Follow(link)
        req.Callback = s.ParseArticle
        results = append(results, req)
    }
    
    // 处理分页
    nextPage := resp.CSS("a.next-page").Attr("href")
    if nextPage != "" {
        nextReq := resp.Follow(nextPage)
        nextReq.Callback = s.Parse
        results = append(results, nextReq)
    }
    
    return results
}

// ParseArticle 解析文章详情页
func (s *NewsSpider) ParseArticle(resp *response.Response) []interface{} {
    results := make([]interface{}, 0)
    
    // 提取文章数据
    item := &NewsItem{
        Title:       resp.CSS("h1.article-title").Text(),
        Content:     resp.CSS("div.article-content").Text(),
        Author:      resp.CSS("span.author").Text(),
        PublishTime: resp.CSS("time.publish-date").Attr("datetime"),
        URL:         resp.URL,
    }
    
    // 提取标签
    tagElements := resp.CSS("div.tags a")
    for _, tag := range tagElements.Texts() {
        item.Tags = append(item.Tags, strings.TrimSpace(tag))
    }
    
    // 数据验证
    if item.Title != "" && item.Content != "" {
        results = append(results, item)
        fmt.Printf("📰 提取文章: %s\n", item.Title)
    }
    
    return results
}

// 使用示例
func main() {
    e := engine.NewEngine()
    
    // 配置引擎
    e.SetConcurrency(8)
    e.SetDownloadDelay(time.Second * 2)
    
    // 添加管道
    e.AddPipeline(pipeline.NewJSONPipeline("news.json"))
    
    // 运行爬虫
    spider := NewNewsSpider()
    e.Run(spider)
}
```

## 🏗️ 核心架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        Scrago 架构                          │
├─────────────────────────────────────────────────────────────┤
│  CLI Tool  │  Web UI  │  API Server  │  Monitoring        │
├─────────────────────────────────────────────────────────────┤
│                      Engine (引擎)                          │
├─────────────────────────────────────────────────────────────┤
│ Scheduler │ Downloader │ Middleware │ Pipeline │ Spider    │
├─────────────────────────────────────────────────────────────┤
│  Request  │  Response  │  Selector  │ Settings │ Stats     │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 核心组件

### 1. 🚀 爬虫引擎 (Engine)

引擎是整个框架的大脑，协调所有组件的工作：

```go
type Engine struct {
    scheduler   Scheduler     // 请求调度器
    downloader  Downloader    // 下载器
    middlewares []Middleware  // 中间件链
    pipelines   []Pipeline    // 数据管道
    settings    *Settings     // 配置管理
    stats       *Stats        // 统计信息
}

// 核心功能
func (e *Engine) Run(spider Spider) error
func (e *Engine) SetConcurrency(n int)
func (e *Engine) AddMiddleware(m Middleware)
func (e *Engine) AddPipeline(p Pipeline)
```

**主要职责：**
- 🔄 协调各组件工作流程
- 📊 管理并发和资源分配
- 📈 收集运行统计信息
- ⚙️ 处理配置和设置
- 🛡️ 异常处理和恢复

### 2. 🕷️ 爬虫 (Spider)

定义具体的爬取逻辑和数据提取规则：

```go
type Spider interface {
    Name() string                                    // 爬虫名称
    StartRequests() []*request.Request              // 初始请求
    Parse(resp *response.Response) []interface{}    // 解析响应
}

// 高级爬虫接口
type AdvancedSpider interface {
    Spider
    AllowedDomains() []string                       // 允许的域名
    CustomSettings() map[string]interface{}         // 自定义设置
    HandleError(req *request.Request, err error)    // 错误处理
}
```

### 3. 📋 调度器 (Scheduler)

智能的请求调度和队列管理：

```go
type Scheduler interface {
    Push(req *request.Request) error    // 添加请求
    Pop() (*request.Request, error)     // 获取请求
    Size() int                          // 队列大小
    Close() error                       // 关闭调度器
}

// 支持的调度策略
- FIFO: 先进先出 (默认)
- LIFO: 后进先出
- Priority: 优先级队列
- Random: 随机调度
- Custom: 自定义策略
```

### 4. 🌐 下载器 (Downloader)

高性能的 HTTP 客户端：

```go
type Downloader interface {
    Download(req *request.Request) (*response.Response, error)
    SetProxy(proxy string) error
    SetTimeout(timeout time.Duration)
    SetRetryTimes(times int)
}

// 特性支持
- HTTP/HTTPS 协议
- 代理服务器支持
- 自动重试机制
- Cookie 和会话管理
- 自定义请求头
- 文件上传下载
- 压缩传输 (gzip, deflate)
```

### 5. 🎯 选择器 (Selector)

强大的数据提取工具集：

```go
// CSS 选择器 - 类似 jQuery
titles := resp.CSS("h1.title").Texts()
links := resp.CSS("a[href]").Attrs("href")
first := resp.CSS("div.content").First().Text()

// XPath 表达式 - 强大的 XML 路径
nodes := resp.XPath("//div[@class='item']//text()").Texts()
attrs := resp.XPath("//a/@href").Strings()

// 正则表达式 - 灵活的模式匹配
emails := resp.Regex(`[\w\.-]+@[\w\.-]+\.\w+`).Strings()
phones := resp.Regex(`\d{3}-\d{3}-\d{4}`).Strings()

// JSON 数据提取
data := resp.JSON("data.items[*].name").Strings()
```

### 6. 🔌 中间件 (Middleware)

可插拔的请求/响应处理组件：

```go
type Middleware interface {
    ProcessRequest(req *request.Request) error
    ProcessResponse(resp *response.Response) error
}

// 内置中间件
- UserAgentMiddleware: 浏览器标识轮换
- ProxyMiddleware: 代理服务器轮换  
- DelayMiddleware: 智能延迟控制
- RetryMiddleware: 失败重试处理
- CacheMiddleware: 响应缓存管理
- CookieMiddleware: Cookie 自动管理
- AuthMiddleware: 身份认证处理
- RobotsTxtMiddleware: robots.txt 遵守
```

### 7. 📊 数据管道 (Pipeline)

数据处理和存储的流水线：

```go
type Pipeline interface {
    ProcessItem(item interface{}) (interface{}, error)
    Open() error
    Close() error
}

// 内置管道
- ConsolePipeline: 控制台输出
- JSONPipeline: JSON 文件存储
- CSVPipeline: CSV 表格存储
- XMLPipeline: XML 文档存储
- DatabasePipeline: 数据库存储
- FilterPipeline: 数据过滤清洗
- TransformPipeline: 数据格式转换
- ValidationPipeline: 数据验证检查
```

## ⚙️ 配置选项

### 基础配置

```go
// 创建配置
settings := &settings.Settings{
    // 🚀 性能配置
    Concurrency:        16,                     // 并发协程数
    DownloadDelay:      time.Second,            // 下载延迟
    RandomizeDelay:     true,                   // 随机化延迟 (0.5-1.5倍)
    AutoThrottle:       true,                   // 自动限速
    
    // 🌐 网络配置  
    UserAgent:          "Scrago/1.0",          // User-Agent
    Timeout:            30 * time.Second,       // 请求超时
    KeepAlive:          true,                   // 连接保持
    MaxIdleConns:       100,                    // 最大空闲连接
    
    // 🔄 重试配置
    RetryTimes:         3,                      // 重试次数
    RetryHTTPCodes:     []int{500, 502, 503},   // 重试状态码
    RetryDelay:         time.Second * 2,        // 重试延迟
    
    // 🛡️ 安全配置
    RobotsTxtObey:      true,                   // 遵守 robots.txt
    AllowedDomains:     []string{},             // 允许的域名
    DeniedDomains:      []string{},             // 禁止的域名
    
    // 📊 输出配置
    LogLevel:           "INFO",                 // 日志级别
    StatsEnabled:       true,                   // 统计信息
    OutputFormat:       "json",                 // 输出格式
}

// 应用配置
engine := engine.NewEngine()
engine.ApplySettings(settings)
```

### 高级配置

```go
// 代理配置
proxyConfig := &middleware.ProxyConfig{
    Proxies: []string{
        "http://proxy1:8080",
        "http://proxy2:8080",
        "socks5://proxy3:1080",
    },
    RotateMode: "random",  // random, round_robin, failover
}

// 缓存配置
cacheConfig := &middleware.CacheConfig{
    Enabled:    true,
    TTL:        time.Hour * 24,
    MaxSize:    1000,
    Storage:    "memory",  // memory, redis, file
}

// 数据库配置
dbConfig := &pipeline.DatabaseConfig{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "scrago",
    Username: "root",
    Password: "password",
}
```

## 📚 示例项目

### 🎯 内置爬虫示例

| 爬虫名称 | 目标网站 | 数据类型 | 运行命令 |
|---------|---------|---------|----------|
| **douban_movie** | 豆瓣电影 | 电影信息 | `scrago crawl douban_movie` |
| **quotes** | Quotes to Scrape | 名言警句 | `scrago crawl quotes` |
| **books** | Books to Scrape | 图书信息 | `scrago crawl books` |
| **news** | 新闻网站 | 新闻文章 | `scrago crawl news` |

### 🛠️ 自定义项目示例

```bash
# 创建新项目
scrago startproject myproject
cd myproject

# 生成爬虫
scrago genspider quotes quotes.toscrape.com
scrago genspider books books.toscrape.com

# 运行爬虫
scrago crawl quotes -o quotes.json
scrago crawl books -o books.csv -s CONCURRENT_REQUESTS=32
```

### 📊 实际应用场景

#### 🛒 电商数据监控
```bash
# 商品价格监控
scrago crawl ecommerce \
    --set DOWNLOAD_DELAY=2 \
    --set CONCURRENT_REQUESTS=8 \
    --output products.json \
    --log-level INFO
```

#### 📰 新闻内容聚合
```bash
# 多源新闻抓取
scrago crawl news \
    --set USER_AGENT_LIST=user_agents.txt \
    --set PROXY_LIST=proxies.txt \
    --output news.csv \
    --format csv
```

#### 📈 社交媒体分析
```bash
# 社交媒体数据
scrago crawl social \
    --set COOKIES_ENABLED=true \
    --set SESSION_PERSISTENCE=true \
    --output social_data.json \
    --pipeline database
```

## 🚀 运行示例

### 命令行工具

```bash
# 查看帮助信息
scrago --help

# 查看可用爬虫
scrago list

# 运行内置爬虫
scrago crawl douban_movie
scrago crawl quotes -o output.json
scrago crawl books -o books.csv -s CONCURRENT_REQUESTS=16

# 创建新项目
scrago startproject myproject

# 生成爬虫模板
scrago genspider myspider example.com

# 检查爬虫语法
scrago check myspider

# 查看爬虫统计
scrago stats douban_movie
```

### 程序化运行

```bash
# 直接运行 Go 程序
go run cmd/scrago/main.go crawl douban_movie
go run cmd/scrago/main.go list

# 编译后运行
go build -o scrago cmd/scrago/main.go
./scrago crawl douban_movie
```

## 🔧 中间件系统

### 内置中间件

| 中间件 | 功能描述 | 配置示例 |
|--------|----------|----------|
| **UserAgentMiddleware** | 随机 User-Agent 轮换 | `USER_AGENT_LIST: ["Chrome/91.0", "Firefox/89.0"]` |
| **ProxyMiddleware** | 代理服务器支持 | `PROXY_LIST: ["http://proxy1:8080"]` |
| **RetryMiddleware** | 智能重试机制 | `RETRY_TIMES: 3, RETRY_HTTP_CODES: [500, 502]` |
| **CacheMiddleware** | HTTP 缓存支持 | `CACHE_ENABLED: true, CACHE_TTL: "24h"` |
| **RobotsTxtMiddleware** | robots.txt 遵守 | `ROBOTSTXT_OBEY: true` |
| **CookieMiddleware** | Cookie 管理 | `COOKIES_ENABLED: true` |
| **CompressionMiddleware** | 响应压缩处理 | `COMPRESSION_ENABLED: true` |
| **AuthMiddleware** | 身份验证支持 | `AUTH_USERNAME: "user", AUTH_PASSWORD: "pass"` |

### 自定义中间件

```go
// 实现中间件接口
type CustomMiddleware struct {
    config *MiddlewareConfig
}

// 请求预处理
func (m *CustomMiddleware) ProcessRequest(req *http.Request, spider Spider) error {
    // 添加自定义请求头
    req.Header.Set("X-Custom-Header", "MyValue")
    req.Header.Set("X-Request-ID", generateRequestID())
    
    // 请求日志记录
    log.Printf("Processing request: %s", req.URL.String())
    
    return nil
}

// 响应后处理
func (m *CustomMiddleware) ProcessResponse(resp *http.Response, req *http.Request, spider Spider) (*http.Response, error) {
    // 响应状态检查
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
    }
    
    // 响应时间记录
    duration := time.Since(req.Context().Value("start_time").(time.Time))
    log.Printf("Response received in %v", duration)
    
    return resp, nil
}

// 异常处理
func (m *CustomMiddleware) ProcessException(err error, req *http.Request, spider Spider) error {
    // 错误日志记录
    log.Printf("Request failed: %s, Error: %v", req.URL.String(), err)
    
    // 可以选择忽略某些错误
    if isIgnorableError(err) {
        return nil
    }
    
    return err
}

// 注册中间件
func init() {
    middleware.Register("custom", &CustomMiddleware{})
}
```

### 中间件配置

```go
// 在设置中启用中间件
settings := &settings.Settings{
    Middlewares: []string{
        "user_agent",
        "proxy", 
        "retry",
        "cache",
        "custom",  // 自定义中间件
    },
    
    // 中间件优先级
    MiddlewarePriority: map[string]int{
        "user_agent": 100,
        "proxy":      200,
        "retry":      300,
        "cache":      400,
        "custom":     500,
    },
}
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

## 🤝 贡献指南

我们欢迎所有形式的贡献！无论是 bug 报告、功能请求、文档改进还是代码贡献。

### 🐛 报告 Bug

如果您发现了 bug，请创建一个 issue 并包含以下信息：

- **Bug 描述**: 清晰简洁的描述
- **复现步骤**: 详细的复现步骤
- **期望行为**: 您期望发生什么
- **实际行为**: 实际发生了什么
- **环境信息**: Go 版本、操作系统等
- **相关日志**: 错误日志或堆栈跟踪

### 💡 功能请求

对于新功能建议，请创建一个 issue 并说明：

- **功能描述**: 您希望添加的功能
- **使用场景**: 为什么需要这个功能
- **实现建议**: 如果有的话，提供实现思路

### 🔧 代码贡献

1. **Fork 项目**
   ```bash
   git clone https://github.com/bgspiders/scrago.git
   cd scrago
   ```

2. **创建功能分支**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **开发和测试**
   ```bash
   # 安装依赖
   go mod tidy
   
   # 运行测试
   go test ./...
   
   # 运行 linter
   golangci-lint run
   
   # 格式化代码
   go fmt ./...
   ```

4. **提交更改**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

5. **推送并创建 PR**
   ```bash
   git push origin feature/amazing-feature
   ```

### 📝 代码规范

- 遵循 Go 官方代码风格
- 添加适当的注释和文档
- 编写单元测试
- 确保所有测试通过
- 使用有意义的提交信息

### 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/spider

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📄 许可证

本项目采用 **MIT 许可证** - 查看 [LICENSE](LICENSE) 文件了解详情。

```
MIT License

Copyright (c) 2024 Scrago Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## 🙏 致谢

感谢所有为 Scrago 项目做出贡献的开发者！

- 特别感谢 [Scrapy](https://scrapy.org/) 项目提供的设计灵感
- 感谢 Go 社区提供的优秀工具和库
- 感谢所有提交 issue 和 PR 的贡献者

## 📞 联系我们

- **GitHub Issues**: [提交问题](https://github.com/bgspiders/scrago/issues)
- **讨论区**: [GitHub Discussions](https://github.com/bgspiders/scrago/discussions)
- **邮箱**: scrago@example.com

---

<div align="center">

**⭐ 如果这个项目对您有帮助，请给我们一个 Star！⭐**

[🏠 首页](https://github.com/bgspiders/scrago) • 
[📖 文档](https://scrago.readthedocs.io) • 
[🐛 报告 Bug](https://github.com/bgspiders/scrago/issues) • 
[💡 功能请求](https://github.com/bgspiders/scrago/issues)

</div>