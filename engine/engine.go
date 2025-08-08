package engine

import (
	"context"
	"fmt"
	"scrago/downloader"
	"scrago/middleware"
	"scrago/pipeline"
	"scrago/request"
	"scrago/scheduler"
	"scrago/spider"
	"sync"
	"time"
)

// Engine 爬虫引擎核心
type Engine struct {
	scheduler   scheduler.Scheduler
	downloader  downloader.Downloader
	pipelines   []pipeline.Pipeline
	middlewares []middleware.Middleware
	
	// 并发控制
	concurrency int
	workers     chan struct{}
	wg          sync.WaitGroup
	
	// 🚀 结果处理协程池 - 专门处理yield返回的请求
	resultPool     chan interface{}
	resultWorkers  int
	resultWg       sync.WaitGroup
	
	// 🚀 当前运行的spider实例（用于异步批量处理）
	currentSpider  spider.Spider
	spiderMutex    sync.RWMutex
	
	// 🚀 异步批量处理等待机制
	batchWg        sync.WaitGroup
	
	// 统计信息
	stats       *Stats
	
	// 配置
	settings    *Settings
}

// Stats 统计信息
type Stats struct {
	RequestsTotal    int64
	RequestsSuccess  int64
	RequestsFailed   int64
	ItemsScraped     int64
	StartTime        time.Time
	mu               sync.RWMutex
}

// Settings 引擎配置
type Settings struct {
	Concurrency      int
	DownloadDelay    time.Duration
	RandomizeDelay   bool
	UserAgent        string
	RobotsTxtObey    bool
	AutoThrottle     bool
	RetryTimes       int
	RetryHTTPCodes   []int
}

// NewEngine 创建新的爬虫引擎
func NewEngine() *Engine {
	settings := &Settings{
		Concurrency:    16,
		DownloadDelay:  time.Second,
		RandomizeDelay: true,
		UserAgent:      "Go-Scrapy/1.0",
		RobotsTxtObey:  true,
		AutoThrottle:   true,
		RetryTimes:     3,
		RetryHTTPCodes: []int{500, 502, 503, 504, 408, 429},
	}
	
	resultWorkers := settings.Concurrency / 2
	if resultWorkers < 2 {
		resultWorkers = 2
	}
	
	return &Engine{
		scheduler:   scheduler.NewChannelScheduler(settings.Concurrency * 4), // 使用高性能调度器
		downloader:  downloader.NewHTTPDownloader(),
		pipelines:   make([]pipeline.Pipeline, 0),
		middlewares: make([]middleware.Middleware, 0),
		concurrency: settings.Concurrency,
		workers:     make(chan struct{}, settings.Concurrency),
		
		// 🚀 初始化结果处理协程池
		resultPool:    make(chan interface{}, settings.Concurrency * 8),
		resultWorkers: resultWorkers,
		
		stats: &Stats{
			StartTime: time.Now(),
		},
		settings: settings,
	}
}

// AddPipeline 添加数据管道
func (e *Engine) AddPipeline(p pipeline.Pipeline) {
	e.pipelines = append(e.pipelines, p)
}

// AddMiddleware 添加中间件
func (e *Engine) AddMiddleware(m middleware.Middleware) {
	e.middlewares = append(e.middlewares, m)
}

// SetConcurrency 设置并发数
func (e *Engine) SetConcurrency(concurrency int) {
	e.concurrency = concurrency
	e.settings.Concurrency = concurrency
	e.workers = make(chan struct{}, concurrency)
}

// Run 运行爬虫
func (e *Engine) Run(s spider.Spider) error {
	fmt.Printf("Starting spider: %s\n", s.Name())
	
	// 🚀 打开所有管道
	for _, p := range e.pipelines {
		if err := p.Open(); err != nil {
			return fmt.Errorf("failed to open pipeline: %w", err)
		}
	}
	
	// 🚀 确保在函数结束时关闭所有管道
	defer func() {
		for _, p := range e.pipelines {
			if err := p.Close(); err != nil {
				fmt.Printf("Warning: failed to close pipeline: %v\n", err)
			}
		}
	}()
	
	// 🚀 设置当前spider实例（用于异步批量处理）
	e.spiderMutex.Lock()
	e.currentSpider = s
	e.spiderMutex.Unlock()
	
	// 初始化爬虫
	startRequests := s.StartRequests()
	for _, req := range startRequests {
		e.scheduler.Enqueue(req)
	}
	
	// 启动上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 🚀 启动结果处理协程池
	e.startResultWorkers(ctx)
	
	// 启动主工作协程池
	for i := 0; i < e.concurrency; i++ {
		e.wg.Add(1)
		go e.worker(ctx, s)
	}
	
	// 等待所有任务完成
	e.wg.Wait()
	
	// 🚀 等待所有异步批量处理完成
	fmt.Printf("⏳ 等待异步批量处理完成...\n")
	e.batchWg.Wait()
	fmt.Printf("✅ 所有异步批量处理已完成\n")
	
	// 关闭结果处理协程池
	close(e.resultPool)
	e.resultWg.Wait()
	
	// 打印统计信息
	e.printStats()
	
	return nil
}

// startResultWorkers 启动结果处理协程池
func (e *Engine) startResultWorkers(ctx context.Context) {
	for i := 0; i < e.resultWorkers; i++ {
		e.resultWg.Add(1)
		go func(workerID int) {
			defer e.resultWg.Done()
			fmt.Printf("🚀 Result worker %d started\n", workerID)
			
			for {
				select {
				case <-ctx.Done():
					return
				case result, ok := <-e.resultPool:
					if !ok {
						return // channel已关闭
					}
					e.processResult(result)
				}
			}
		}(i)
	}
	fmt.Printf("🚀 Started %d result workers for yield processing\n", e.resultWorkers)
}


// worker 工作协程
func (e *Engine) worker(ctx context.Context, s spider.Spider) {
	defer e.wg.Done()
	
	emptyCount := 0
	maxEmptyCount := 100 // 连续空闲100次后退出（约100ms）
	
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req := e.scheduler.Dequeue()
			if req == nil {
				// 没有更多请求，检查是否应该退出
				if e.scheduler.Empty() {
					emptyCount++
					if emptyCount >= maxEmptyCount {
						// 连续空闲足够长时间，可能所有yield请求都处理完了
						return
					}
				} else {
					emptyCount = 0 // 重置计数器
				}
				// 短暂等待后继续
				time.Sleep(1 * time.Millisecond)
				continue
			}
			
			emptyCount = 0 // 有请求时重置计数器
			// 直接处理请求（已经在协程中）
			e.processRequest(req, s)
		}
	}
}

// processRequest 处理单个请求
func (e *Engine) processRequest(req *request.Request, s spider.Spider) {
	e.updateStats("request_total", 1)
	
	// 应用下载中间件
	for _, mw := range e.middlewares {
		req = mw.ProcessRequest(req)
		if req == nil {
			return
		}
	}
	
	// 下载
	resp, err := e.downloader.Download(req)
	if err != nil {
		e.updateStats("request_failed", 1)
		fmt.Printf("Download failed: %v\n", err)
		return
	}
	
	e.updateStats("request_success", 1)
	
	// 应用响应中间件
	for _, mw := range e.middlewares {
		resp = mw.ProcessResponse(req, resp)
		if resp == nil {
			return
		}
	}
	
	// 解析响应
	results := s.Parse(resp)
	
	// 🚀 协程模式处理解析结果 - 关键优化点！
	e.processResultsConcurrently(results)
}

// processRequestAsync 异步处理单个请求
func (e *Engine) processRequestAsync(req *request.Request, s spider.Spider) {
	e.updateStats("request_total", 1)
	
	// 应用下载中间件
	for _, mw := range e.middlewares {
		req = mw.ProcessRequest(req)
		if req == nil {
			return
		}
	}
	
	// 🚀 异步下载
	resultChan := e.downloader.DownloadAsync(req)
	
	// 异步处理下载结果
	go func() {
		asyncResult := <-resultChan
		
		if asyncResult.Error != nil {
			e.updateStats("request_failed", 1)
			fmt.Printf("Async download failed: %v\n", asyncResult.Error)
			return
		}
		
		e.updateStats("request_success", 1)
		
		resp := asyncResult.Response
		
		// 应用响应中间件
		for _, mw := range e.middlewares {
			resp = mw.ProcessResponse(req, resp)
			if resp == nil {
				return
			}
		}
		
		// 解析响应
		results := s.Parse(resp)
		
		// 🚀 协程模式处理解析结果
		e.processResultsConcurrently(results)
	}()
}

// processBatchRequests 批量异步处理请求
func (e *Engine) processBatchRequests(reqs []*request.Request, s spider.Spider) {
	if len(reqs) == 0 {
		return
	}
	
	fmt.Printf("🚀 开始批量异步处理 %d 个请求\n", len(reqs))
	
	// 应用下载中间件
	validReqs := make([]*request.Request, 0, len(reqs))
	for _, req := range reqs {
		e.updateStats("request_total", 1)
		
		for _, mw := range e.middlewares {
			req = mw.ProcessRequest(req)
			if req == nil {
				break
			}
		}
		
		if req != nil {
			validReqs = append(validReqs, req)
		}
	}
	
	if len(validReqs) == 0 {
		return
	}
	
	// 🚀 批量异步下载 - 先发送所有请求
	resultChan := e.downloader.DownloadBatch(validReqs)
	
	// 🚀 异步处理所有响应
	go func() {
		for asyncResult := range resultChan {
			if asyncResult.Error != nil {
				e.updateStats("request_failed", 1)
				fmt.Printf("Batch download failed for %s: %v\n", asyncResult.Request.URL, asyncResult.Error)
				continue
			}
			
			e.updateStats("request_success", 1)
			
			resp := asyncResult.Response
			req := asyncResult.Request
			
			// 应用响应中间件
			for _, mw := range e.middlewares {
				resp = mw.ProcessResponse(req, resp)
				if resp == nil {
					break
				}
			}
			
			if resp != nil {
				// 解析响应
				results := s.Parse(resp)
				
				// 🚀 协程模式处理解析结果
				e.processResultsConcurrently(results)
			}
		}
		fmt.Printf("✅ 批量异步处理完成\n")
	}()
}

// processResultsConcurrently 并发处理解析结果
func (e *Engine) processResultsConcurrently(results []interface{}) {
	if len(results) == 0 {
		return
	}
	
	// 🚀 分离请求和数据项
	requests := make([]*request.Request, 0)
	items := make([]interface{}, 0)
	
	for _, result := range results {
		switch r := result.(type) {
		case *request.Request:
			requests = append(requests, r)
		default:
			items = append(items, result)
		}
	}
	
	// 🚀 批量异步处理请求（如果有多个请求）
	if len(requests) > 1 {
		fmt.Printf("🚀 检测到 %d 个请求，启用批量异步模式\n", len(requests))
		// 直接批量处理，不通过结果池
		e.batchWg.Add(1)
		go func() {
			defer e.batchWg.Done()
			e.processBatchRequestsDirectly(requests)
		}()
	} else {
		// 单个请求或无请求，使用原有逻辑
		for _, req := range requests {
			select {
			case e.resultPool <- req:
				// 成功发送到协程池
			default:
				// 协程池满时，异步发送避免阻塞
				go func(r interface{}) {
					e.resultPool <- r
				}(req)
			}
		}
	}
	
	// 🚀 处理数据项
	for _, item := range items {
		select {
		case e.resultPool <- item:
			// 成功发送到协程池
		default:
			// 协程池满时，异步发送避免阻塞
			go func(r interface{}) {
				e.resultPool <- r
			}(item)
		}
	}
}

// processBatchRequestsDirectly 直接批量处理请求（绕过结果池）
func (e *Engine) processBatchRequestsDirectly(requests []*request.Request) {
	if len(requests) == 0 {
		return
	}
	
	// 应用下载中间件
	validReqs := make([]*request.Request, 0, len(requests))
	for _, req := range requests {
		e.updateStats("request_total", 1)
		
		for _, mw := range e.middlewares {
			req = mw.ProcessRequest(req)
			if req == nil {
				break
			}
		}
		
		if req != nil {
			validReqs = append(validReqs, req)
		}
	}
	
	if len(validReqs) == 0 {
		return
	}
	
	fmt.Printf("🚀 批量异步发送 %d 个请求\n", len(validReqs))
	
	// 🚀 打印即将发送的请求URL
	for i, req := range validReqs {
		fmt.Printf("📤 准备发送请求 %d: %s\n", i+1, req.URL)
	}
	
	// 🚀 批量异步下载 - 先发送所有请求
	fmt.Printf("🔄 调用下载器批量异步下载...\n")
	resultChan := e.downloader.DownloadBatch(validReqs)
	fmt.Printf("✅ 下载器返回结果通道，开始监听响应...\n")
	
	// 🚀 异步处理所有响应
	processedCount := 0
	fmt.Printf("🔍 开始循环监听结果通道...\n")
	for asyncResult := range resultChan {
		processedCount++
		fmt.Printf("🔍 处理第 %d 个异步结果: %s\n", processedCount, asyncResult.Request.URL)
		
		if asyncResult.Error != nil {
			e.updateStats("request_failed", 1)
			fmt.Printf("❌ 批量下载失败 %s: %v\n", asyncResult.Request.URL, asyncResult.Error)
			continue
		}
		
		if asyncResult.Response == nil {
			e.updateStats("request_failed", 1)
			fmt.Printf("❌ 批量下载响应为空 %s\n", asyncResult.Request.URL)
			continue
		}
		
		e.updateStats("request_success", 1)
		fmt.Printf("✅ 批量下载成功 %s (状态码: %d)\n", asyncResult.Request.URL, asyncResult.Response.StatusCode)
		
		resp := asyncResult.Response
		req := asyncResult.Request
		
		// 应用响应中间件
		for _, mw := range e.middlewares {
			resp = mw.ProcessResponse(req, resp)
			if resp == nil {
				break
			}
		}
		
		if resp != nil {
			// 🚀 获取spider实例并解析响应
			e.spiderMutex.RLock()
			currentSpider := e.currentSpider
			e.spiderMutex.RUnlock()
			
			if currentSpider != nil {
				fmt.Printf("✅ 批量异步响应处理: %s (状态码: %d)\n", resp.URL, resp.StatusCode)
				
				// 🚀 使用spider解析响应
				results := currentSpider.Parse(resp)
				
				// 🚀 递归处理解析结果
				e.processResultsConcurrently(results)
			} else {
				fmt.Printf("⚠️ 无法获取spider实例，跳过响应解析: %s\n", resp.URL)
			}
		}
	}
	fmt.Printf("🔍 结果通道已关闭，退出循环\n")
	fmt.Printf("✅ 批量异步处理完成，共处理了 %d 个响应\n", processedCount)
}

// processResult 处理单个结果
func (e *Engine) processResult(result interface{}) {
	switch r := result.(type) {
	case *request.Request:
		// 直接入队新请求（已在协程池中）
		e.scheduler.Enqueue(r)
	case map[string]interface{}:
		// 直接处理数据项（已在协程池中）
		e.processItem(r)
	default:
		// 🚀 处理任意类型的数据项（如结构体）
		e.processAnyItem(r)
	}
}

// processItem 处理数据项
func (e *Engine) processItem(item map[string]interface{}) {
	e.updateStats("items_scraped", 1)
	
	// 通过管道处理数据
	for _, p := range e.pipelines {
		item = p.ProcessItem(item)
		if item == nil {
			return
		}
	}
}

// processAnyItem 处理任意类型的数据项
func (e *Engine) processAnyItem(item interface{}) {
	e.updateStats("items_scraped", 1)
	
	// 🚀 将任意类型转换为map[string]interface{}供管道处理
	var mapItem map[string]interface{}
	
	// 如果已经是map类型，直接使用
	if m, ok := item.(map[string]interface{}); ok {
		mapItem = m
	} else {
		// 🚀 对于结构体等其他类型，创建一个包装map
		mapItem = map[string]interface{}{
			"data": item,
			"type": fmt.Sprintf("%T", item),
		}
	}
	
	// 通过管道处理数据
	for _, p := range e.pipelines {
		mapItem = p.ProcessItem(mapItem)
		if mapItem == nil {
			return
		}
	}
}

// updateStats 更新统计信息
func (e *Engine) updateStats(key string, value int64) {
	e.stats.mu.Lock()
	defer e.stats.mu.Unlock()
	
	switch key {
	case "request_total":
		e.stats.RequestsTotal += value
	case "request_success":
		e.stats.RequestsSuccess += value
	case "request_failed":
		e.stats.RequestsFailed += value
	case "items_scraped":
		e.stats.ItemsScraped += value
	}
}

// printStats 打印统计信息
func (e *Engine) printStats() {
	e.stats.mu.RLock()
	defer e.stats.mu.RUnlock()
	
	duration := time.Since(e.stats.StartTime)
	
	fmt.Println("\n=== Crawl Stats ===")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Requests Total: %d\n", e.stats.RequestsTotal)
	fmt.Printf("Requests Success: %d\n", e.stats.RequestsSuccess)
	fmt.Printf("Requests Failed: %d\n", e.stats.RequestsFailed)
	fmt.Printf("Items Scraped: %d\n", e.stats.ItemsScraped)
	
	if duration.Seconds() > 0 {
		fmt.Printf("Requests/sec: %.2f\n", float64(e.stats.RequestsTotal)/duration.Seconds())
	}
}