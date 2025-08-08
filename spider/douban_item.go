package spider

import (
	"scrago/selector"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DetailedMovieItem 详细电影数据结构
type DetailedMovieItem struct {
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

// NewDetailedMovieItem 创建新的电影项目
func NewDetailedMovieItem(movieID, url string) *DetailedMovieItem {
	return &DetailedMovieItem{
		ID:        movieID,
		URL:       url,
		ScrapedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// SetBasicInfo 设置基础信息（从API获取的数据）
func (item *DetailedMovieItem) SetBasicInfo(basicInfo map[string]interface{}) {
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
func (item *DetailedMovieItem) ExtractFromHTML(htmlContent string) {
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
	
	// 提取简介 - 使用XPath和CSS选择器多种方式
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
func (item *DetailedMovieItem) IsValid() bool {
	return item.ID != "" && item.Title != ""
}

// GetDisplayInfo 获取用于显示的简要信息
func (item *DetailedMovieItem) GetDisplayInfo() string {
	return item.Title + " (" + strconv.FormatFloat(item.Rating, 'f', 1, 64) + "分)"
}