package commands

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SpiderInfo 爬虫信息
type SpiderInfo struct {
	Name        string
	Description string
	StartURLs   []string
}

// ListCommand 处理 list 命令
func ListCommand(args []string) {
	fmt.Println("📋 可用的爬虫列表:")
	fmt.Println(strings.Repeat("=", 50))

	spiders := getAvailableSpiders()
	
	if len(spiders) == 0 {
		fmt.Println("❌ 没有找到可用的爬虫")
		fmt.Println("💡 提示: 使用 'scrago genspider <name> <domain>' 创建新爬虫")
		return
	}

	for i, spider := range spiders {
		fmt.Printf("%d. %s\n", i+1, spider.Name)
		if spider.Description != "" {
			fmt.Printf("   📝 %s\n", spider.Description)
		}
		if len(spider.StartURLs) > 0 {
			fmt.Printf("   🌐 起始URL: %s\n", strings.Join(spider.StartURLs, ", "))
		}
		fmt.Println()
	}

	fmt.Printf("总共找到 %d 个爬虫\n", len(spiders))
	fmt.Println("\n💡 使用方法: scrago crawl <spider_name>")
}

// getAvailableSpiders 动态获取可用的爬虫列表
func getAvailableSpiders() []SpiderInfo {
	spiders := make([]SpiderInfo, 0)
	
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("⚠️  获取当前目录失败: %v\n", err)
		return spiders
	}
	
	// 构建spiders目录路径
	spidersDir := filepath.Join(currentDir, "spiders")
	
	// 检查spiders目录是否存在
	if _, err := os.Stat(spidersDir); os.IsNotExist(err) {
		fmt.Printf("⚠️  spiders目录不存在: %s\n", spidersDir)
		return spiders
	}
	
	// 扫描spiders目录
	err = filepath.WalkDir(spidersDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// 只处理.go文件，排除README等文件
		if !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			spiderInfo := parseSpiderFile(path)
			if spiderInfo != nil {
				spiders = append(spiders, *spiderInfo)
			}
		}
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("⚠️  扫描spiders目录失败: %v\n", err)
	}
	
	return spiders
}

// parseSpiderFile 解析爬虫文件，提取爬虫信息
func parseSpiderFile(filePath string) *SpiderInfo {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	
	fileContent := string(content)
	
	// 解析Go源码
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, fileContent, parser.ParseComments)
	if err != nil {
		return nil
	}
	
	var spiderInfo SpiderInfo
	var spiderNames []string
	var startURLs []string
	var description string
	
	// 遍历AST节点
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			// 查找Spider结构体
			if strings.HasSuffix(x.Name.Name, "Spider") {
				spiderName := convertSpiderName(x.Name.Name)
				spiderNames = append(spiderNames, spiderName)
			}
		case *ast.FuncDecl:
			// 查找New*Spider函数
			if x.Name != nil && strings.HasPrefix(x.Name.Name, "New") && strings.HasSuffix(x.Name.Name, "Spider") {
				spiderName := convertSpiderName(strings.TrimPrefix(strings.TrimSuffix(x.Name.Name, "Spider"), "New"))
				if spiderName != "" {
					spiderNames = append(spiderNames, spiderName)
				}
			}
		case *ast.GenDecl:
			// 查找注释中的描述信息
			if x.Doc != nil {
				for _, comment := range x.Doc.List {
					text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
					if strings.Contains(text, "爬虫") || strings.Contains(text, "Spider") {
						description = text
						break
					}
				}
			}
		}
		return true
	})
	
	// 使用正则表达式提取起始URL
	startURLs = extractStartURLs(fileContent)
	
	// 如果没有找到描述，尝试从文件名生成
	if description == "" {
		fileName := filepath.Base(filePath)
		fileName = strings.TrimSuffix(fileName, ".go")
		description = generateDescription(fileName)
	}
	
	// 如果找到了爬虫信息
	if len(spiderNames) > 0 {
		// 去重并选择最合适的名称
		uniqueNames := removeDuplicates(spiderNames)
		primaryName := selectPrimaryName(uniqueNames)
		
		spiderInfo = SpiderInfo{
			Name:        primaryName,
			Description: description,
			StartURLs:   startURLs,
		}
		
		return &spiderInfo
	}
	
	return nil
}

// convertSpiderName 将Spider类名转换为爬虫名称
func convertSpiderName(name string) string {
	// 移除Spider后缀
	name = strings.TrimSuffix(name, "Spider")
	
	// 转换驼峰命名为下划线命名
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	name = re.ReplaceAllString(name, "${1}_${2}")
	
	return strings.ToLower(name)
}

// extractStartURLs 从文件内容中提取起始URL
func extractStartURLs(content string) []string {
	var urls []string
	
	// 匹配字符串中的URL
	urlPattern := regexp.MustCompile(`"(https?://[^"]+)"`)
	matches := urlPattern.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			url := match[1]
			// 过滤掉明显不是起始URL的链接
			if !strings.Contains(url, "example.com") && 
			   !strings.Contains(url, "localhost") &&
			   !strings.Contains(url, "127.0.0.1") {
				urls = append(urls, url)
			}
		}
	}
	
	// 去重
	return removeDuplicates(urls)
}

// generateDescription 根据文件名生成描述
func generateDescription(fileName string) string {
	// 移除常见后缀
	fileName = strings.TrimSuffix(fileName, "_spider")
	fileName = strings.TrimSuffix(fileName, "spider")
	
	// 替换下划线为空格并首字母大写
	parts := strings.Split(fileName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	
	return strings.Join(parts, " ") + " 爬虫"
}

// removeDuplicates 去除重复项
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] && item != "" {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// selectPrimaryName 选择主要的爬虫名称
func selectPrimaryName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	
	// 优先选择较短的名称
	primary := names[0]
	for _, name := range names {
		if len(name) < len(primary) {
			primary = name
		}
	}
	
	return primary
}