package settings

import (
	"encoding/json"
	"io/ioutil"
	"runtime"
	"time"
)

// Settings 爬虫设置配置
type Settings struct {
	// 基础设置
	BotName     string `json:"bot_name"`
	UserAgent   string `json:"user_agent"`
	RobotstxtObey bool `json:"robotstxt_obey"`
	
	// 并发设置
	ConcurrentRequests         int `json:"concurrent_requests"`
	ConcurrentRequestsPerDomain int `json:"concurrent_requests_per_domain"`
	
	// 下载设置
	DownloadDelay         time.Duration `json:"download_delay"`
	RandomizeDownloadDelay bool         `json:"randomize_download_delay"`
	DownloadTimeout       time.Duration `json:"download_timeout"`
	
	// 重试设置
	RetryEnabled bool `json:"retry_enabled"`
	RetryTimes   int  `json:"retry_times"`
	RetryHTTPCodes []int `json:"retry_http_codes"`
	
	// 中间件设置
	DownloaderMiddlewares map[string]int `json:"downloader_middlewares"`
	SpiderMiddlewares     map[string]int `json:"spider_middlewares"`
	
	// 管道设置
	ItemPipelines map[string]int `json:"item_pipelines"`
	
	// 输出设置
	FeedsExport map[string]FeedExportSettings `json:"feeds_export"`
	
	// 日志设置
	LogLevel  string `json:"log_level"`
	LogFile   string `json:"log_file"`
	LogStdout bool   `json:"log_stdout"`
	
	// 缓存设置
	CacheEnabled bool   `json:"cache_enabled"`
	CacheExpire  int    `json:"cache_expire"`
	CacheDir     string `json:"cache_dir"`
	
	// 自定义设置
	Custom map[string]interface{} `json:"custom"`
}

// FeedExportSettings 输出设置
type FeedExportSettings struct {
	Format   string            `json:"format"`
	URI      string            `json:"uri"`
	Fields   []string          `json:"fields"`
	Encoding string            `json:"encoding"`
	Headers  map[string]string `json:"headers"`
}

// DefaultSettings 默认设置
func DefaultSettings() *Settings {
	numCPU := runtime.NumCPU()
	concurrency := numCPU * 4
	if concurrency > 48 {
		concurrency = 48
	}
	if concurrency < 4 {
		concurrency = 4
	}

	return &Settings{
		// 基础设置
		BotName:       "go-scrapy",
		UserAgent:     "go-scrapy/1.0 (+https://github.com/go-scrapy/go-scrapy)",
		RobotstxtObey: false,
		
		// 并发设置
		ConcurrentRequests:         concurrency,
		ConcurrentRequestsPerDomain: 8,
		
		// 下载设置
		DownloadDelay:         100 * time.Millisecond,
		RandomizeDownloadDelay: true,
		DownloadTimeout:       30 * time.Second,
		
		// 重试设置
		RetryEnabled:   true,
		RetryTimes:     3,
		RetryHTTPCodes: []int{500, 502, 503, 504, 408, 429},
		
		// 中间件设置
		DownloaderMiddlewares: map[string]int{
			"UserAgentMiddleware": 400,
			"DelayMiddleware":     500,
		},
		SpiderMiddlewares: map[string]int{},
		
		// 管道设置
		ItemPipelines: map[string]int{
			"ConsolePipeline": 100,
			"JSONPipeline":    200,
		},
		
		// 输出设置
		FeedsExport: map[string]FeedExportSettings{
			"items.json": {
				Format:   "json",
				URI:      "items.json",
				Encoding: "utf-8",
			},
		},
		
		// 日志设置
		LogLevel:  "INFO",
		LogStdout: true,
		
		// 缓存设置
		CacheEnabled: false,
		CacheExpire:  3600,
		CacheDir:     ".scrapy/cache",
		
		// 自定义设置
		Custom: make(map[string]interface{}),
	}
}

// DoubanSettings 豆瓣专用设置
func DoubanSettings() *Settings {
	settings := DefaultSettings()
	
	// 豆瓣优化配置
	settings.BotName = "douban-spider"
	settings.ConcurrentRequests = runtime.NumCPU() * 3
	if settings.ConcurrentRequests > 36 {
		settings.ConcurrentRequests = 36
	}
	settings.DownloadDelay = 50 * time.Millisecond
	settings.DownloadTimeout = 20 * time.Second
	settings.RetryTimes = 2
	
	// 豆瓣专用User-Agent
	settings.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	
	// 输出设置
	settings.FeedsExport = map[string]FeedExportSettings{
		"douban_movies.json": {
			Format:   "json",
			URI:      "douban_movies.json",
			Encoding: "utf-8",
		},
	}
	
	// 自定义豆瓣设置
	settings.Custom["douban_api_base"] = "https://movie.douban.com/j/search_subjects"
	settings.Custom["douban_headers"] = map[string]string{
		"accept":          "application/json, text/plain, */*",
		"accept-language": "zh-CN,zh;q=0.9,en;q=0.8",
		"cache-control":   "no-cache",
		"origin":          "https://movie.douban.com",
		"referer":         "https://movie.douban.com/explore",
	}
	
	return settings
}

// FastSettings 快速模式设置
func FastSettings() *Settings {
	settings := DoubanSettings()
	
	// 快速模式优化
	settings.BotName = "fast-douban-spider"
	settings.ConcurrentRequests = runtime.NumCPU() * 4
	if settings.ConcurrentRequests > 48 {
		settings.ConcurrentRequests = 48
	}
	settings.DownloadDelay = 30 * time.Millisecond
	settings.RandomizeDownloadDelay = false
	settings.RetryTimes = 1
	
	// 快速模式输出
	settings.FeedsExport = map[string]FeedExportSettings{
		"fast_douban_movies.json": {
			Format:   "json",
			URI:      "fast_douban_movies.json",
			Encoding: "utf-8",
		},
	}
	
	return settings
}

// Get 获取设置值
func (s *Settings) Get(key string, defaultValue interface{}) interface{} {
	switch key {
	case "BOT_NAME":
		return s.BotName
	case "USER_AGENT":
		return s.UserAgent
	case "ROBOTSTXT_OBEY":
		return s.RobotstxtObey
	case "CONCURRENT_REQUESTS":
		return s.ConcurrentRequests
	case "CONCURRENT_REQUESTS_PER_DOMAIN":
		return s.ConcurrentRequestsPerDomain
	case "DOWNLOAD_DELAY":
		return s.DownloadDelay
	case "RANDOMIZE_DOWNLOAD_DELAY":
		return s.RandomizeDownloadDelay
	case "DOWNLOAD_TIMEOUT":
		return s.DownloadTimeout
	case "RETRY_ENABLED":
		return s.RetryEnabled
	case "RETRY_TIMES":
		return s.RetryTimes
	case "RETRY_HTTP_CODES":
		return s.RetryHTTPCodes
	case "DOWNLOADER_MIDDLEWARES":
		return s.DownloaderMiddlewares
	case "SPIDER_MIDDLEWARES":
		return s.SpiderMiddlewares
	case "ITEM_PIPELINES":
		return s.ItemPipelines
	case "FEEDS_EXPORT":
		return s.FeedsExport
	case "LOG_LEVEL":
		return s.LogLevel
	case "LOG_FILE":
		return s.LogFile
	case "LOG_STDOUT":
		return s.LogStdout
	case "CACHE_ENABLED":
		return s.CacheEnabled
	case "CACHE_EXPIRE":
		return s.CacheExpire
	case "CACHE_DIR":
		return s.CacheDir
	default:
		if val, exists := s.Custom[key]; exists {
			return val
		}
		return defaultValue
	}
}

// Set 设置值
func (s *Settings) Set(key string, value interface{}) {
	switch key {
	case "BOT_NAME":
		if v, ok := value.(string); ok {
			s.BotName = v
		}
	case "USER_AGENT":
		if v, ok := value.(string); ok {
			s.UserAgent = v
		}
	case "CONCURRENT_REQUESTS":
		if v, ok := value.(int); ok {
			s.ConcurrentRequests = v
		}
	case "DOWNLOAD_DELAY":
		if v, ok := value.(time.Duration); ok {
			s.DownloadDelay = v
		}
	default:
		s.Custom[key] = value
	}
}

// LoadFromFile 从JSON文件加载设置
func LoadFromFile(filename string) (*Settings, error) {
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var jsonSettings struct {
		BotName                    string            `json:"bot_name"`
		UserAgent                  string            `json:"user_agent"`
		RobotstxtObey             bool              `json:"robotstxt_obey"`
		ConcurrentRequests         int               `json:"concurrent_requests"`
		ConcurrentRequestsPerDomain int              `json:"concurrent_requests_per_domain"`
		DownloadDelay              string            `json:"download_delay"`
		RandomizeDownloadDelay     bool              `json:"randomize_download_delay"`
		DownloadTimeout            string            `json:"download_timeout"`
		RetryEnabled               bool              `json:"retry_enabled"`
		RetryTimes                 int               `json:"retry_times"`
		RetryHTTPCodes            []int             `json:"retry_http_codes"`
		DownloaderMiddlewares      map[string]int    `json:"downloader_middlewares"`
		SpiderMiddlewares          map[string]int    `json:"spider_middlewares"`
		ItemPipelines              map[string]int    `json:"item_pipelines"`
		FeedsExport               map[string]FeedExportSettings `json:"feeds_export"`
		LogLevel                   string            `json:"log_level"`
		LogFile                    string            `json:"log_file"`
		LogStdout                  bool              `json:"log_stdout"`
		CacheEnabled               bool              `json:"cache_enabled"`
		CacheExpire                int               `json:"cache_expire"`
		CacheDir                   string            `json:"cache_dir"`
		Custom                     map[string]interface{} `json:"custom"`
	}
	
	if err := json.Unmarshal(data, &jsonSettings); err != nil {
		return nil, err
	}
	
	settings := &Settings{
		BotName:                    jsonSettings.BotName,
		UserAgent:                  jsonSettings.UserAgent,
		RobotstxtObey:             jsonSettings.RobotstxtObey,
		ConcurrentRequests:         jsonSettings.ConcurrentRequests,
		ConcurrentRequestsPerDomain: jsonSettings.ConcurrentRequestsPerDomain,
		RandomizeDownloadDelay:     jsonSettings.RandomizeDownloadDelay,
		RetryEnabled:               jsonSettings.RetryEnabled,
		RetryTimes:                 jsonSettings.RetryTimes,
		RetryHTTPCodes:            jsonSettings.RetryHTTPCodes,
		DownloaderMiddlewares:      jsonSettings.DownloaderMiddlewares,
		SpiderMiddlewares:          jsonSettings.SpiderMiddlewares,
		ItemPipelines:              jsonSettings.ItemPipelines,
		FeedsExport:               jsonSettings.FeedsExport,
		LogLevel:                   jsonSettings.LogLevel,
		LogFile:                    jsonSettings.LogFile,
		LogStdout:                  jsonSettings.LogStdout,
		CacheEnabled:               jsonSettings.CacheEnabled,
		CacheExpire:                jsonSettings.CacheExpire,
		CacheDir:                   jsonSettings.CacheDir,
		Custom:                     jsonSettings.Custom,
	}
	
	// 解析时间字符串
	if jsonSettings.DownloadDelay != "" {
		if duration, err := time.ParseDuration(jsonSettings.DownloadDelay); err == nil {
			settings.DownloadDelay = duration
		}
	}
	
	if jsonSettings.DownloadTimeout != "" {
		if duration, err := time.ParseDuration(jsonSettings.DownloadTimeout); err == nil {
			settings.DownloadTimeout = duration
		}
	}
	
	return settings, nil
}