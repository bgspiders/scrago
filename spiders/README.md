# 豆瓣电影爬虫示例

这个目录包含了一个完整的豆瓣电影爬虫示例，展示了如何使用 scrago 框架构建一个功能完整的爬虫。

## 文件说明

### `douban_movie_spider.go`
完整的豆瓣电影爬虫实现，包含：

- **DoubanMovieItem**: 电影数据结构，包含所有电影信息字段
- **DoubanMovieSpider**: 爬虫主体，实现了豆瓣电影的抓取逻辑
- **数据提取功能**: 
  - 从豆瓣API获取电影列表
  - 抓取电影详情页面
  - 使用CSS选择器和XPath提取详细信息

## 功能特性

### 🎬 数据抓取
- 电影基本信息（标题、评分、年份等）
- 详细信息（导演、演员、类型、简介等）
- 技术信息（片长、国家、语言、上映日期等）

### 🚀 性能优化
- 并发请求处理
- 智能延迟控制
- 用户代理轮换
- 错误重试机制

### 📊 数据输出
- JSON格式输出
- 控制台实时显示
- 完整的性能统计

## 运行示例

### 方法1：使用测试程序
```bash
# 运行豆瓣电影爬虫测试
go run cmd/douban_test/main.go
```

### 方法2：在代码中使用
```go
package main

import (
    "scrago/spiders"
    "scrago/settings"
    "scrago/engine"
)

func main() {
    // 创建设置
    settings := settings.DoubanSettings()
    
    // 创建爬虫
    spider := spiders.NewDoubanMovieSpider(settings)
    
    // 创建引擎并运行
    engine := engine.NewEngine()
    engine.Run(spider)
}
```

## 配置说明

### 并发设置
- `ConcurrentRequests`: 并发请求数（默认16）
- `DownloadDelay`: 下载延迟（默认100ms）

### 输出设置
- 数据保存到 `douban_movies_test.json`
- 控制台实时显示抓取进度

### 中间件
- **UserAgentMiddleware**: 用户代理轮换
- **DelayMiddleware**: 请求延迟控制

### 管道
- **ConsolePipeline**: 控制台输出
- **JSONPipeline**: JSON文件输出

## 数据结构示例

```json
{
  "data": {
    "id": "35929770",
    "title": "魔法蓝精灵 Smurfs",
    "rating": 6.1,
    "year": "2025",
    "directors": ["克里斯·米勒"],
    "actors": ["蕾哈娜", "詹姆斯·柯登"],
    "genres": ["喜剧", "动画", "奇幻", "冒险"],
    "cover": "https://img9.doubanio.com/view/photo/s_ratio_poster/public/p2923060466.webp",
    "url": "https://movie.douban.com/subject/35929770/",
    "summary": "蓝精灵村庄每日载歌载舞一片欢乐...",
    "duration": "90分钟(中国大陆)",
    "country": ["美国"],
    "language": ["英语"],
    "release_date": "2025-07-18(美国/中国大陆)",
    "scraped_at": "2025-08-08 13:30:29"
  },
  "type": "*spiders.DoubanMovieItem"
}
```

## 性能表现

测试结果显示：
- ⏱️ 总耗时: ~4.3秒
- 🚀 并发数: 16
- 📊 请求总数: 63个
- ✅ 成功率: 100%
- 🎬 抓取电影: 60部
- ⚡ 请求速率: ~14.6 请求/秒

## 扩展建议

1. **增加更多数据字段**: 可以提取更多电影信息
2. **支持不同分类**: 修改API参数抓取不同类型电影
3. **添加数据清洗**: 对提取的数据进行进一步处理
4. **支持增量更新**: 避免重复抓取已有数据
5. **添加代理支持**: 提高抓取稳定性

## 注意事项

- 请遵守豆瓣的robots.txt规则
- 合理设置请求频率，避免对服务器造成压力
- 仅用于学习和研究目的
- 商业使用请确保符合相关法律法规