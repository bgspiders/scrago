package spider

import (
	"scrago/request"
	"scrago/response"
)

// Spider 爬虫接口
type Spider interface {
	Name() string
	StartRequests() []*request.Request
	Parse(resp *response.Response) []interface{}
}

// BaseSpider 基础爬虫实现
type BaseSpider struct {
	name       string
	startUrls  []string
	allowedDomains []string
}

// NewBaseSpider 创建基础爬虫
func NewBaseSpider(name string, startUrls []string) *BaseSpider {
	return &BaseSpider{
		name:      name,
		startUrls: startUrls,
	}
}

// Name 返回爬虫名称
func (s *BaseSpider) Name() string {
	return s.name
}

// StartRequests 生成初始请求
func (s *BaseSpider) StartRequests() []*request.Request {
	requests := make([]*request.Request, 0, len(s.startUrls))
	for _, url := range s.startUrls {
		req := request.NewRequest("GET", url)
		requests = append(requests, req)
	}
	return requests
}

// Parse 默认解析方法
func (s *BaseSpider) Parse(resp *response.Response) []interface{} {
	// 默认实现，子类应该重写此方法
	return []interface{}{}
}

// ExampleSpider 示例爬虫
type ExampleSpider struct {
	*BaseSpider
}

// NewExampleSpider 创建示例爬虫
func NewExampleSpider() *ExampleSpider {
	base := NewBaseSpider("example", []string{
		"https://httpbin.org/html",
		"https://httpbin.org/json",
	})
	
	return &ExampleSpider{
		BaseSpider: base,
	}
}

// Parse 解析响应
func (s *ExampleSpider) Parse(resp *response.Response) []interface{} {
	results := make([]interface{}, 0)
	
	// 提取数据
	item := map[string]interface{}{
		"url":    resp.URL,
		"status": resp.StatusCode,
		"title":  resp.Selector().Find("title").Text(),
		"body_length": len(resp.Body),
	}
	
	results = append(results, item)
	
	// 可以生成新的请求
	// newReq := request.NewRequest("GET", "https://example.com/page2")
	// results = append(results, newReq)
	
	return results
}