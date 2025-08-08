package middleware

import (
	"fmt"
	"scrago/request"
	"scrago/response"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Middleware 中间件接口
type Middleware interface {
	ProcessRequest(req *request.Request) *request.Request
	ProcessResponse(req *request.Request, resp *response.Response) *response.Response
}

// UserAgentMiddleware User-Agent中间件
type UserAgentMiddleware struct {
	userAgents []string
	random     bool
}

// NewUserAgentMiddleware 创建User-Agent中间件
func NewUserAgentMiddleware(userAgents []string, random bool) *UserAgentMiddleware {
	if len(userAgents) == 0 {
		userAgents = []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		}
	}
	
	return &UserAgentMiddleware{
		userAgents: userAgents,
		random:     random,
	}
}

// ProcessRequest 处理请求
func (m *UserAgentMiddleware) ProcessRequest(req *request.Request) *request.Request {
	if req.Headers.Get("User-Agent") == "" {
		var userAgent string
		if m.random && len(m.userAgents) > 1 {
			userAgent = m.userAgents[rand.Intn(len(m.userAgents))]
		} else {
			userAgent = m.userAgents[0]
		}
		req.SetHeader("User-Agent", userAgent)
	}
	return req
}

// ProcessResponse 处理响应
func (m *UserAgentMiddleware) ProcessResponse(req *request.Request, resp *response.Response) *response.Response {
	return resp
}

// ProxyMiddleware 代理中间件
type ProxyMiddleware struct {
	proxies []string
	random  bool
}

// NewProxyMiddleware 创建代理中间件
func NewProxyMiddleware(proxies []string, random bool) *ProxyMiddleware {
	return &ProxyMiddleware{
		proxies: proxies,
		random:  random,
	}
}

// ProcessRequest 处理请求
func (m *ProxyMiddleware) ProcessRequest(req *request.Request) *request.Request {
	if req.Proxy == "" && len(m.proxies) > 0 {
		var proxy string
		if m.random && len(m.proxies) > 1 {
			proxy = m.proxies[rand.Intn(len(m.proxies))]
		} else {
			proxy = m.proxies[0]
		}
		req.SetProxy(proxy)
	}
	return req
}

// ProcessResponse 处理响应
func (m *ProxyMiddleware) ProcessResponse(req *request.Request, resp *response.Response) *response.Response {
	return resp
}

// RetryMiddleware 重试中间件
type RetryMiddleware struct {
	maxRetries     int
	retryHTTPCodes []int
}

// NewRetryMiddleware 创建重试中间件
func NewRetryMiddleware(maxRetries int, retryHTTPCodes []int) *RetryMiddleware {
	if len(retryHTTPCodes) == 0 {
		retryHTTPCodes = []int{500, 502, 503, 504, 408, 429}
	}
	
	return &RetryMiddleware{
		maxRetries:     maxRetries,
		retryHTTPCodes: retryHTTPCodes,
	}
}

// ProcessRequest 处理请求
func (m *RetryMiddleware) ProcessRequest(req *request.Request) *request.Request {
	return req
}

// ProcessResponse 处理响应
func (m *RetryMiddleware) ProcessResponse(req *request.Request, resp *response.Response) *response.Response {
	// 检查是否需要重试
	if m.shouldRetry(resp.StatusCode) && req.RetryTimes < m.maxRetries {
		req.RetryTimes++
		// 这里应该重新调度请求，但由于架构限制，我们只是标记
		fmt.Printf("Retrying request %s (attempt %d/%d)\n", req.URL, req.RetryTimes, m.maxRetries)
	}
	
	return resp
}

// shouldRetry 检查是否应该重试
func (m *RetryMiddleware) shouldRetry(statusCode int) bool {
	for _, code := range m.retryHTTPCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// CookieMiddleware Cookie中间件
type CookieMiddleware struct {
	cookieJar map[string][]*http.Cookie
}

// NewCookieMiddleware 创建Cookie中间件
func NewCookieMiddleware() *CookieMiddleware {
	return &CookieMiddleware{
		cookieJar: make(map[string][]*http.Cookie),
	}
}

// ProcessRequest 处理请求
func (m *CookieMiddleware) ProcessRequest(req *request.Request) *request.Request {
	// 从cookie jar中获取cookies
	if cookies, exists := m.cookieJar[m.getDomain(req.URL)]; exists {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}
	return req
}

// ProcessResponse 处理响应
func (m *CookieMiddleware) ProcessResponse(req *request.Request, resp *response.Response) *response.Response {
	// 保存响应中的cookies
	if setCookies := resp.Headers["Set-Cookie"]; len(setCookies) > 0 {
		domain := m.getDomain(req.URL)
		for _, setCookie := range setCookies {
			if cookie := m.parseCookie(setCookie); cookie != nil {
				m.cookieJar[domain] = append(m.cookieJar[domain], cookie)
			}
		}
	}
	return resp
}

// getDomain 获取域名
func (m *CookieMiddleware) getDomain(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return url
}

// parseCookie 解析Cookie
func (m *CookieMiddleware) parseCookie(setCookie string) *http.Cookie {
	// 简单的Cookie解析，实际应该使用更完善的解析器
	parts := strings.Split(setCookie, ";")
	if len(parts) == 0 {
		return nil
	}
	
	nameValue := strings.Split(strings.TrimSpace(parts[0]), "=")
	if len(nameValue) != 2 {
		return nil
	}
	
	return &http.Cookie{
		Name:  nameValue[0],
		Value: nameValue[1],
	}
}

// DelayMiddleware 延迟中间件
type DelayMiddleware struct {
	delay      time.Duration
	randomize  bool
	lastAccess map[string]time.Time
	mutex      sync.RWMutex
}

// NewDelayMiddleware 创建延迟中间件
func NewDelayMiddleware(delay time.Duration, randomize bool) *DelayMiddleware {
	return &DelayMiddleware{
		delay:      delay,
		randomize:  randomize,
		lastAccess: make(map[string]time.Time),
	}
}

// ProcessRequest 处理请求
func (m *DelayMiddleware) ProcessRequest(req *request.Request) *request.Request {
	domain := m.getDomain(req.URL)
	
	m.mutex.RLock()
	lastTime, exists := m.lastAccess[domain]
	m.mutex.RUnlock()
	
	if exists {
		elapsed := time.Since(lastTime)
		delay := m.delay
		
		if m.randomize {
			// 随机化延迟时间（0.5 * delay 到 1.5 * delay）
			factor := 0.5 + rand.Float64()
			delay = time.Duration(float64(delay) * factor)
		}
		
		if elapsed < delay {
			time.Sleep(delay - elapsed)
		}
	}
	
	m.mutex.Lock()
	m.lastAccess[domain] = time.Now()
	m.mutex.Unlock()
	
	return req
}

// ProcessResponse 处理响应
func (m *DelayMiddleware) ProcessResponse(req *request.Request, resp *response.Response) *response.Response {
	return resp
}

// getDomain 获取域名
func (m *DelayMiddleware) getDomain(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return url
}

// HeaderMiddleware 请求头中间件
type HeaderMiddleware struct {
	headers map[string]string
}

// NewHeaderMiddleware 创建请求头中间件
func NewHeaderMiddleware(headers map[string]string) *HeaderMiddleware {
	return &HeaderMiddleware{
		headers: headers,
	}
}

// ProcessRequest 处理请求
func (m *HeaderMiddleware) ProcessRequest(req *request.Request) *request.Request {
	for key, value := range m.headers {
		if req.Headers.Get(key) == "" {
			req.SetHeader(key, value)
		}
	}
	return req
}

// ProcessResponse 处理响应
func (m *HeaderMiddleware) ProcessResponse(req *request.Request, resp *response.Response) *response.Response {
	return resp
}