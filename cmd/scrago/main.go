package main

import (
	"fmt"
	"scrago/cmd/scrago/commands"
	"os"
	"strings"
)

const (
	Version = "1.0.0"
	Banner  = `


 _______ _______ _______ _______ _______ _______ 
(  ____ (  ____ (  ____ (  ___  (  ____ (  ___  )
| (    \| (    \| (    )| (   ) | (    \| (   ) |
| (_____| |     | (____)| (___) | |     | |   | |
(_____  | |     |     __|  ___  | | ____| |   | |
      ) | |     | (\ (  | (   ) | | \_  | |   | |
/\____) | (____/| ) \ \_| )   ( | (___) | (___) |
\_______(_______|/   \__|/     \(_______(_______)
                                                 
                                                                         
Scrago v%s - A high-performance web scraping framework for Go
`
)

func main() {
	fmt.Printf(Banner, Version)

	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "crawl":
		commands.CrawlCommand(args)
	case "list":
		commands.ListCommand(args)
	case "startproject":
		commands.StartProjectCommand(args)
	case "genspider":
		commands.GenSpiderCommand(args)
	case "version":
		fmt.Printf("Scrago %s\n", Version)
	case "help", "-h", "--help":
		showHelp()
	default:
		fmt.Printf("❌ 未知命令: %s\n\n", command)
		showHelp()
	}
}

func showHelp() {
	help := `
使用方法:
  scrago <command> [options] [args]

可用命令:
  crawl <spider>     运行指定的爬虫
  list              列出所有可用的爬虫
  startproject      创建新的爬虫项目
  genspider         生成新的爬虫模板
  version           显示版本信息
  help              显示帮助信息

示例:
  scrago crawl douban                    # 运行豆瓣爬虫
  scrago crawl douban -s CONCURRENT=32   # 运行豆瓣爬虫并设置并发数
  scrago list                           # 列出所有爬虫
  scrago startproject myproject         # 创建新项目
  scrago genspider example example.com  # 生成新爬虫

更多信息请访问: https://github.com/bgspiders/scrago
`
	fmt.Println(strings.TrimSpace(help))
}