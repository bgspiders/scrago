package request

import (
	"net/http"
	"net/url"
	"time"
)

// Request 请求结构
type Request struct {
	Method   string
	URL      string
	Headers  http.Header
	Body     []byte
	Meta     map[string]interface{}
	Cookies  []*http.Cookie
	Priority int
	
	// 重试相关
	RetryTimes int
	DontRetry  bool
	
	// 回调函数
	Callback string
	
	// 代理设置
	Proxy string
	
	// 超时设置
	Timeout time.Duration
	
	// 是否跟随重定向
	DontRedirect bool
}

// NewRequest 创建新请求
func NewRequest(method, rawURL string) *Request {
	return &Request{
		Method:   method,
		URL:      rawURL,
		Headers:  make(http.Header),
		Meta:     make(map[string]interface{}),
		Priority: 0,
		Timeout:  30 * time.Second,
	}
}

// SetHeader 设置请求头
func (r *Request) SetHeader(key, value string) *Request {
	r.Headers.Set(key, value)
	return r
}

// AddHeader 添加请求头
func (r *Request) AddHeader(key, value string) *Request {
	r.Headers.Add(key, value)
	return r
}

// SetMeta 设置元数据
func (r *Request) SetMeta(key string, value interface{}) *Request {
	r.Meta[key] = value
	return r
}

// GetMeta 获取元数据
func (r *Request) GetMeta(key string) interface{} {
	return r.Meta[key]
}

// SetPriority 设置优先级
func (r *Request) SetPriority(priority int) *Request {
	r.Priority = priority
	return r
}

// SetProxy 设置代理
func (r *Request) SetProxy(proxy string) *Request {
	r.Proxy = proxy
	return r
}

// SetTimeout 设置超时
func (r *Request) SetTimeout(timeout time.Duration) *Request {
	r.Timeout = timeout
	return r
}

// AddCookie 添加Cookie
func (r *Request) AddCookie(cookie *http.Cookie) *Request {
	r.Cookies = append(r.Cookies, cookie)
	return r
}

// Copy 复制请求
func (r *Request) Copy() *Request {
	newReq := &Request{
		Method:       r.Method,
		URL:          r.URL,
		Headers:      make(http.Header),
		Body:         make([]byte, len(r.Body)),
		Meta:         make(map[string]interface{}),
		Priority:     r.Priority,
		RetryTimes:   r.RetryTimes,
		DontRetry:    r.DontRetry,
		Callback:     r.Callback,
		Proxy:        r.Proxy,
		Timeout:      r.Timeout,
		DontRedirect: r.DontRedirect,
	}
	
	// 复制Headers
	for k, v := range r.Headers {
		newReq.Headers[k] = v
	}
	
	// 复制Body
	copy(newReq.Body, r.Body)
	
	// 复制Meta
	for k, v := range r.Meta {
		newReq.Meta[k] = v
	}
	
	// 复制Cookies
	newReq.Cookies = make([]*http.Cookie, len(r.Cookies))
	copy(newReq.Cookies, r.Cookies)
	
	return newReq
}

// GetURL 获取解析后的URL
func (r *Request) GetURL() (*url.URL, error) {
	return url.Parse(r.URL)
}

// String 返回请求的字符串表示
func (r *Request) String() string {
	return r.Method + " " + r.URL
}