package spiders

import (
	"encoding/json"
	"fmt"
	"scrago/request"
	"scrago/response"
	"scrago/selector"
	"scrago/settings"
	"scrago/spider"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DoubanMovieItem 豆瓣电影数据结构
type DoubanMovieItem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Rating      float64  `json:"rating"`
	Year        string   `json:"year"`
	Directors   []string `json:"directors"`
	Actors      []string `json:"actors"`
	Genres      []string `json:"genres"`
	Cover       string   `json:"cover"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Summary     string   `json:"summary"`
	Duration    string   `json:"duration"`
	Country     []string `json:"country"`
	Language    []string `json:"language"`
	ReleaseDate string   `json:"release_date"`
	ScrapedAt   string   `json:"scraped_at"`
}

// NewDoubanMovieItem 创建新的电影项目
func NewDoubanMovieItem(movieID, url string) *DoubanMovieItem {
	return &DoubanMovieItem{
		ID:        movieID,
		URL:       url,
		ScrapedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// SetBasicInfo 设置基础信息
func (item *DoubanMovieItem) SetBasicInfo(basicInfo map[string]interface{}) {
	if basicInfo == nil {
		return
	}
	
	if title, ok := basicInfo["title"].(string); ok {
		item.Title = title
	}
	if cover, ok := basicInfo["cover"].(string); ok {
		item.Cover = cover
	}
	if ratingStr, ok := basicInfo["rating"].(string); ok && ratingStr != "" {
		if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil {
			item.Rating = rating
		}
	}
}

// ExtractFromHTML 从HTML中提取详细信息
func (item *DoubanMovieItem) ExtractFromHTML(htmlContent string) {
	sel := selector.NewSelector(htmlContent)
	
	// 提取标题（如果没有）
	if item.Title == "" {
		if title := sel.CSS("h1 span[property='v:itemreviewed']").Text(); title != "" {
			item.Title = strings.TrimSpace(title)
		} else if title := sel.CSS("h1 span").First().Text(); title != "" {
			item.Title = strings.TrimSpace(title)
		}
	}
	
	// 提取封面（如果没有）
	if item.Cover == "" {
		if cover := sel.CSS("#mainpic img").Attr("src"); cover != "" {
			item.Cover = cover
		}
	}
	
	// 提取评分（如果没有）
	if item.Rating == 0 {
		if ratingText := sel.CSS("strong.ll.rating_num").Text(); ratingText != "" {
			if rating, err := strconv.ParseFloat(strings.TrimSpace(ratingText), 64); err == nil {
				item.Rating = rating
			}
		}
	}
	
	// 提取年份
	if yearText := sel.CSS("h1 .year").Text(); yearText != "" {
		re := regexp.MustCompile(`\((\d{4})\)`)
		if matches := re.FindStringSubmatch(yearText); len(matches) > 1 {
			item.Year = matches[1]
		}
	}
	
	// 提取导演
	item.Directors = sel.CSS("a[rel='v:directedBy']").Texts()
	
	// 提取主演
	item.Actors = sel.CSS("a[rel='v:starring']").Texts()
	
	// 提取类型
	item.Genres = sel.CSS("span[property='v:genre']").Texts()
	
	// 提取片长
	if duration := sel.CSS("span[property='v:runtime']").Text(); duration != "" {
		item.Duration = strings.TrimSpace(duration)
	}
	
	// 提取制片国家/地区
	infoText := sel.CSS("#info").Text()
	if countryMatch := regexp.MustCompile(`制片国家/地区:\s*([^\n]+)`).FindStringSubmatch(infoText); len(countryMatch) > 1 {
		countries := strings.Split(countryMatch[1], "/")
		for i, country := range countries {
			countries[i] = strings.TrimSpace(country)
		}
		item.Country = countries
	}
	
	// 提取语言
	if langMatch := regexp.MustCompile(`语言:\s*([^\n]+)`).FindStringSubmatch(infoText); len(langMatch) > 1 {
		languages := strings.Split(langMatch[1], "/")
		for i, lang := range languages {
			languages[i] = strings.TrimSpace(lang)
		}
		item.Language = languages
	}
	
	// 提取上映日期
	if releaseDate := sel.CSS("span[property='v:initialReleaseDate']").Text(); releaseDate != "" {
		item.ReleaseDate = strings.TrimSpace(releaseDate)
	}
	
	// 提取简介
	if summary := sel.XPath("//*[@id='link-report-intra']/span").Text(); summary != "" {
		item.Summary = strings.TrimSpace(summary)
	} else if summary := sel.CSS("#link-report-intra span").Text(); summary != "" {
		item.Summary = strings.TrimSpace(summary)
	} else if summary := sel.CSS("#link-report .all.hidden").Text(); summary != "" {
		item.Summary = strings.TrimSpace(summary)
	} else if summary := sel.CSS("#link-report span[property='v:summary']").Text(); summary != "" {
		item.Summary = strings.TrimSpace(summary)
	} else if summary := sel.CSS("#link-report span").Text(); summary != "" {
		item.Summary = strings.TrimSpace(summary)
	}
}

// IsValid 检查电影项目是否有效
func (item *DoubanMovieItem) IsValid() bool {
	return item.ID != "" && item.Title != ""
}

// GetDisplayInfo 获取用于显示的简要信息
func (item *DoubanMovieItem) GetDisplayInfo() string {
	return item.Title + " (" + strconv.FormatFloat(item.Rating, 'f', 1, 64) + "分)"
}

// DoubanMovieSpider 豆瓣电影爬虫
type DoubanMovieSpider struct {
	*spider.BaseSpider
	settings *settings.Settings
}

// NewDoubanMovieSpider 创建豆瓣电影爬虫
func NewDoubanMovieSpider(settings *settings.Settings) *DoubanMovieSpider {
	// 起始URL - 豆瓣电影API
	startURLs := []string{
		"https://movie.douban.com/j/search_subjects?type=movie&tag=热门&sort=recommend&page_limit=20&page_start=0",
	}

	base := spider.NewBaseSpider("douban_movie_spider", startURLs)

	return &DoubanMovieSpider{
		BaseSpider: base,
		settings:   settings,
	}
}

// StartRequests 生成初始请求
func (s *DoubanMovieSpider) StartRequests() []*request.Request {
	baseURL := "https://movie.douban.com/j/search_subjects"
	
	var requests []*request.Request

	// 生成多页请求
	for start := 0; start < 60; start += 20 {
		url := fmt.Sprintf("%s?type=movie&tag=热门&sort=recommend&page_limit=20&page_start=%d", baseURL, start)
		req := request.NewRequest("GET", url)
		s.setAPIHeaders(req)
		req.SetMeta("callback", "parse")
		requests = append(requests, req)
	}

	fmt.Printf("🚀 豆瓣电影爬虫：生成了 %d 个初始请求\n", len(requests))
	return requests
}

// setAPIHeaders 设置API请求头
func (s *DoubanMovieSpider) setAPIHeaders(req *request.Request) {
	req.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.SetHeader("Accept", "application/json, text/plain, */*")
	req.SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.SetHeader("Accept-Encoding", "gzip, deflate, br")
	req.SetHeader("Referer", "https://movie.douban.com/")
	req.SetHeader("X-Requested-With", "XMLHttpRequest")
}

// setDetailHeaders 设置详情页请求头
func (s *DoubanMovieSpider) setDetailHeaders(req *request.Request) {
	req.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.SetHeader("Accept-Encoding", "gzip, deflate, br")
	req.SetHeader("Referer", "https://movie.douban.com/")
	req.SetHeader("Upgrade-Insecure-Requests", "1")
}

// Parse 解析豆瓣电影响应
func (s *DoubanMovieSpider) Parse(resp *response.Response) []interface{} {
	if resp.StatusCode != 200 {
		fmt.Printf("❌ 请求失败，状态码: %d, URL: %s\n", resp.StatusCode, resp.URL)
		return []interface{}{}
	}

	// 检查是否是详情页请求
	if callback, exists := resp.Meta["callback"]; exists && callback == "parseMovieDetail" {
		return s.ParseMovieDetail(resp)
	}

	// 检查URL是否是详情页
	if strings.Contains(resp.URL, "/subject/") {
		return s.ParseMovieDetail(resp)
	}

	// 解析API响应
	var apiResponse struct {
		Subjects []struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Rate     string `json:"rate"`
			Cover    string `json:"cover"`
			URL      string `json:"url"`
			Playable bool   `json:"playable"`
			IsNew    bool   `json:"is_new"`
		} `json:"subjects"`
	}

	if err := json.Unmarshal(resp.Body, &apiResponse); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		fmt.Printf("🔍 响应URL: %s\n", resp.URL)
		if len(resp.Body) > 100 {
			fmt.Printf("🔍 响应前100字符: %s\n", string(resp.Body[:100]))
		} else {
			fmt.Printf("🔍 响应内容: %s\n", string(resp.Body))
		}
		return []interface{}{}
	}

	if len(apiResponse.Subjects) == 0 {
		fmt.Printf("⚠️  空响应，可能遇到反爬虫，URL: %s\n", resp.URL)
		return []interface{}{}
	}

	var results []interface{}
	fmt.Printf("📄 从API获取到 %d 部电影，生成详情页请求...\n", len(apiResponse.Subjects))

	// 为每部电影生成详情页请求
	for _, subject := range apiResponse.Subjects {
		// 创建详情页请求
		detailReq := request.NewRequest("GET", subject.URL)
		s.setDetailHeaders(detailReq)
		detailReq.SetMeta("callback", "parseMovieDetail")
		detailReq.SetMeta("movie_id", subject.ID)
		detailReq.SetMeta("basic_info", map[string]interface{}{
			"id":       subject.ID,
			"title":    subject.Title,
			"rating":   subject.Rate,
			"cover":    subject.Cover,
			"url":      subject.URL,
			"playable": subject.Playable,
			"new":      subject.IsNew,
		})

		results = append(results, detailReq)
	}

	fmt.Printf("✅ 生成了 %d 个详情页请求\n", len(results))
	return results
}

// ParseMovieDetail 解析电影详情页
func (s *DoubanMovieSpider) ParseMovieDetail(resp *response.Response) []interface{} {
	if resp.StatusCode != 200 {
		fmt.Printf("❌ 详情页请求失败，状态码: %d, URL: %s\n", resp.StatusCode, resp.URL)
		return []interface{}{}
	}

	// 从URL提取电影ID
	movieID := ""
	if matches := regexp.MustCompile(`subject/(\d+)`).FindStringSubmatch(resp.URL); len(matches) > 1 {
		movieID = matches[1]
	}

	// 创建电影项目
	movie := NewDoubanMovieItem(movieID, resp.URL)

	// 获取基础信息（如果存在）
	if info, exists := resp.Meta["basic_info"]; exists && info != nil {
		basicInfo := info.(map[string]interface{})
		movie.SetBasicInfo(basicInfo)
	}

	// 从HTML中提取详细信息
	movie.ExtractFromHTML(string(resp.Body))

	// 验证数据有效性
	if !movie.IsValid() {
		fmt.Printf("⚠️  电影数据无效，跳过: %s\n", resp.URL)
		return []interface{}{}
	}

	fmt.Printf("🎬 解析电影: %s\n", movie.GetDisplayInfo())

	return []interface{}{movie}
}