package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"scrago/engine"
	"scrago/middleware"
	"scrago/pipeline"
	"scrago/settings"
	"scrago/spider"
	"scrago/spiders"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CrawlCommand 处理 crawl 命令
func CrawlCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ 错误: 请指定要运行的爬虫名称")
		fmt.Println("用法: scrapy crawl <spider_name>")
		fmt.Println("示例: scrapy crawl douban")
		return
	}

	spiderName := args[0]
	
	// 解析命令行参数
	fs := flag.NewFlagSet("crawl", flag.ExitOnError)
	
	// 设置参数
	settingsFlag := fs.String("s", "", "设置参数 (格式: KEY=VALUE)")
	configFile := fs.String("c", "", "配置文件路径")
	outputFile := fs.String("o", "", "输出文件路径")
	
	// 解析剩余参数
	if len(args) > 1 {
		fs.Parse(args[1:])
	}

	fmt.Printf("🚀 启动爬虫: %s\n", spiderName)

	// 加载配置
	config := loadSettings(*configFile, *settingsFlag)
	
	// 设置输出文件
	if *outputFile != "" {
		setOutputFile(config, *outputFile)
	}

	// 创建并运行爬虫
	if err := runSpider(spiderName, config); err != nil {
		fmt.Printf("❌ 爬虫运行失败: %v\n", err)
		os.Exit(1)
	}
}

// loadSettings 加载配置
func loadSettings(configFile, settingsFlag string) *settings.Settings {
	var config *settings.Settings

	// 如果指定了配置文件，尝试加载
	if configFile != "" {
		if data, err := os.ReadFile(configFile); err == nil {
			config = &settings.Settings{}
			if err := json.Unmarshal(data, config); err != nil {
				fmt.Printf("⚠️  配置文件解析失败，使用默认配置: %v\n", err)
				config = settings.DefaultSettings()
			}
		} else {
			fmt.Printf("⚠️  配置文件读取失败，使用默认配置: %v\n", err)
			config = settings.DefaultSettings()
		}
	} else {
		// 尝试加载默认配置文件
		defaultConfigPaths := []string{
			"scrapy.json",
			"settings.json",
			"config.json",
		}
		
		config = settings.DefaultSettings()
		for _, path := range defaultConfigPaths {
			if data, err := os.ReadFile(path); err == nil {
				tempConfig := &settings.Settings{}
				if err := json.Unmarshal(data, tempConfig); err == nil {
					config = tempConfig
					fmt.Printf("📄 加载配置文件: %s\n", path)
					break
				}
			}
		}
	}

	// 应用命令行设置
	if settingsFlag != "" {
		applyCommandLineSettings(config, settingsFlag)
	}

	return config
}

// applyCommandLineSettings 应用命令行设置
func applyCommandLineSettings(config *settings.Settings, settingsFlag string) {
	pairs := strings.Split(settingsFlag, ",")
	for _, pair := range pairs {
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			
			switch strings.ToUpper(key) {
			case "CONCURRENT", "CONCURRENT_REQUESTS":
				if val, err := strconv.Atoi(value); err == nil {
					config.ConcurrentRequests = val
					fmt.Printf("⚙️  设置并发数: %d\n", val)
				}
			case "DOWNLOAD_DELAY":
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					config.DownloadDelay = time.Duration(val * float64(time.Second))
					fmt.Printf("⚙️  设置下载延迟: %v\n", config.DownloadDelay)
				}
			case "USER_AGENT":
				config.UserAgent = value
				fmt.Printf("⚙️  设置User-Agent: %s\n", value)
			case "RANDOMIZE_DOWNLOAD_DELAY":
				if val, err := strconv.ParseBool(value); err == nil {
					config.RandomizeDownloadDelay = val
					fmt.Printf("⚙️  设置随机延迟: %v\n", val)
				}
			default:
				fmt.Printf("⚠️  未知设置: %s=%s\n", key, value)
			}
		}
	}
}

// setOutputFile 设置输出文件
func setOutputFile(config *settings.Settings, outputFile string) {
	// 确保输出目录存在
	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("⚠️  创建输出目录失败: %v\n", err)
		return
	}

	// 根据文件扩展名确定格式
	ext := strings.ToLower(filepath.Ext(outputFile))
	format := "json"
	switch ext {
	case ".json":
		format = "json"
	case ".csv":
		format = "csv"
	case ".xml":
		format = "xml"
	}

	// 设置输出配置
	config.FeedsExport = map[string]settings.FeedExportSettings{
		outputFile: {
			Format:   format,
			URI:      outputFile,
			Encoding: "utf-8",
		},
	}

	fmt.Printf("📁 输出文件: %s (格式: %s)\n", outputFile, format)
}

// runSpider 运行指定的爬虫
func runSpider(spiderName string, config *settings.Settings) error {
	// 创建引擎
	eng := engine.NewEngine()
	
	// 添加中间件
	userAgents := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
	eng.AddMiddleware(middleware.NewUserAgentMiddleware(userAgents, true))
	eng.AddMiddleware(middleware.NewDelayMiddleware(config.DownloadDelay, config.RandomizeDownloadDelay))
	
	// 添加管道
	if len(config.FeedsExport) > 0 {
		for _, feedConfig := range config.FeedsExport {
			jsonPipeline := pipeline.NewJSONPipeline(feedConfig.URI)
			eng.AddPipeline(jsonPipeline)
		}
	} else {
		// 默认输出到文件
		defaultOutput := fmt.Sprintf("%s_output.json", spiderName)
		jsonPipeline := pipeline.NewJSONPipeline(defaultOutput)
		eng.AddPipeline(jsonPipeline)
		fmt.Printf("📁 使用默认输出文件: %s\n", defaultOutput)
	}

	// 根据爬虫名称创建爬虫实例
	var spider spider.Spider
	switch strings.ToLower(spiderName) {
	case "douban", "douban_movie":
		spider = spiders.NewDoubanMovieSpider(config)
	default:
		return fmt.Errorf("未知的爬虫: %s", spiderName)
	}

	// 设置引擎配置
	eng.SetConcurrency(config.ConcurrentRequests)

	fmt.Printf("⚙️  并发数: %d\n", config.ConcurrentRequests)
	fmt.Printf("⏱️  下载延迟: %v\n", config.DownloadDelay)
	fmt.Printf("🎲 随机延迟: %v\n", config.RandomizeDownloadDelay)
	fmt.Println("🕷️  开始爬取...")

	// 记录开始时间
	startTime := time.Now()

	// 运行爬虫
	eng.Run(spider)

	// 显示统计信息
	duration := time.Since(startTime)
	fmt.Printf("\n✅ 爬取完成！总耗时: %v\n", duration)

	return nil
}