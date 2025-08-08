package downloader

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"scrago/request"
	"scrago/response"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
)

// Downloader 下载器接口
type Downloader interface {
	Download(req *request.Request) (*response.Response, error)
	DownloadAsync(req *request.Request) <-chan *AsyncResult
	DownloadBatch(reqs []*request.Request) <-chan *AsyncResult
}

// AsyncResult 异步下载结果
type AsyncResult struct {
	Request  *request.Request
	Response *response.Response
	Error    error
}

// HTTPDownloader HTTP下载器
type HTTPDownloader struct {
	client   *http.Client
	userAgent string
}

// NewHTTPDownloader 创建HTTP下载器
func NewHTTPDownloader() *HTTPDownloader {
	// 创建自定义Transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过SSL验证
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 默认跟随重定向，最多10次
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}
	
	return &HTTPDownloader{
		client:    client,
		userAgent: "Go-Scrapy/1.0",
	}
}

// Download 下载请求
func (d *HTTPDownloader) Download(req *request.Request) (*response.Response, error) {
	// 创建HTTP请求
	httpReq, err := d.buildHTTPRequest(req)
	if err != nil {
		return nil, fmt.Errorf("build request failed: %w", err)
	}
	
	// 创建独立的客户端副本以避免并发问题
	client := &http.Client{
		Transport: d.client.Transport,
		Timeout:   10 * time.Second, // 减少超时时间以快速发现问题
		CheckRedirect: d.client.CheckRedirect,
	}
	
	// 设置代理
	if req.Proxy != "" {
		proxyURL, err := url.Parse(req.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		
		// 创建独立的transport副本
		transport := d.client.Transport.(*http.Transport).Clone()
		transport.Proxy = http.ProxyURL(proxyURL)
		client.Transport = transport
	}
	
	// 设置超时（现在是线程安全的）
	if req.Timeout > 0 {
		client.Timeout = req.Timeout
	}
	
	// 添加网络诊断日志
	fmt.Printf("🌐 开始执行HTTP请求: %s\n", req.URL)
	start := time.Now()
	
	// 执行请求
	httpResp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ HTTP请求失败 (%v): %s - %v\n", time.Since(start), req.URL, err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()
	
	fmt.Printf("✅ HTTP请求成功 (%v): %s - 状态码: %d\n", time.Since(start), req.URL, httpResp.StatusCode)
	
	// 检查并处理压缩的响应
	var bodyReader io.Reader = httpResp.Body
	contentEncoding := httpResp.Header.Get("Content-Encoding")
	
	switch contentEncoding {
	case "gzip":
		fmt.Printf("🗜️  检测到gzip压缩，正在解压缩: %s\n", req.URL)
		gzipReader, err := gzip.NewReader(httpResp.Body)
		if err != nil {
			fmt.Printf("❌ gzip解压缩失败: %s - %v\n", req.URL, err)
			return nil, fmt.Errorf("gzip decompression failed: %w", err)
		}
		defer gzipReader.Close()
		bodyReader = gzipReader
	case "br":
		fmt.Printf("🗜️  检测到Brotli压缩，正在解压缩: %s\n", req.URL)
		bodyReader = brotli.NewReader(httpResp.Body)
	}
	
	// 读取响应体
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		fmt.Printf("❌ 读取响应体失败: %s - %v\n", req.URL, err)
		return nil, fmt.Errorf("read response body failed: %w", err)
	}
	
	fmt.Printf("📄 响应体读取完成: %s - 大小: %d bytes (编码: %s)\n", req.URL, len(body), contentEncoding)
	
	// 创建响应对象
	resp := response.NewResponse(
		httpResp.Request.URL.String(),
		httpResp.StatusCode,
		httpResp.Header,
		body,
		req,
	)
	
	return resp, nil
}

// buildHTTPRequest 构建HTTP请求
func (d *HTTPDownloader) buildHTTPRequest(req *request.Request) (*http.Request, error) {
	var body io.Reader
	if len(req.Body) > 0 {
		body = strings.NewReader(string(req.Body))
	}
	
	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return nil, err
	}
	
	// 设置请求头
	for key, values := range req.Headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}
	
	// 设置默认User-Agent（如果没有设置的话）
	if httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", d.userAgent)
	}
	
	// 设置Cookies
	for _, cookie := range req.Cookies {
		httpReq.AddCookie(cookie)
	}
	
	return httpReq, nil
}

// SetUserAgent 设置User-Agent
func (d *HTTPDownloader) SetUserAgent(userAgent string) {
	d.userAgent = userAgent
}

// SetTimeout 设置超时时间
func (d *HTTPDownloader) SetTimeout(timeout time.Duration) {
	d.client.Timeout = timeout
}

// SetProxy 设置代理
func (d *HTTPDownloader) SetProxy(proxyURL string) error {
	if proxyURL == "" {
		// 清除代理
		transport := d.client.Transport.(*http.Transport)
		transport.Proxy = nil
		return nil
	}
	
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}
	
	transport := d.client.Transport.(*http.Transport)
	transport.Proxy = http.ProxyURL(proxy)
	
	return nil
}

// EnableCookieJar 启用Cookie管理
func (d *HTTPDownloader) EnableCookieJar() {
	// 这里可以实现Cookie Jar功能
	// d.client.Jar = cookiejar.New(nil)
}

// SetTLSConfig 设置TLS配置
func (d *HTTPDownloader) SetTLSConfig(config *tls.Config) {
	transport := d.client.Transport.(*http.Transport)
	transport.TLSClientConfig = config
}

// DownloadAsync 异步下载单个请求
func (d *HTTPDownloader) DownloadAsync(req *request.Request) <-chan *AsyncResult {
	resultChan := make(chan *AsyncResult, 1)
	
	go func() {
		defer close(resultChan)
		
		resp, err := d.Download(req)
		resultChan <- &AsyncResult{
			Request:  req,
			Response: resp,
			Error:    err,
		}
	}()
	
	return resultChan
}

// DownloadBatch 批量异步下载请求
func (d *HTTPDownloader) DownloadBatch(reqs []*request.Request) <-chan *AsyncResult {
	resultChan := make(chan *AsyncResult, len(reqs))
	
	var wg sync.WaitGroup
	
	fmt.Printf("🔧 下载器：创建缓冲通道，容量 %d\n", len(reqs))
	fmt.Printf("🔧 下载器：准备启动 %d 个并发请求\n", len(reqs))
	
	// 🚀 异步发送所有请求
	for i, req := range reqs {
		wg.Add(1)
		fmt.Printf("🔧 下载器：启动 goroutine %d for URL: %s\n", i+1, req.URL)
		
		go func(index int, r *request.Request) {
			defer wg.Done()
			
			fmt.Printf("🔧 下载器：[%d] 开始处理请求: %s\n", index+1, r.URL)
			
			// 异步下载
			resp, err := d.Download(r)
			
			var statusCode int
			if resp != nil {
				statusCode = resp.StatusCode
			}
			
			fmt.Printf("🔧 下载器：[%d] 请求完成: %s (状态码: %d, 错误: %v)\n", index+1, r.URL, statusCode, err)
			
			// 发送结果到通道
			result := &AsyncResult{
				Request:  r,
				Response: resp,
				Error:    err,
			}
			
			fmt.Printf("🔧 下载器：[%d] 发送结果到通道: %s\n", index+1, r.URL)
			resultChan <- result
			fmt.Printf("🔧 下载器：[%d] 结果已发送: %s\n", index+1, r.URL)
		}(i, req)
	}
	
	// 等待所有请求完成后关闭通道
	go func() {
		fmt.Printf("🔧 下载器：等待所有请求完成...\n")
		wg.Wait()
		fmt.Printf("🔧 下载器：所有请求完成，关闭通道\n")
		close(resultChan)
		fmt.Printf("🔧 下载器：通道已关闭\n")
	}()
	
	return resultChan
}

// DownloadBatchWithContext 带上下文的批量异步下载
func (d *HTTPDownloader) DownloadBatchWithContext(ctx context.Context, reqs []*request.Request) <-chan *AsyncResult {
	resultChan := make(chan *AsyncResult, len(reqs))
	
	var wg sync.WaitGroup
	
	// 🚀 异步发送所有请求
	for _, req := range reqs {
		wg.Add(1)
		go func(r *request.Request) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				// 上下文取消
				resultChan <- &AsyncResult{
					Request:  r,
					Response: nil,
					Error:    ctx.Err(),
				}
				return
			default:
				// 异步下载
				resp, err := d.Download(r)
				
				// 发送结果到通道
				select {
				case resultChan <- &AsyncResult{
					Request:  r,
					Response: resp,
					Error:    err,
				}:
				case <-ctx.Done():
					// 上下文取消，丢弃结果
				}
			}
		}(req)
	}
	
	// 等待所有请求完成后关闭通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	return resultChan
}