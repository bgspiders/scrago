package response

import (
	"bytes"
	"encoding/json"
	"scrago/request"
	"scrago/selector"
	"net/http"
	"net/url"
	"strings"
)

// Response 响应结构
type Response struct {
	URL        string
	StatusCode int
	Headers    http.Header
	Body       []byte
	Request    *request.Request
	Meta       map[string]interface{}
	
	// 编码信息
	Encoding string
	
	// 缓存的选择器
	selector *selector.Selector
}

// NewResponse 创建新响应
func NewResponse(url string, statusCode int, headers http.Header, body []byte, req *request.Request) *Response {
	return &Response{
		URL:        url,
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
		Request:    req,
		Meta:       make(map[string]interface{}),
		Encoding:   "utf-8",
	}
}

// Text 获取响应文本
func (r *Response) Text() string {
	return string(r.Body)
}

// JSON 解析JSON响应
func (r *Response) JSON() (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(r.Body, &result)
	return result, err
}

// Selector 获取选择器
func (r *Response) Selector() *selector.Selector {
	if r.selector == nil {
		r.selector = selector.NewSelector(string(r.Body))
	}
	return r.selector
}

// CSS 使用CSS选择器
func (r *Response) CSS(cssSelector string) *selector.Selection {
	return r.Selector().CSS(cssSelector)
}

// XPath 使用XPath选择器
func (r *Response) XPath(xpathExpr string) *selector.Selection {
	return r.Selector().XPath(xpathExpr)
}

// Follow 跟随链接
func (r *Response) Follow(href string) *request.Request {
	absoluteURL := r.urljoin(href)
	return request.NewRequest("GET", absoluteURL)
}

// FollowAll 跟随所有链接
func (r *Response) FollowAll(hrefs []string) []*request.Request {
	requests := make([]*request.Request, 0, len(hrefs))
	for _, href := range hrefs {
		req := r.Follow(href)
		requests = append(requests, req)
	}
	return requests
}

// urljoin 合并URL
func (r *Response) urljoin(href string) string {
	base, err := url.Parse(r.URL)
	if err != nil {
		return href
	}
	
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	
	return base.ResolveReference(ref).String()
}

// GetMeta 获取元数据
func (r *Response) GetMeta(key string) interface{} {
	return r.Meta[key]
}

// SetMeta 设置元数据
func (r *Response) SetMeta(key string, value interface{}) {
	r.Meta[key] = value
}

// Copy 复制响应
func (r *Response) Copy() *Response {
	newResp := &Response{
		URL:        r.URL,
		StatusCode: r.StatusCode,
		Headers:    make(http.Header),
		Body:       make([]byte, len(r.Body)),
		Request:    r.Request,
		Meta:       make(map[string]interface{}),
		Encoding:   r.Encoding,
	}
	
	// 复制Headers
	for k, v := range r.Headers {
		newResp.Headers[k] = v
	}
	
	// 复制Body
	copy(newResp.Body, r.Body)
	
	// 复制Meta
	for k, v := range r.Meta {
		newResp.Meta[k] = v
	}
	
	return newResp
}

// String 返回响应的字符串表示
func (r *Response) String() string {
	return r.URL
}

// IsHTML 检查是否为HTML响应
func (r *Response) IsHTML() bool {
	contentType := r.Headers.Get("Content-Type")
	return strings.Contains(strings.ToLower(contentType), "text/html")
}

// IsJSON 检查是否为JSON响应
func (r *Response) IsJSON() bool {
	contentType := r.Headers.Get("Content-Type")
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// IsXML 检查是否为XML响应
func (r *Response) IsXML() bool {
	contentType := r.Headers.Get("Content-Type")
	return strings.Contains(strings.ToLower(contentType), "text/xml") ||
		   strings.Contains(strings.ToLower(contentType), "application/xml")
}

// Reader 获取Body的Reader
func (r *Response) Reader() *bytes.Reader {
	return bytes.NewReader(r.Body)
}